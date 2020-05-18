// Copyright 2018-2020 Celer Network

package testing

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	msgrpc "github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/webapi/rpc"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type ClientController struct {
	keystorePath, dbPath string
	process              *os.Process
	apiClient            rpc.WebApiClient
	internalApiClient    rpc.InternalWebApiClient
}

func StartClientWithoutProxy(
	keystorePath string, profileFileName string, clientID string, extraArgs ...string) (*ClientController, error) {
	return StartClientController(
		GetNextClientPort(), keystorePath, outRootDir+profileFileName, outRootDir+clientID+"Store", clientID, extraArgs...)
}

func StartC1WithoutProxy(
	keystorePath string, extraArgs ...string) (*ClientController, error) {
	extraArgs = append(extraArgs, "-extsign") // always use external signer for c1
	return StartClientController(
		GetNextClientPort(), keystorePath, outRootDir+"profile.json", outRootDir+"c1Store", "c1", extraArgs...)
}

func StartC2WithoutProxy(
	keystorePath string, extraArgs ...string) (*ClientController, error) {
	return StartClientController(
		GetNextClientPort(), keystorePath, outRootDir+"profile.json", outRootDir+"c2Store", "c2", extraArgs...)
}

func StartClientController(
	listenPort string,
	keystorePath string,
	configPath string,
	dataPath string,
	logPrefix string,
	extraArgs ...string) (*ClientController, error) {
	// assume keystorepath like /tmp/celer_e2e_1568841320/ksdir/UTC--2019-09-18T21-16-18.937701241Z--df99f4583541aa366d93d319f7b1ea4c47f75c3e
	slist := strings.Split(keystorePath, "--")
	eth := slist[len(slist)-1]
	logPrefix += "_" + eth[:4]
	args := append(
		[]string{
			"-port", listenPort,
			"-keystore", keystorePath,
			"-config", configPath,
			"-datadir", dataPath,
			"-logprefix", logPrefix,
			"-logcolor",
			"-dropmsg",
		},
		extraArgs...)
	os.MkdirAll(dataPath, os.ModePerm)
	process := StartProcess(outRootDir+"testclient", args...)
	time.Sleep(1 * time.Second)
	conn, err := grpc.Dial(
		"localhost:"+listenPort,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(4*time.Second))
	if err != nil {
		process.Kill()
		return nil, err
	}
	go func() {
		ps, err2 := process.Wait()
		if err2 != nil {
			log.Error(err2)
		}
		log.Infof("client %s stopped. reason: %s. eth: %s", logPrefix, ps.String(), eth)
	}()
	log.Infoln("client controller started.", "localhost:"+listenPort, eth)
	return &ClientController{
		keystorePath:      keystorePath,
		dbPath:            path.Join(dataPath, eth),
		process:           process,
		apiClient:         rpc.NewWebApiClient(conn),
		internalApiClient: rpc.NewInternalWebApiClient(conn),
	}, nil
}

// KillAndRemoveDB kill process and delete the dataPath/xxxx folder
// keystore is kept so we can test same eth start from scratch case
// Note this depends on cnode/kvstore init logic that creates eth address folder
func (cc *ClientController) KillAndRemoveDB() {
	defer os.RemoveAll(cc.dbPath)
	KillProcess(cc.process)
}

// RunSQL calls sqlite3 query directly on db file
// q needs to quote args properly
func (cc *ClientController) RunSQL(q string) ([]byte, error) {
	return exec.Command("sqlite3", path.Join(cc.dbPath, "sqlite/celer.db"), q).Output()
}

func (cc *ClientController) SendPayment(
	destination string,
	amountWei string,
	tokenType entity.TokenType,
	tokenAddress string) (string, error) {
	return cc.SendPaymentWithBooleanConditions(
		destination, amountWei, tokenType, tokenAddress, []*entity.Condition{}, 100)
}

func (cc *ClientController) SendPaymentWithBooleanConditions(
	destination string,
	amountWei string,
	tokenType entity.TokenType,
	tokenAddress string,
	conditions []*entity.Condition,
	timeout uint64) (string, error) {
	rpcConditions := make([]*rpc.Condition, len(conditions))
	for i, condition := range conditions {
		var onChainDeployed bool
		var contractAddress []byte
		if condition.ConditionType == entity.ConditionType_DEPLOYED_CONTRACT {
			onChainDeployed = true
			contractAddress = condition.DeployedContractAddress
		} else {
			contractAddress = condition.VirtualContractAddress
		}
		rpcConditions[i] = &rpc.Condition{
			OnChainDeployed: onChainDeployed,
			ContractAddress: ctype.Bytes2Hex(contractAddress),
			IsFinalizedArgs: condition.ArgsQueryFinalization,
			GetOutcomeArgs:  condition.ArgsQueryOutcome,
		}
	}
	resp, err := cc.apiClient.SendConditionalPayment(
		context.Background(),
		&rpc.SendConditionalPaymentRequest{
			TokenInfo: &rpc.TokenInfo{
				TokenType:    tokenType,
				TokenAddress: tokenAddress,
			},
			Destination: destination,
			Amount:      amountWei,
			Conditions:  rpcConditions,
			Timeout:     timeout,
		})
	if err != nil {
		return "", err
	}
	return resp.PaymentId, nil
}

func (cc *ClientController) SubscribeIncomingPayments(paymentChan chan string) error {
	sub, err := cc.apiClient.SubscribeIncomingPayments(context.Background(), &empty.Empty{})
	if err != nil {
		return err
	}
	go func() {
		for {
			payment, err := sub.Recv()
			if err != nil {
				return
			}
			paymentChan <- payment.PaymentId
		}
	}()
	return nil
}

func (cc *ClientController) SubscribeOutgoingPayments(
	completeChan chan string, errChan chan string) error {
	sub, err := cc.apiClient.SubscribeOutgoingPayments(context.Background(), &empty.Empty{})
	if err != nil {
		return err
	}
	go func() {
		for {
			paymentInfo, err := sub.Recv()
			if err != nil {
				return
			}
			if paymentInfo.ErrorReason != "" {
				errChan <- paymentInfo.ErrorReason
			} else {
				completeChan <- paymentInfo.Payment.PaymentId
			}
		}
	}()
	return nil
}

func (cc *ClientController) GetIncomingPaymentStatus(paymentID string) (int, error) {
	status, err := cc.apiClient.GetIncomingPaymentStatus(
		context.Background(), &rpc.PaymentID{
			PaymentId: paymentID,
		})
	if err != nil {
		return 0, err
	}
	return int(status.Status), nil
}

func (cc *ClientController) GetOutgoingPaymentStatus(paymentID string) (int, error) {
	status, err := cc.apiClient.GetOutgoingPaymentStatus(
		context.Background(), &rpc.PaymentID{
			PaymentId: paymentID,
		})
	if err != nil {
		return 0, err
	}
	return int(status.Status), nil
}

func (cc *ClientController) ConfirmBooleanPay(paymentID string) error {
	_, err := cc.apiClient.ConfirmOutgoingPayment(
		context.Background(),
		&rpc.PaymentID{
			PaymentId: paymentID,
		})
	return err
}

func (cc *ClientController) RejectBooleanPay(paymentID string) error {
	_, err := cc.apiClient.RejectIncomingPayment(
		context.Background(),
		&rpc.PaymentID{
			PaymentId: paymentID,
		})
	return err
}

func (cc *ClientController) SettleOnChainResolvedPay(paymentID string) error {
	_, err := cc.apiClient.SettleOnChainResolvedIncomingPayment(
		context.Background(),
		&rpc.PaymentID{
			PaymentId: paymentID,
		})
	return err
}

func (cc *ClientController) ConfirmOnChainResolvedPays(
	tokenType entity.TokenType, tokenAddr string) error {
	_, err := cc.apiClient.ConfirmOnChainResolvedPayments(
		context.Background(),
		&rpc.TokenInfo{
			TokenType:    tokenType,
			TokenAddress: tokenAddr,
		})
	return err
}

func (cc *ClientController) SettleExpiredPays(tokenType entity.TokenType, tokenAddr string) error {
	_, err := cc.apiClient.SettleExpiredPayments(
		context.Background(),
		&rpc.TokenInfo{
			TokenType:    tokenType,
			TokenAddress: tokenAddr,
		})
	return err
}

func (cc *ClientController) Deposit(
	tokenType entity.TokenType, tokenAddr, amt string) (*rpc.DepositOrWithdrawJob, error) {
	return cc.apiClient.Deposit(
		context.Background(),
		&rpc.DepositOrWithdrawRequest{
			TokenInfo: &rpc.TokenInfo{TokenType: tokenType, TokenAddress: tokenAddr},
			Amount:    amt,
		})
}

func (cc *ClientController) DepositNonBlocking(
	tokenType entity.TokenType, tokenAddr, amt string) (*rpc.DepositOrWithdrawJob, error) {
	return cc.internalApiClient.DepositNonBlocking(
		context.Background(),
		&rpc.DepositOrWithdrawRequest{
			TokenInfo: &rpc.TokenInfo{TokenType: tokenType, TokenAddress: tokenAddr},
			Amount:    amt,
		})
}

func (cc *ClientController) MonitorDepositJob(jobID string) (*rpc.DepositOrWithdrawJob, error) {
	return cc.apiClient.MonitorDepositJob(
		context.Background(),
		&rpc.DepositOrWithdrawJob{JobId: jobID})
}

func (cc *ClientController) IntendSettlePaymentChannel(
	tokenType entity.TokenType, tokenAddr string) error {
	_, err := cc.apiClient.IntendSettlePaymentChannel(
		context.Background(),
		&rpc.TokenInfo{
			TokenType:    tokenType,
			TokenAddress: tokenAddr,
		})
	return err
}

func (cc *ClientController) ConfirmSettlePaymentChannel(
	tokenType entity.TokenType, tokenAddr string) error {
	_, err := cc.apiClient.ConfirmSettlePaymentChannel(
		context.Background(),
		&rpc.TokenInfo{
			TokenType:    tokenType,
			TokenAddress: tokenAddr,
		})
	return err
}

// GetPayHistory returns paginated historical pays. The returned boolean is true if there is more result to fetch.
func (cc *ClientController) GetPayHistory(fromStart bool, itemsPerPage int32) ([]*msgrpc.OneHistoricalPay, bool, error) {
	resp, err := cc.apiClient.GetPayHistory(
		context.Background(),
		&rpc.GetPayHistoryRequest{
			FromStart:    fromStart,
			ItemsPerPage: itemsPerPage,
		})
	if err != nil {
		return nil, false, err
	}
	return resp.GetPays(), resp.GetHasMoreResult(), nil
}

func (cc *ClientController) GetSettleFinalizedTime(
	tokenType entity.TokenType, tokenAddr string) (uint64, error) {
	blknum, err := cc.apiClient.GetSettleFinalizedTimeForPaymentChannel(
		context.Background(),
		&rpc.TokenInfo{
			TokenType:    tokenType,
			TokenAddress: tokenAddr,
		})
	if err != nil {
		return 0, err
	}
	return blknum.BlockNumber, nil
}

func (cc *ClientController) OpenChannel(
	selfAddress string,
	tokenType entity.TokenType,
	tokenAddr string,
	selfAmt string,
	peerAmt string) (*rpc.ChannelID, error) {
	tokenInfo := &rpc.TokenInfo{
		TokenType:    tokenType,
		TokenAddress: tokenAddr,
	}
	resp, err := cc.apiClient.OpenPaymentChannel(
		context.Background(),
		&rpc.OpenPaymentChannelRequest{
			TokenInfo:  tokenInfo,
			Amount:     selfAmt,
			PeerAmount: peerAmt,
		})
	if err != nil {
		return nil, fmt.Errorf("OpenPaymentChannel err: %w", err)
	}
	const retryLimit = 5
	for retry := 0; retry < retryLimit; retry++ {
		status, err2 := cc.apiClient.GetPeerFreeBalance(
			context.Background(),
			&rpc.GetPeerFreeBalanceRequest{
				TokenInfo:   tokenInfo,
				PeerAddress: selfAddress,
			})
		if err2 != nil {
			return nil, fmt.Errorf("GetPeerFreeBalance err: %w", err2)
		}
		if status.GetJoinStatus() == int32(msgrpc.JoinCelerStatus_LOCAL) {
			return resp, nil
		}
		time.Sleep(time.Second)
	}
	return resp, fmt.Errorf("channel not opened, cid %s", resp.GetChannelId())
}

func (cc *ClientController) TcbOpenChannel(
	selfAddress string,
	tokenType entity.TokenType,
	tokenAddr string,
	peerAmt string) (*rpc.ChannelID, error) {
	resp, err := cc.internalApiClient.OpenTrustedPaymentChannel(
		context.Background(),
		&rpc.OpenPaymentChannelRequest{
			TokenInfo: &rpc.TokenInfo{
				TokenType:    tokenType,
				TokenAddress: tokenAddr,
			},
			PeerAmount: peerAmt,
		})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (cc *ClientController) InstantiateChannel(
	tokenType entity.TokenType,
	tokenAddr string) (*rpc.ChannelID, error) {
	resp, err := cc.internalApiClient.InstantiateTrustedPaymentChannel(
		context.Background(),
		&rpc.TokenInfo{
			TokenType:    tokenType,
			TokenAddress: tokenAddr,
		})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (cc *ClientController) GetBalance(tokenAddress string) (string, string, string, error) {
	tokenAddr := ctype.Hex2Addr(tokenAddress)
	var tokenType entity.TokenType
	if tokenAddr == ctype.Hex2Addr(ctype.EthTokenAddrStr) {
		tokenType = entity.TokenType_ETH
	} else {
		tokenType = entity.TokenType_ERC20
	}
	resp, err :=
		cc.apiClient.GetBalance(
			context.Background(),
			&rpc.TokenInfo{TokenType: tokenType, TokenAddress: tokenAddress})
	if err != nil {
		log.Error(err)
		return "", "", "", err
	}
	return resp.FreeBalance, resp.LockedBalance, resp.ReceivingCapacity, err
}

func (cc *ClientController) GetAccountBalance(
	tokenAddr string, ownerAddr string, conn *ethclient.Client) (*big.Int, error) {
	if tokenAddr != ctype.EthTokenAddrStr {
		// ERC20 token
		erc20Contract, err := chain.NewERC20(ctype.Hex2Addr(tokenAddr), conn)
		if err != nil {
			return big.NewInt(0), err
		}
		return erc20Contract.BalanceOf(&bind.CallOpts{}, ctype.Hex2Addr(ownerAddr))
	}
	return conn.BalanceAt(context.Background(), ctype.Hex2Addr(ownerAddr), nil)
}

func (cc *ClientController) Kill() {
	KillProcess(cc.process)
}

func (cc *ClientController) SetDelegation(tokens []string, duration int64) error {
	tokenInfos := make([]*rpc.TokenInfo, 0, len(tokens))
	for _, tk := range tokens {
		token := &rpc.TokenInfo{
			TokenType:    entity.TokenType_ERC20,
			TokenAddress: tk,
		}
		if tk == ctype.EthTokenAddrStr {
			token.TokenType = entity.TokenType_ETH
		}
		tokenInfos = append(tokenInfos, token)
	}
	_, err := cc.apiClient.SetDelegation(context.Background(), &rpc.SetDelegationRequest{
		TokenInfos:    tokenInfos,
		BlockDuration: duration,
	})
	return err
}
func (cc *ClientController) KillWithoutRemovingKeystore() {
	KillProcess(cc.process)
}

func (cc *ClientController) AssertBalance(
	tokenAddr string, expectedFree string, expectedLocked string, expectedReceiveCap string) error {
	const retryLimit = 5
	var err error
	for retry := 0; retry < retryLimit; retry++ {
		free, locked, receiveCap, getBalanceErr := cc.GetBalance(tokenAddr)
		if getBalanceErr != nil {
			err = getBalanceErr
		} else if free != expectedFree || locked != expectedLocked ||
			receiveCap != expectedReceiveCap {
			err = fmt.Errorf(
				"Wrong balance, expected %s %s %s, got %s %s %s",
				expectedFree, expectedLocked, expectedReceiveCap,
				free, locked, receiveCap)
		} else {
			return nil
		}
		log.Warnf("AssertBalance retry: %d error: %s", retry, err)
		time.Sleep(time.Second)
	}
	return err
}

func (cc *ClientController) IsConnectedToCeler(
	tokenAddress string, address string) (string, error) {
	tokenAddr := ctype.Hex2Addr(tokenAddress)
	var tokenType entity.TokenType
	if tokenAddr == ctype.Hex2Addr(ctype.EthTokenAddrStr) {
		tokenType = entity.TokenType_ETH
	} else {
		tokenType = entity.TokenType_ERC20
	}
	resp, err := cc.apiClient.GetPeerFreeBalance(
		context.Background(),
		&rpc.GetPeerFreeBalanceRequest{
			PeerAddress: address,
			TokenInfo: &rpc.TokenInfo{
				TokenType:    tokenType,
				TokenAddress: tokenAddress,
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.FreeBalance, err
}

func (cc *ClientController) GetCurrentBlockNumber() (uint64, error) {
	blkNum, err := cc.apiClient.GetBlockNumber(
		context.Background(),
		&empty.Empty{},
	)
	if err != nil {
		return 0, err
	}
	return blkNum.BlockNumber, nil
}

func (cc *ClientController) CooperativeWithdraw(
	tokenType entity.TokenType,
	tokenAddr string,
	withdrawAmount string) (*rpc.DepositOrWithdrawJob, error) {
	return cc.apiClient.CooperativeWithdraw(
		context.Background(),
		&rpc.DepositOrWithdrawRequest{
			TokenInfo: &rpc.TokenInfo{
				TokenType:    tokenType,
				TokenAddress: tokenAddr,
			},
			Amount: withdrawAmount,
		})
}

func (cc *ClientController) CooperativeWithdrawNonBlocking(
	tokenType entity.TokenType,
	tokenAddr string,
	withdrawAmount string) (*rpc.DepositOrWithdrawJob, error) {
	return cc.internalApiClient.CooperativeWithdrawNonBlocking(
		context.Background(),
		&rpc.DepositOrWithdrawRequest{
			TokenInfo: &rpc.TokenInfo{
				TokenType:    tokenType,
				TokenAddress: tokenAddr,
			},
			Amount: withdrawAmount,
		})
}

func (cc *ClientController) MonitorCooperativeWithdrawJob(jobID string) (*rpc.DepositOrWithdrawJob, error) {
	return cc.apiClient.MonitorCooperativeWithdrawJob(
		context.Background(),
		&rpc.DepositOrWithdrawJob{JobId: jobID})
}

func (cc *ClientController) IntendWithdraw(
	tokenType entity.TokenType,
	tokenAddr string,
	withdrawAmount string) error {
	_, err := cc.apiClient.IntendWithdraw(
		context.Background(),
		&rpc.DepositOrWithdrawRequest{
			TokenInfo: &rpc.TokenInfo{
				TokenType:    tokenType,
				TokenAddress: tokenAddr,
			},
			Amount: withdrawAmount,
		})
	return err
}

func (cc *ClientController) ConfirmWithdraw(
	tokenType entity.TokenType,
	tokenAddr string) error {
	_, err := cc.apiClient.ConfirmWithdraw(
		context.Background(),
		&rpc.TokenInfo{
			TokenType:    tokenType,
			TokenAddress: tokenAddr,
		})
	return err
}

func (cc *ClientController) SettleConditionalPayOnChain(paymentID string) (string, uint64, error) {
	rpcPaymentID := &rpc.PaymentID{PaymentId: paymentID}
	_, err :=
		cc.apiClient.ResolveIncomingPaymentOnChain(
			context.Background(),
			rpcPaymentID)
	if err != nil {
		return "", 0, err
	}
	info, err := cc.apiClient.GetOnChainPaymentInfo(context.Background(), rpcPaymentID)
	if err != nil {
		return "", 0, err
	}
	return info.Amount, info.ResolveDeadline, nil
}

func (cc *ClientController) SignOutgoingState(sessionID string, state []byte) ([]byte, error) {
	signedState, err := cc.apiClient.SignOutgoingState(
		context.Background(),
		&rpc.SignOutgoingStateRequest{SessionId: sessionID, State: state})
	return signedState.SignedState, err
}

func (cc *ClientController) SignData(data []byte) ([]byte, error) {
	signature, err := cc.apiClient.SignData(context.Background(), &rpc.Data{Data: data})
	if err != nil {
		return nil, err
	}
	return signature.Signature, nil
}

func (cc *ClientController) SyncOnChainChannelStates(tokenType entity.TokenType, tokenAddr string) error {
	_, err := cc.apiClient.SyncOnChainPaymentChannelStatus(
		context.Background(),
		&rpc.TokenInfo{
			TokenType:    tokenType,
			TokenAddress: tokenAddr,
		})
	return err
}

func (cc *ClientController) NewAppChannelOnVirtualContract(
	byteCode []byte, constructor []byte, nonce uint64, timeout uint64) (string, error) {
	sessionID, err := cc.apiClient.CreateAppSessionOnVirtualContract(
		context.Background(),
		&rpc.CreateAppSessionOnVirtualContractRequest{
			ContractBin:         ctype.Bytes2Hex(byteCode),
			ContractConstructor: ctype.Bytes2Hex(constructor),
			Nonce:               nonce,
			OnChainTimeout:      timeout,
		})
	if err != nil {
		return "", err
	}
	return sessionID.SessionId, nil
}

func (cc *ClientController) NewAppChannelOnDeployedContract(
	contractAddress string, nonce uint64, participants []string, timeout uint64) (string, error) {
	sessionID, err := cc.apiClient.CreateAppSessionOnDeployedContract(
		context.Background(),
		&rpc.CreateAppSessionOnDeployedContractRequest{
			ContractAddress: contractAddress,
			Nonce:           nonce,
			Participants:    participants,
			OnChainTimeout:  timeout,
		})
	if err != nil {
		return "", err
	}
	return sessionID.SessionId, nil
}

func (cc *ClientController) SettleAppChannel(cid string, stateproof []byte) error {
	_, err := cc.apiClient.SettleAppSession(
		context.Background(),
		&rpc.SettleAppSessionRequest{
			SessionId:  cid,
			StateProof: stateproof,
		})
	return err
}

func (cc *ClientController) SettleAppChannelBySigTimeout(cid string, oracleProof []byte) error {
	_, err := cc.apiClient.SettleAppSessionBySigTimeout(
		context.Background(),
		&rpc.SettleAppSessionByTimeoutRequest{
			SessionId:   cid,
			OracleProof: oracleProof,
		})
	return err
}

func (cc *ClientController) SettleAppChannelByMoveTimeout(cid string, oracleProof []byte) error {
	_, err := cc.apiClient.SettleAppSessionByMoveTimeout(
		context.Background(),
		&rpc.SettleAppSessionByTimeoutRequest{
			SessionId:   cid,
			OracleProof: oracleProof,
		})
	return err
}

func (cc *ClientController) SettleAppChannelByInvalidTurn(cid string, oracleProof []byte, cosignedStateProof []byte) error {
	_, err := cc.apiClient.SettleAppSessionByInvalidTurn(
		context.Background(),
		&rpc.SettleAppSessionByInvalidityRequest{
			SessionId:          cid,
			OracleProof:        oracleProof,
			CosignedStateProof: cosignedStateProof,
		})
	return err
}

func (cc *ClientController) SettleAppChannelByInvalidState(cid string, oracleProof []byte, cosignedStateProof []byte) error {
	_, err := cc.apiClient.SettleAppSessionByInvalidState(
		context.Background(),
		&rpc.SettleAppSessionByInvalidityRequest{
			SessionId:          cid,
			OracleProof:        oracleProof,
			CosignedStateProof: cosignedStateProof,
		})
	return err
}

func (cc *ClientController) DeleteAppChannel(cid string) error {
	_, err := cc.apiClient.DeleteAppSession(
		context.Background(),
		&rpc.SessionID{
			SessionId: cid,
		})
	return err
}

func (cc *ClientController) GetAppChannelState(cid string, key int64) ([]byte, error) {
	resp, err := cc.apiClient.GetStateForAppSession(
		context.Background(),
		&rpc.GetStateForAppSessionRequest{
			SessionId: cid,
			Key:       key,
		})
	if err != nil {
		return nil, err
	}
	return resp.State, err
}

func (cc *ClientController) GetAppChannelBooleanOutcome(
	cid string, query []byte) (bool, bool, error) {
	resp, err := cc.apiClient.GetBooleanOutcomeForAppSession(
		context.Background(),
		&rpc.GetBooleanOutcomeForAppSessionRequest{
			SessionId: cid,
			Query:     query,
		})
	if err != nil {
		return false, false, err
	}
	return resp.Finalized, resp.Outcome, err
}

func (cc *ClientController) GetAppChannelSettleFinalizedTime(cid string) (uint64, error) {
	blkNum, err := cc.apiClient.GetSettleFinalizedTimeForAppSession(
		context.Background(),
		&rpc.SessionID{
			SessionId: cid,
		})
	if err != nil {
		return 0, err
	}
	return blkNum.BlockNumber, err
}

func (cc *ClientController) ApplyAppChannelAction(cid string, action []byte) error {
	_, err := cc.apiClient.ApplyActionForAppSession(
		context.Background(),
		&rpc.ApplyActionForAppSessionRequest{
			SessionId: cid,
			Action:    action,
		})
	return err
}

func (cc *ClientController) GetAppChannelActionDeadline(cid string) (uint64, error) {
	blkNum, err := cc.apiClient.GetActionDeadlineForAppSession(
		context.Background(),
		&rpc.SessionID{
			SessionId: cid,
		})
	if err != nil {
		return 0, err
	}
	return blkNum.BlockNumber, nil
}

func (cc *ClientController) FinalizeAppChannelOnActionTimeout(cid string) error {
	_, err := cc.apiClient.FinalizeOnActionTimeoutForAppSession(
		context.Background(),
		&rpc.SessionID{
			SessionId: cid,
		})
	return err
}

func (cc *ClientController) WaitUntilDeadline(deadline uint64) error {
	log.Infoln("Wait until deadline", deadline)
	for {
		current, err := cc.GetCurrentBlockNumber()
		if err != nil {
			return err
		}
		log.Infoln("-- current block number --", current)
		if current > deadline {
			return nil
		}
		err = AdvanceBlocks(1)
		if err != nil {
			return err
		}
	}
}

func (cc *ClientController) SetMsgDropper(dropRecv, dropSend bool) error {
	_, err := cc.apiClient.SetMsgDropper(
		context.Background(),
		&rpc.SetMsgDropReq{
			DropRecv: dropRecv,
			DropSend: dropSend,
		})
	return err
}
