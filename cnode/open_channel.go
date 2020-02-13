// Copyright 2018 Celer Network

package cnode

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler-oss/chain"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/cnode/openchannelts"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/common/intfs"
	"github.com/celer-network/goCeler-oss/common/structs"
	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/fsm"
	"github.com/celer-network/goCeler-oss/metrics"
	"github.com/celer-network/goCeler-oss/monitor"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/rtconfig"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/transactor"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	openChannelRpcDeadlineSec    = 5
	maxOpenChannelTimeoutMinutes = 10.0
)

type routingTableBuilder interface {
	AddEdge(p1 ctype.Addr, p2 ctype.Addr, cid ctype.CidType, tokenAddr ctype.Addr) error
	GetOspInfo() map[ctype.Addr]uint64
	Build(tokenAddr ctype.Addr) (map[ctype.Addr]ctype.CidType, error)
}
type openChannelProcessor struct {
	nodeConfig          common.GlobalNodeConfig
	signer              common.Crypto
	transactor          *transactor.Transactor
	dal                 *storage.DAL
	ledger              *chain.BoundContract
	connectionManager   *rpc.ConnectionManager
	monitorService      intfs.MonitorService
	callbacks           map[string]event.OpenChannelCallback // TokenAddr to OpenChannelEvent callback
	callbacksLock       sync.Mutex
	utils               openChannelUtils
	eventMonitorName    string
	masterLock          sync.Mutex
	lockPerTokenPerPeer map[string]*sync.Mutex
	routeBuilder        routingTableBuilder
	// keepMonitor describes whether this instance is constantly monitoring on-chain open channel event.
	// clients only monitors open channel when they initialize the process while OSP is constantly monitoring.
	// There is a monitor bit to persist if a client was in monitoring state before crash or restart.
	// OSP who initialize open channel process with another osp will ignore the bit.
	keepMonitor bool
}

type openChannelUtils interface {
	GetCurrentBlockNumber() *big.Int
}

func startOpenChannelProcessor(
	nodeConfig common.GlobalNodeConfig,
	signer common.Crypto,
	transactor *transactor.Transactor,
	dal *storage.DAL,
	ledger *chain.BoundContract,
	connectionManager *rpc.ConnectionManager,
	monitorService intfs.MonitorService,
	utils openChannelUtils,
	routeBuilder routingTableBuilder,
	keepMonitor bool) (*openChannelProcessor, error) {
	p := &openChannelProcessor{
		nodeConfig:        nodeConfig,
		signer:            signer,
		transactor:        transactor,
		dal:               dal,
		ledger:            ledger,
		connectionManager: connectionManager,
		monitorService:    monitorService,
		callbacks:         make(map[string]event.OpenChannelCallback),
		utils:             utils,
		routeBuilder:      routeBuilder,
		eventMonitorName: fmt.Sprintf(
			"%s-%s", ctype.Addr2Hex(ledger.GetAddr()), event.OpenChannel),
		lockPerTokenPerPeer: make(map[string]*sync.Mutex),
		keepMonitor:         keepMonitor,
	}
	if keepMonitor {
		p.monitorEvent()
	} else {
		// Restore event monitoring
		has, err := p.dal.HasEventMonitorBit(p.eventMonitorName)
		if err != nil {
			return nil, err
		}
		if has {
			p.monitorSingleEvent(false)
		}
	}
	return p, nil
}

func invokeErrorCallback(openCallback event.OpenChannelCallback, err *common.E) {
	if openCallback != nil {
		go func() {
			openCallback.HandleOpenChannelErr(err)
		}()
	}
}
func (p *openChannelProcessor) prepareChannelInitializer(
	peer ctype.Addr,
	amtSelf *big.Int,
	amtPeer *big.Int,
	tokenInfo *entity.TokenInfo,
) (*entity.PaymentChannelInitializer, error) {
	if tokenInfo.TokenType == entity.TokenType_ERC20 && !ethcommon.IsHexAddress(ctype.Bytes2Hex(tokenInfo.TokenAddress)) {
		log.Errorln(common.ErrInvalidTokenAddress.Error())
		return nil, common.ErrInvalidTokenAddress
	}
	if amtSelf == nil || amtPeer == nil {
		log.Error("amt is nil")
		return nil, common.ErrInvalidAmount
	}

	selfAddress := p.transactor.Address().Bytes()
	// Construct request
	msgValueReceiver := uint64(0)
	lowAddrDist := &entity.AccountAmtPair{
		Account: selfAddress,
		Amt:     amtSelf.Bytes(),
	}
	highAddrDist := &entity.AccountAmtPair{
		Account: peer.Bytes(),
		Amt:     amtPeer.Bytes(),
	}
	// We need ascending order in terms of address
	if bytes.Compare(peer.Bytes(), selfAddress) == -1 {
		lowAddrDist, highAddrDist = highAddrDist, lowAddrDist
		msgValueReceiver = 1
	}
	initializer := &entity.PaymentChannelInitializer{
		InitDistribution: &entity.TokenDistribution{
			Token: tokenInfo,
			Distribution: []*entity.AccountAmtPair{
				lowAddrDist, highAddrDist,
			},
		},
		DisputeTimeout:   config.ChannelDisputeTimeout,
		MsgValueReceiver: msgValueReceiver,
	}
	return initializer, nil
}

func (p *openChannelProcessor) openChannel(
	peer ctype.Addr,
	amtSelf *big.Int,
	amtPeer *big.Int,
	tokenInfo *entity.TokenInfo,
	ospToOspOpen bool,
	openCallback event.OpenChannelCallback) error {
	initializer, err := p.prepareChannelInitializer(peer, amtSelf, amtPeer, tokenInfo)
	if err != nil {
		invokeErrorCallback(openCallback, &common.E{Reason: err.Error(), Code: 1})
		return err
	}
	// Record call back
	if openCallback != nil {
		tokenAddr := common.EthContractAddr
		if tokenInfo.GetTokenType() == entity.TokenType_ERC20 {
			tokenAddr = ctype.Bytes2Hex(tokenInfo.GetTokenAddress())
		}
		p.callbacksLock.Lock()
		p.callbacks[tokenAddr] = openCallback
		p.callbacksLock.Unlock()
	}
	initializer.OpenDeadline = p.utils.GetCurrentBlockNumber().Uint64() + config.OpenChannelTimeout
	initializerBytes, err := proto.Marshal(initializer)
	if err != nil {
		p.processOpenError(openCallback, err)
		return err
	}
	sig, err := p.signer.Sign(initializerBytes)
	if err != nil {
		p.processOpenError(openCallback, err)
		return err
	}
	openBy := rpc.OpenChannelBy_OPEN_CHANNEL_PROPOSER
	// Cold bootstrap
	if big.NewInt(0).Cmp(amtSelf) == 0 {
		openBy = rpc.OpenChannelBy_OPEN_CHANNEL_APPROVER
	}

	req := &rpc.OpenChannelRequest{
		ChannelInitializer: initializerBytes,
		RequesterSig:       sig,
		OpenBy:             openBy,
		OspToOsp:           ospToOspOpen,
	}

	rc, err := p.connectionManager.GetClient(ctype.Addr2Hex(peer))
	if err != nil {
		p.processOpenError(openCallback, err)
		return err
	}

	if !p.keepMonitor {
		// If the open-channel monitor (watcher) was auto-restarted after
		// a crash, there is nothing left to do.
		has, hasBitErr := p.dal.HasEventMonitorBit(p.eventMonitorName)
		if hasBitErr != nil {
			log.Debugln("CelerOpenChannel cannot check event monitor bit:", hasBitErr)
			p.processOpenError(openCallback, hasBitErr)
			return hasBitErr
		} else if has {
			log.Debugln("CelerOpenChannel event monitor already running")
			return nil
		}

		// Before sending the request, start monitoring until config.OpenChannelTimeout
		p.monitorSingleEvent(true)
		putBitErr := p.dal.PutEventMonitorBit(p.eventMonitorName)
		if putBitErr != nil {
			log.Debugln("CelerOpenChannel cannot put event monitor bit:", putBitErr)
			p.processOpenError(openCallback, putBitErr)
			return putBitErr
		}
	}
	resp, err := rc.CelerOpenChannel(context.Background(), req)
	if err != nil {
		log.Debugln("CelerOpenChannel rpc error", err)
		p.processOpenError(openCallback, err)
		return err
	}
	log.Debugln("OpenChannelResponse status", resp.Status)
	switch resp.Status {
	case rpc.OpenChannelStatus_UNDEFINED_OPEN_CHANNEL_STATUS:
		undefinedStatusErr := errors.New("Status undefined in response")
		p.processOpenError(openCallback, undefinedStatusErr)
		return undefinedStatusErr
	case rpc.OpenChannelStatus_OPEN_CHANNEL_TX_SUBMITTED:
		return nil
	case rpc.OpenChannelStatus_OPEN_CHANNEL_APPROVED:
		// Approved, send transaction
	}
	txValue := amtSelf
	if tokenInfo.TokenType == entity.TokenType_ERC20 {
		approveErr := p.approveErc20Allowance(amtSelf, tokenInfo, openCallback)
		if approveErr != nil {
			p.processOpenError(openCallback, approveErr)
			return approveErr
		}
		txValue = ctype.ZeroBigInt
	}

	log.Debug("Sending OpenChannel tx")
	return p.sendOpenChannelTransaction(
		resp, p.transactor.Address().Bytes(), peer.Bytes(), txValue, openCallback)
}
func (p *openChannelProcessor) approveErc20Allowance(amtSelf *big.Int, tokenInfo *entity.TokenInfo, openCallback event.OpenChannelCallback) error {
	// Deposit ERC20
	tokenAddress := ctype.Bytes2Addr(tokenInfo.TokenAddress)
	log.Debugln("Token address:", ctype.Addr2Hex(tokenAddress))
	erc20, err := chain.NewERC20Caller(tokenAddress, p.transactor.ContractCaller())
	if err != nil {
		p.processOpenError(openCallback, err)
		return err
	}
	owner := p.transactor.Address()
	spender := p.nodeConfig.GetLedgerContract().GetAddr()
	allowance, err := erc20.Allowance(&bind.CallOpts{}, owner, spender)
	if err != nil {
		p.processOpenError(openCallback, err)
		return err
	}
	if allowance.Cmp(amtSelf) < 0 {
		receiptChan := make(chan *types.Receipt, 1)
		_, err = p.transactor.Transact(
			&transactor.TransactionMinedHandler{
				OnMined: func(receipt *types.Receipt) {
					receiptChan <- receipt
				},
			},
			big.NewInt(0),
			func(
				transactor bind.ContractTransactor,
				opts *bind.TransactOpts) (*types.Transaction, error) {
				erc20, err := chain.NewERC20Transactor(tokenAddress, transactor)
				if err != nil {
					return nil, err
				}
				return erc20.Approve(opts, spender, amtSelf)
			})
		receipt := <-receiptChan
		approveTxHash := receipt.TxHash.String()
		if receipt.Status == types.ReceiptStatusSuccessful {
			log.Debugf("Approve transaction 0x%x succeeded", approveTxHash)
		} else {
			errMsg := fmt.Sprintf("Approve transaction 0x%x failed", approveTxHash)
			log.Error(errMsg)
			approveErr := errors.New(errMsg)
			p.processOpenError(openCallback, approveErr)
			return approveErr
		}
	}
	return nil
}

func (p *openChannelProcessor) sendOpenChannelTransaction(
	peerResponse *rpc.OpenChannelResponse,
	requesterAddr []byte,
	approverAddr []byte,
	txValue *big.Int,
	openCallback event.OpenChannelCallback) error {
	sig0 := peerResponse.RequesterSig
	sig1 := peerResponse.ApproverSig
	if bytes.Compare(requesterAddr, approverAddr) == 1 {
		sig0, sig1 = sig1, sig0
	}
	req := &chain.OpenChannelRequest{
		ChannelInitializer: peerResponse.ChannelInitializer,
		Sigs:               [][]byte{sig0, sig1},
	}
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		log.Debugln("Marshal OpenChannelRequest error", err)
		p.processOpenError(openCallback, err)
		return err
	}
	initializer := &entity.PaymentChannelInitializer{}
	marshalErr := proto.Unmarshal(peerResponse.ChannelInitializer, initializer)
	if marshalErr != nil {
		log.Errorf("Can't unmarshal initializer addr 0x%x", requesterAddr)
		return marshalErr
	}
	log.Infoln("Sending OpenChannel:", utils.PrintChannelInitializer(initializer))
	_, err = p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				txHash := receipt.TxHash
				tokenAddr := initializer.InitDistribution.Token.TokenAddress
				if receipt.Status == types.ReceiptStatusSuccessful {
					log.Debugf("OpenChannel transaction 0x%x succeeded, addr 0x%x, token 0x%x", txHash, requesterAddr, tokenAddr)
				} else {
					errMsg := fmt.Sprintf("OpenChannel transaction 0x%x failed, addr 0x%x, token 0x%x", txHash, requesterAddr, tokenAddr)
					log.Error(errMsg)
					openErr := errors.New(errMsg)
					p.processOpenError(openCallback, openErr)
					recordErr := p.dal.Transactional(
						p.recordOpenChannelFinishTx, ctype.Bytes2Addr(requesterAddr), ctype.Bytes2Addr(tokenAddr))
					if recordErr != nil {
						log.Errorln(recordErr, ctype.Bytes2Hex(requesterAddr))
					}
				}
			},
		},
		txValue,
		func(
			transactor bind.ContractTransactor,
			opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := ledger.NewCelerLedgerTransactor(p.ledger.GetAddr(), transactor)
			if err2 != nil {
				p.processOpenError(openCallback, err2)
				return nil, err2
			}
			return contract.OpenChannel(opts, reqBytes)
		})
	if err != nil {
		log.Errorf("%s, Can't open channel for 0x%x", err, requesterAddr)
		p.processOpenError(openCallback, err)
	}
	return err
}

func (p *openChannelProcessor) revertInflightOpenChannelTx(tx *storage.DALTx, args ...interface{}) error {
	peerAddr := args[0].(ctype.Addr)
	tokenAddr := args[1].(ctype.Addr)
	return tx.DeleteOpenChannelTs(peerAddr, tokenAddr)
}
func (p *openChannelProcessor) recordOpenChannelFinishTx(tx *storage.DALTx, args ...interface{}) error {
	peerAddr := args[0].(ctype.Addr)
	tokenAddr := args[1].(ctype.Addr)
	hasTs, err := tx.HasOpenChannelTs(peerAddr, tokenAddr)
	if err != nil {
		log.Errorln(err, "checking pending open channel for peer:", peerAddr.Hex(), "token:", tokenAddr)
		return err
	}
	if hasTs {
		timeNow := time.Now()
		ts, getErr := tx.GetOpenChannelTs(peerAddr, tokenAddr)
		if getErr != nil {
			log.Errorln(err, "getting open channel for peer:", peerAddr.Hex(), "token:", tokenAddr)
			return err
		}
		ts.FinishTs = &timeNow
		err = tx.PutOpenChannelTs(peerAddr, tokenAddr, ts)
		if err != nil {
			log.Errorln(err, "recording open channel for peer:", peerAddr.Hex(), "token:", tokenAddr)
			return err
		}
	}
	return nil
}
func (p *openChannelProcessor) recordInflightOpenChannelTx(tx *storage.DALTx, args ...interface{}) error {
	peerAddr := args[0].(ctype.Addr)
	tokenAddr := args[1].(ctype.Addr)
	allow := args[2].(*bool)
	*allow = false
	// Dedup open requests during processing open request.
	timeNow := time.Now()
	hasTs, err := tx.HasOpenChannelTs(peerAddr, tokenAddr)
	if err != nil {
		log.Errorln(err, "checking pending open channel for peer:", peerAddr.Hex(), "token:", tokenAddr)
		return err
	}
	if hasTs {
		ts, getErr := tx.GetOpenChannelTs(peerAddr, tokenAddr)
		if getErr != nil {
			log.Errorln(getErr, "checking pending open channel for peer:", peerAddr.Hex(), "token:", tokenAddr)
			return err
		}
		// No route on (peer, token) but FinishTs set, which means channel has been closed, allow.
		if ts.FinishTs != nil {
			err = tx.PutOpenChannelTs(peerAddr, tokenAddr, &openchannelts.OpenChannelTs{RequestTs: &timeNow})
			if err != nil {
				log.Errorln(err, "recording open channel for peer:", peerAddr.Hex(), "token:", tokenAddr)
				return err
			}
			*allow = true
			return nil
		}
		// Having ts but doesn't finish. Check if max deadline has passed.
		// This could happen for case e.g. user sent a open request but didn't send tx on-chain.
		if timeNow.Sub(*ts.RequestTs).Minutes() > maxOpenChannelTimeoutMinutes {
			err = tx.PutOpenChannelTs(peerAddr, tokenAddr, &openchannelts.OpenChannelTs{RequestTs: &timeNow})
			if err != nil {
				log.Errorln(err, "recording open channel for peer:", peerAddr.Hex(), "token:", tokenAddr)
				return err
			}
			*allow = true
			return nil
		}
		return nil
	}
	// No ts set, definitely allow.
	err = tx.PutOpenChannelTs(peerAddr, tokenAddr, &openchannelts.OpenChannelTs{RequestTs: &timeNow})
	if err != nil {
		log.Errorln(err, "recording open channel for peer:", peerAddr.Hex(), "token:", tokenAddr)
		return err
	}
	*allow = true
	return nil
}
func (p *openChannelProcessor) processOpenChannelRequest(req *rpc.OpenChannelRequest) (*rpc.OpenChannelResponse, error) {
	if req != nil {
		var initializer entity.PaymentChannelInitializer
		err := proto.Unmarshal(req.ChannelInitializer, &initializer)
		errResp := &rpc.OpenChannelResponse{
			Status: rpc.OpenChannelStatus_UNDEFINED_OPEN_CHANNEL_STATUS,
		}

		if err != nil {
			log.Error("Cannot parse channel initializer")
			return errResp, status.Error(codes.InvalidArgument, "channel initializer not parsable")
		}
		log.Infoln("process openchannel request", utils.PrintChannelInitializer(&initializer))
		// distribution is sorted based on address. So we need to figure out who's requester.
		accnt0 := initializer.InitDistribution.Distribution[0].Account
		accnt1 := initializer.InitDistribution.Distribution[1].Account
		myAddr := p.transactor.Address().Bytes()
		if bytes.Compare(accnt0, accnt1) != -1 {
			return errResp, status.Error(codes.InvalidArgument, "wrong distribution address order")
		}
		if bytes.Compare(myAddr, accnt0) != 0 && bytes.Compare(myAddr, accnt1) != 0 {
			return errResp, status.Error(codes.InvalidArgument, "wrong channel peers")
		}
		if RequestStandardDeposit(p.utils.GetCurrentBlockNumber().Uint64(), p.transactor.Address(), &initializer, req.GetOspToOsp())&AllowStandardOpenChannel == 0 {
			return errResp, status.Error(codes.InvalidArgument, "distribution breaks policy")
		}
		if initializer.DisputeTimeout > rtconfig.GetMaxDisputeTimeout() || initializer.DisputeTimeout < rtconfig.GetMinDisputeTimeout() {
			return errResp, status.Error(codes.InvalidArgument, "dipspute timeout breaks policy")
		}
		requester := initializer.InitDistribution.Distribution[0].Account
		approver := initializer.InitDistribution.Distribution[1].Account
		if bytes.Compare(requester, myAddr) == 0 {
			requester, approver = approver, requester
		}
		tokenAddr := initializer.InitDistribution.Token.TokenAddress

		tokenInfo := utils.GetTokenInfoFromAddress(ctype.Bytes2Addr(tokenAddr))
		// Critical section to open channel (for each requester, token pair).
		existingCid, exist, errCidByPeerToken := p.dal.GetCidByPeerAndTokenWithErr(requester, tokenInfo)
		if errCidByPeerToken != nil {
			return errResp, status.Error(codes.Internal, errCidByPeerToken.Error())
		}
		if exist && existingCid != ctype.ZeroCid {
			// Note that we don't delete the entry when a channel is closed. We set it to ZeroCid.
			// Getting a closed channel will not return error but the value will be zero.
			return errResp, status.Error(codes.AlreadyExists, "cid="+existingCid.Hex())
		}
		allowToOpen := false
		err = p.dal.Transactional(p.recordInflightOpenChannelTx, ctype.Bytes2Addr(requester), ctype.Bytes2Addr(tokenAddr), &allowToOpen)
		if err != nil {
			return errResp, status.Error(codes.Internal, err.Error())
		}
		if !allowToOpen {
			return errResp, status.Error(codes.AlreadyExists, "more than one inflight open channel request.")
		}

		mySig, err := p.signer.Sign(req.ChannelInitializer)
		if err != nil {
			revertErr := p.dal.Transactional(p.revertInflightOpenChannelTx, ctype.Bytes2Addr(requester), ctype.Bytes2Addr(tokenAddr))
			if revertErr != nil {
				log.Errorln(revertErr, ctype.Bytes2Hex(requester), ctype.Bytes2Hex(approver))
			}
			return errResp, status.Error(
				codes.InvalidArgument, "failed to sign channel initializer")
		}
		resp := &rpc.OpenChannelResponse{
			ChannelInitializer: req.ChannelInitializer,
			RequesterSig:       req.RequesterSig,
			ApproverSig:        mySig,
		}
		switch req.OpenBy {
		case rpc.OpenChannelBy_OPEN_CHANNEL_PROPOSER:
			resp.Status = rpc.OpenChannelStatus_OPEN_CHANNEL_APPROVED
			return &rpc.OpenChannelResponse{
				ChannelInitializer: req.GetChannelInitializer(),
				RequesterSig:       req.GetRequesterSig(),
				ApproverSig:        mySig,
				Status:             rpc.OpenChannelStatus_OPEN_CHANNEL_APPROVED,
			}, nil
		case rpc.OpenChannelBy_OPEN_CHANNEL_APPROVER:
			resp.Status = rpc.OpenChannelStatus_OPEN_CHANNEL_TX_SUBMITTED
			err := p.sendOpenChannelTransaction(
				resp, requester, approver, big.NewInt(0) /*txValue*/, nil /*openCallback*/)
			if err != nil {
				revertErr := p.dal.Transactional(p.revertInflightOpenChannelTx, ctype.Bytes2Addr(requester), ctype.Bytes2Addr(tokenAddr))
				if revertErr != nil {
					log.Errorln(revertErr, ctype.Bytes2Hex(requester), ctype.Bytes2Hex(approver))
				}
				return errResp, status.Error(codes.Internal, "Can't send open channel tx on-chain.")
			}
			return resp, nil
		default:
			return errResp, status.Error(codes.InvalidArgument, "OpenBy not set")
		}
	}
	return nil, nil
}
func (p *openChannelProcessor) computePscID(
	channelInitializerBytes []byte, ledgerAdr, walletAddr ethcommon.Address) (ctype.CidType, error) {
	nonce := crypto.Keccak256(channelInitializerBytes)
	// Does same thing as abi.encodePack in solidity.
	packed := make([]byte, 0, len(walletAddr.Bytes())+len(ledgerAdr.Bytes())+len(nonce))
	packed = append(packed, walletAddr.Bytes()...)
	packed = append(packed, ledgerAdr.Bytes()...)
	packed = append(packed, nonce...)
	walledID := crypto.Keccak256(packed)
	return ctype.Bytes2Cid(walledID), nil
}

type openedChannelDescriptor struct {
	cid          ctype.CidType
	participants [2]ethcommon.Address
	initDeposits [2]*big.Int
	tokenType    entity.TokenType
	tokenAddress ethcommon.Address
}

func (p *openChannelProcessor) maybeHandleEvent(descriptor *openedChannelDescriptor,
	channelFsmTransfer func(tx *storage.DALTx, args ...interface{}) error) bool {
	cid := descriptor.cid
	log.Infoln("Handle open channel", cid.Hex())

	if len(descriptor.participants) != 2 || len(descriptor.initDeposits) != 2 {
		log.Error("on chain balances length not match")
		return false
	}
	self := p.transactor.Address()
	var myIndex int
	if descriptor.participants[0] == self {
		myIndex = 0
	} else if descriptor.participants[1] == self {
		myIndex = 1
	} else {
		return false
	}
	peer := descriptor.participants[1-myIndex]
	onChainBalance := &structs.OnChainBalance{
		MyDeposit:      descriptor.initDeposits[myIndex],
		MyWithdrawal:   big.NewInt(0),
		PeerDeposit:    descriptor.initDeposits[1-myIndex],
		PeerWithdrawal: big.NewInt(0),
	}

	p.dal.PutTokenContractAddr(cid, ctype.Addr2Hex(descriptor.tokenAddress))
	tokenAddr := ctype.Addr2Hex(descriptor.tokenAddress)

	selfStr := ctype.Addr2Hex(self)
	peerStr := ctype.Addr2Hex(peer)
	txBody := func(tx *storage.DALTx, args ...interface{}) error {
		err := tx.PutCidForPeerAndToken(
			peer.Bytes(), utils.GetTokenInfoFromAddress(descriptor.tokenAddress), cid)
		if err != nil {
			log.Error(err)
			return err
		}
		hasCState, err := tx.HasChannelState(cid)
		if err != nil {
			log.Errorf("%s, can't get channel state %s", err, cid.Hex())
			return err
		}
		if hasCState {
			log.Errorf("Receiving open event with existing state, cid %s", cid.Hex())
			return common.ErrOpenEventOnWrongState
		}

		if exist, err := tx.HasPeer(cid); err != nil {
			return err
		} else if !exist {
			if err := tx.PutPeer(cid, peerStr); err != nil {
				return err
			}
		}
		mySimplex, _, _ := tx.GetSimplexPaymentChannel(cid, selfStr)
		if mySimplex == nil {
			emptySimplex := &entity.SimplexPaymentChannel{
				ChannelId: descriptor.cid[:],
				PeerFrom:  self.Bytes(),
				SeqNum:    0,
				TransferToPeer: &entity.TokenTransfer{
					Token: &entity.TokenInfo{
						TokenType:    descriptor.tokenType,
						TokenAddress: descriptor.tokenAddress.Bytes(),
					},
					Receiver: &entity.AccountAmtPair{
						Amt:     []byte{0},
						Account: peer.Bytes(),
					},
				},
				PendingPayIds:          &entity.PayIdList{},
				LastPayResolveDeadline: 0,
				TotalPendingAmount:     []byte{0},
			}
			emptySimplexByte, _ := proto.Marshal(emptySimplex)
			mySig, _ := p.signer.Sign(emptySimplexByte)
			simplexState := &rpc.SignedSimplexState{
				SimplexState:  emptySimplexByte,
				SigOfPeerFrom: mySig,
			}
			tx.PutSimplexState(cid, selfStr, simplexState)
		}
		peerSimplex, _, _ := tx.GetSimplexPaymentChannel(cid, peerStr)
		if peerSimplex == nil {
			emptySimplex := &entity.SimplexPaymentChannel{
				ChannelId: descriptor.cid[:],
				PeerFrom:  peer.Bytes(),
				SeqNum:    0,
				TransferToPeer: &entity.TokenTransfer{
					Token: &entity.TokenInfo{
						TokenType:    descriptor.tokenType,
						TokenAddress: descriptor.tokenAddress.Bytes(),
					},
					Receiver: &entity.AccountAmtPair{
						Amt:     []byte{0},
						Account: self.Bytes(),
					},
				},
				PendingPayIds:          &entity.PayIdList{},
				LastPayResolveDeadline: 0,
				TotalPendingAmount:     []byte{0},
			}
			emptySimplexByte, _ := proto.Marshal(emptySimplex)
			mySig, _ := p.signer.Sign(emptySimplexByte)
			simplexState := &rpc.SignedSimplexState{
				SimplexState: emptySimplexByte,
				SigOfPeerTo:  mySig,
			}
			tx.PutSimplexState(cid, peerStr, simplexState)
		}
		return tx.PutOnChainBalance(cid, onChainBalance)
	}
	if err := p.dal.Transactional(txBody); err != nil {
		log.Error(err)
		return false
	}
	if err := p.dal.Transactional(channelFsmTransfer, cid); err != nil {
		log.Debug(err)
		return false
	}
	recordErr := p.dal.Transactional(
		p.recordOpenChannelFinishTx, peer, descriptor.tokenAddress)
	if recordErr != nil {
		log.Errorln(recordErr, peer)
	}
	// Non-blocking notification.
	p.callbacksLock.Lock()
	cb, ok := p.callbacks[tokenAddr]
	if ok {
		go cb.HandleOpenChannelFinish(cid)
		delete(p.callbacks, tokenAddr)
	}
	p.callbacksLock.Unlock()
	return true
}

func (p *openChannelProcessor) monitorEvent() {
	_, err := p.monitorService.Monitor(
		event.OpenChannel,
		p.ledger,
		p.utils.GetCurrentBlockNumber(),
		nil,
		false, /* quickCatch */
		false,
		func(id monitor.CallbackID, eLog types.Log) {
			e := &ledger.CelerLedgerOpenChannel{}
			if err := p.ledger.ParseEvent(event.OpenChannel, eLog, e); err != nil {
				log.Error(err)
			}
			channelDescriptor := &openedChannelDescriptor{
				cid:          ctype.CidType(e.ChannelId),
				participants: e.PeerAddrs,
				initDeposits: e.InitialDeposits,
				tokenAddress: e.TokenAddress,
				tokenType:    entity.TokenType(e.TokenType.Int64()),
			}
			// set channel open state to err and maybe change it after event hanlded
			chanOpen := metrics.CNodeOpenChanErr
			if p.maybeHandleEvent(channelDescriptor, fsm.OnPscAuthOpen) {
				chanOpen = metrics.CNodeOpenChanOK
			}
			if len(e.PeerAddrs) == 2 && p.routeBuilder != nil {
				// check length to avoid seg fault if there is any unexpected PeerAddrs.
				go func() {
					p.routeBuilder.AddEdge(e.PeerAddrs[0], e.PeerAddrs[1], e.ChannelId, e.TokenAddress)
					// trigger rt rebuild if the edge is between osps.
					osps := p.routeBuilder.GetOspInfo()
					_, p0IsOsp := osps[e.PeerAddrs[0]]
					_, p1IsOsp := osps[e.PeerAddrs[1]]
					if p0IsOsp && p1IsOsp {
						p.routeBuilder.Build(e.TokenAddress)
					}
				}()
			}

			metrics.IncCNodeOpenChanEventCnt(metrics.CNodeStandardChan, chanOpen)
		})
	if err != nil {
		log.Error(err)
	}
}

func (p *openChannelProcessor) monitorSingleEvent(reset bool) {
	startBlock := p.utils.GetCurrentBlockNumber()
	duration := new(big.Int)
	duration.SetUint64(config.OpenChannelTimeout)
	endBlock := new(big.Int).Add(startBlock, duration)
	_, err := p.monitorService.Monitor(
		event.OpenChannel,
		p.ledger,
		startBlock,
		endBlock,
		false, /* quickCatch */
		reset,
		func(id monitor.CallbackID, eLog types.Log) {
			e := &ledger.CelerLedgerOpenChannel{}
			if err := p.ledger.ParseEvent(event.OpenChannel, eLog, e); err != nil {
				log.Error(err)
			}
			channelDescriptor := &openedChannelDescriptor{
				cid:          ctype.CidType(e.ChannelId),
				participants: e.PeerAddrs,
				initDeposits: e.InitialDeposits,
				tokenAddress: e.TokenAddress,
				tokenType:    entity.TokenType(e.TokenType.Int64()),
			}
			if p.maybeHandleEvent(channelDescriptor, fsm.OnPscAuthOpen) {
				go func() {
					if len(e.PeerAddrs) == 2 && p.routeBuilder != nil {
						// check length to avoid seg fault if there is any unexpected PeerAddrs.
						p.routeBuilder.AddEdge(e.PeerAddrs[0], e.PeerAddrs[1], e.ChannelId, e.TokenAddress)
					}
				}()
				p.monitorService.RemoveEvent(id)
				p.dal.DeleteEventMonitorBit(p.eventMonitorName)
			}
		})
	if err != nil {
		log.Error(err)
	}
}

func initDistributionBreaksPolicy(initDist *entity.TokenDistribution, myAddr []byte) bool {
	// Distribution is sorted based on address. Need to figure out who's me.
	requesterDist := initDist.Distribution[0]
	myDist := initDist.Distribution[1]
	if bytes.Compare(myAddr, requesterDist.Account) == 0 {
		requesterDist, myDist = myDist, requesterDist
	}
	requesterDeposit := big.NewInt(0)
	requesterDeposit.SetBytes(requesterDist.Amt)
	myDeposit := big.NewInt(0)
	myDeposit.SetBytes(myDist.Amt)

	if requesterDeposit.Cmp(big.NewInt(0)) == 0 {
		// 0-balance bootstrap
		myMax := big.NewInt(0)
		tokenType := initDist.Token.TokenType
		if tokenType == entity.TokenType_ETH {
			myMax = rtconfig.GetEthColdBootstrapDeposit()
		} else if tokenType == entity.TokenType_ERC20 {
			myMax = rtconfig.GetErc20ColdBootstrapDeposit(initDist.Token.TokenAddress)
		}
		if myMax.Cmp(myDeposit) == -1 {
			return true
		}
	} else {
		// Otherwise, don't deposit more than a multiplier
		if myDeposit.Cmp(requesterDeposit.Mul(requesterDeposit, big.NewInt(rtconfig.GetOspDepositMultiplier()))) == 1 {
			return true
		}
	}
	return false
}

func (c *CNode) OpenChannel(
	peer ctype.Addr,
	amtSelf *big.Int,
	amtPeer *big.Int,
	tokenInfo *entity.TokenInfo,
	ospToOspOpen bool,
	openCallback event.OpenChannelCallback) error {
	return c.openChannelProcessor.openChannel(peer, amtSelf, amtPeer, tokenInfo, ospToOspOpen, openCallback)
}

func (c *CNode) ProcessOpenChannelRequest(in *rpc.OpenChannelRequest) (*rpc.OpenChannelResponse, error) {
	response, err := c.openChannelProcessor.processOpenChannelRequest(in)
	if err != nil {
		log.Error(err)
	}
	return response, err
}

func (p *openChannelProcessor) processOpenError(openCallback event.OpenChannelCallback, err error) {
	if !p.keepMonitor {
		p.dal.DeleteEventMonitorBit(p.eventMonitorName)
	}
	invokeErrorCallback(openCallback, &common.E{Reason: err.Error(), Code: 1})
}
