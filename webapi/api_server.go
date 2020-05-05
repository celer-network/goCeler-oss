// Copyright 2018-2020 Celer Network

package webapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/celer-network/goCeler/celersdk"
	"github.com/celer-network/goCeler/celersdkintf"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/cobj"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	msgrpc "github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/webapi/rpc"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

type ApiServer struct {
	webPort                   int
	grpcPort                  int
	allowedOrigins            string
	apiClient                 *celersdk.Client
	payIter                   *celersdk.PayHistoryIterator
	callbackImpl              *callbackImpl
	appSessionMap             map[string]*celersdk.AppSession
	appSessionMapLock         sync.Mutex
	appSessionCallbackMap     map[string]*appSessionCallback
	appSessionCallbackMapLock sync.Mutex
}

// implement celersdk.ExternalSignerCallback interface
// also embed common.Signer so celersdk.ExternalSignerManager can tell the difference and
// avoid double hash
type extSigner struct {
	common.Signer
}

func (es *extSigner) OnSignMessage(reqid int, msg []byte) {
	res, _ := es.SignEthMessage(msg)
	celersdk.PublishSignedResult(reqid, res)
}

func (es *extSigner) OnSignTransaction(reqid int, rawtx []byte) {
	res, _ := es.SignEthTransaction(rawtx)
	celersdk.PublishSignedResult(reqid, res)
}

func NewApiServer(
	webPort int,
	grpcPort int,
	allowedOrigins string,
	keystore string,
	password string,
	dataPath string,
	config string,
	useExtSigner bool) *ApiServer {
	callbackImpl := NewCallbackImpl()
	s := &ApiServer{
		webPort:               webPort,
		grpcPort:              grpcPort,
		allowedOrigins:        allowedOrigins,
		callbackImpl:          callbackImpl,
		appSessionMap:         make(map[string]*celersdk.AppSession),
		appSessionCallbackMap: make(map[string]*appSessionCallback),
	}
	if !useExtSigner {
		go celersdk.InitClient(
			&celersdk.Account{Keystore: keystore, Password: password},
			config,
			dataPath,
			callbackImpl)
	} else { // exercise external signer code path
		addr, priv, _ := utils.GetAddrAndPrivKey(keystore, password)
		signer, _ := cobj.NewCelerSigner(priv)
		go celersdk.InitClientWithSigner(ctype.Addr2Hex(addr), config, dataPath, callbackImpl, &extSigner{signer})
	}

	select {
	case client := <-callbackImpl.clientReady:
		s.apiClient = client
	case err := <-callbackImpl.clientInitErr:
		log.Fatal(err)
		return nil
	}
	return s
}

func (s *ApiServer) Start() {
	gs := grpc.NewServer()
	rpc.RegisterWebApiServer(gs, s)
	s.serve(gs)
}

func (s *ApiServer) serve(gs *grpc.Server) {
	errChan := make(chan error)
	if s.grpcPort != -1 {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		addr := "localhost:" + strconv.Itoa(s.grpcPort)
		log.Info("Serving Celer gRPC APIs on http://" + addr)
		go func() {
			errChan <- gs.Serve(lis)
		}()
	}
	if s.webPort != -1 {
		wrappedSvr := grpcweb.WrapServer(gs)
		addr := "localhost:" + strconv.Itoa(s.webPort)
		httpSvr := &http.Server{
			Addr: addr,
			Handler: cors.New(cors.Options{
				AllowedHeaders: []string{
					"Origin", "Accept", "Content-Type", "X-Requested-With", "X-Grpc-Web"},
				AllowedOrigins:   strings.Split(s.allowedOrigins, ","),
				AllowCredentials: true,
			}).Handler(wrappedSvr),
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       120 * time.Second,
		}
		log.Info("Serving Celer Web APIs on http://" + addr)
		go func() {
			errChan <- httpSvr.ListenAndServe()
		}()
	}
	err := <-errChan
	gs.GracefulStop()
	log.Fatal(err)
}
func (s *ApiServer) SetDelegation(context context.Context, request *rpc.SetDelegationRequest) (*empty.Empty, error) {
	log.Debugf("ApiServer: SetDelegation #tks %d duration %d", len(request.GetTokenInfos()), request.GetBlockDuration())
	tokens := make([]*celersdk.Token, 0, len(request.GetTokenInfos()))
	for _, tk := range request.GetTokenInfos() {
		token := &celersdk.Token{
			Erctype: "ERC20",
			Addr:    tk.GetTokenAddress(),
		}
		if tk.GetTokenAddress() == ctype.EthTokenAddrStr {
			token.Erctype = "ETH"
		}
		tokens = append(tokens, token)
	}
	err := s.apiClient.SetDelegation(tokens, request.GetBlockDuration())
	return &empty.Empty{}, err
}
func (s *ApiServer) OpenPaymentChannel(
	context context.Context, request *rpc.OpenPaymentChannelRequest) (*rpc.ChannelID, error) {
	callbackImpl := s.callbackImpl
	tokenInfo := request.TokenInfo
	switch entity.TokenType(tokenInfo.TokenType) {
	case entity.TokenType_ETH:
		go s.apiClient.OpenETHChannel(
			&celersdk.Deposit{Myamtwei: request.Amount, Peeramtwei: request.PeerAmount},
			s.callbackImpl)
	case entity.TokenType_ERC20:
		go s.apiClient.OpenTokenChannel(
			&celersdk.Token{Erctype: "ERC20", Addr: tokenInfo.TokenAddress},
			&celersdk.Deposit{Myamtwei: request.Amount, Peeramtwei: request.PeerAmount},
			s.callbackImpl)
	default:
		return nil, errors.New("Unknown token type")
	}
	select {
	case cid := <-callbackImpl.channelOpened:
		return &rpc.ChannelID{ChannelId: cid}, nil
	case errMsg := <-callbackImpl.openChannelError:
		return nil, errors.New(errMsg)
	}
}

type depositCallback struct {
	apiClient *celersdk.Client
	txHash    chan string
	err       chan string
}

func (cb *depositCallback) OnDeposit(jobID string, txHash string) {
	log.Infoln("deposit succeeded:", jobID, txHash)
	cb.apiClient.RemoveDepositJob(jobID)
	cb.txHash <- txHash
}

func (cb *depositCallback) OnError(jobID string, err string) {
	log.Infoln("deposit failed:", jobID, err)
	cb.apiClient.RemoveDepositJob(jobID)
	cb.err <- err
}

func (s *ApiServer) Deposit(
	context context.Context, request *rpc.DepositOrWithdrawRequest) (*rpc.DepositOrWithdrawJob, error) {
	cb := &depositCallback{
		apiClient: s.apiClient,
		txHash:    make(chan string),
		err:       make(chan string),
	}
	var err error
	var jobID string
	tokenInfo := request.TokenInfo
	switch entity.TokenType(tokenInfo.TokenType) {
	case entity.TokenType_ETH:
		jobID, err = s.apiClient.DepositETH(request.Amount, cb)
	case entity.TokenType_ERC20:
		jobID, err = s.apiClient.DepositERC20(
			&celersdk.Token{Erctype: "ERC20", Addr: tokenInfo.TokenAddress},
			request.Amount,
			cb)
	default:
		err = errors.New("Unknown token type")
	}
	if err != nil {
		return nil, err
	}
	select {
	case txHash := <-cb.txHash:
		return &rpc.DepositOrWithdrawJob{JobId: jobID, TxHash: txHash}, nil
	case errMsg := <-cb.err:
		return nil, errors.New(errMsg)
	}
}

func (s *ApiServer) MonitorDepositJob(
	context context.Context, request *rpc.DepositOrWithdrawJob) (*rpc.DepositOrWithdrawJob, error) {
	cb := &depositCallback{
		apiClient: s.apiClient,
		txHash:    make(chan string),
		err:       make(chan string),
	}
	s.apiClient.MonitorDepositJob(request.JobId, cb)
	select {
	case txHash := <-cb.txHash:
		return &rpc.DepositOrWithdrawJob{JobId: request.JobId, TxHash: txHash}, nil
	case errMsg := <-cb.err:
		return nil, errors.New(errMsg)
	}
}

type withdrawCallback struct {
	apiClient *celersdk.Client
	txHash    chan string
	err       chan string
}

func (cb *withdrawCallback) OnWithdraw(withdrawHash string, txHash string) {
	log.Infoln("withdrawal succeeded:", withdrawHash, txHash)
	cb.apiClient.RemoveCooperativeWithdrawJob(withdrawHash)
	cb.txHash <- txHash
}

func (cb *withdrawCallback) OnError(withdrawHash string, err string) {
	log.Errorln("withdrawal failed:", withdrawHash, err)
	cb.apiClient.RemoveCooperativeWithdrawJob(withdrawHash)
	cb.err <- err
}

func (s *ApiServer) CooperativeWithdraw(
	context context.Context, request *rpc.DepositOrWithdrawRequest) (*rpc.DepositOrWithdrawJob, error) {
	cb := &withdrawCallback{
		apiClient: s.apiClient,
		txHash:    make(chan string),
		err:       make(chan string),
	}

	var err error
	var jobID string
	tokenInfo := request.TokenInfo
	switch entity.TokenType(tokenInfo.TokenType) {
	case entity.TokenType_ETH:
		jobID, err = s.apiClient.WithdrawETH(request.Amount, cb)
	case entity.TokenType_ERC20:
		jobID, err = s.apiClient.WithdrawERC20(
			&celersdk.Token{Erctype: "ERC20", Addr: tokenInfo.TokenAddress},
			request.Amount,
			cb)
	default:
		err = errors.New("Unknown token type")
	}
	if err != nil {
		return nil, err
	}

	select {
	case txHash := <-cb.txHash:
		return &rpc.DepositOrWithdrawJob{JobId: jobID, TxHash: txHash}, nil
	case errMsg := <-cb.err:
		return nil, errors.New(errMsg)
	}
}

func (s *ApiServer) MonitorCooperativeWithdrawJob(
	context context.Context, request *rpc.DepositOrWithdrawJob) (*rpc.DepositOrWithdrawJob, error) {
	cb := &withdrawCallback{
		apiClient: s.apiClient,
		txHash:    make(chan string),
		err:       make(chan string),
	}
	s.apiClient.MonitorCooperativeWithdrawJob(request.JobId, cb)
	select {
	case txHash := <-cb.txHash:
		return &rpc.DepositOrWithdrawJob{JobId: request.JobId, TxHash: txHash}, nil
	case errMsg := <-cb.err:
		return nil, errors.New(errMsg)
	}
}

func (s *ApiServer) GetBalance(
	context context.Context, request *rpc.TokenInfo) (*rpc.GetBalanceResponse, error) {
	var balance *celersdk.Balance
	var err error
	switch entity.TokenType(request.TokenType) {
	case entity.TokenType_ETH:
		balance, err = s.apiClient.GetBalance()
	case entity.TokenType_ERC20:
		balance, err =
			s.apiClient.GetBalanceERC20(request.TokenAddress)
	default:
		return nil, errors.New("Unknown token type")
	}
	if err != nil {
		return nil, err
	}
	return &rpc.GetBalanceResponse{
		FreeBalance:       balance.Available,
		LockedBalance:     balance.Pending,
		ReceivingCapacity: balance.ReceivingCap,
	}, nil
}

func (s *ApiServer) GetPeerFreeBalance(
	context context.Context, request *rpc.GetPeerFreeBalanceRequest) (*rpc.FreeBalance, error) {
	var err error
	var status *celersdk.CelerStatus
	tokenInfo := request.TokenInfo
	switch tokenInfo.TokenType {
	case entity.TokenType_ETH:
		status, err = s.apiClient.QueryReceivingCapacity(request.PeerAddress)
	case entity.TokenType_ERC20:
		status, err = s.apiClient.QueryReceivingCapacityOnToken(tokenInfo.TokenAddress, request.PeerAddress)
	}
	if err != nil {
		return nil, err
	}
	return &rpc.FreeBalance{FreeBalance: status.FreeBalance, JoinStatus: status.JoinStatus}, nil
}

func (s *ApiServer) SendConditionalPayment(
	context context.Context,
	request *rpc.SendConditionalPaymentRequest) (*rpc.PaymentID, error) {
	conditions := request.Conditions
	tokenInfo := request.TokenInfo
	tokenType := entity.TokenType(tokenInfo.TokenType)
	if tokenType != entity.TokenType_ETH && tokenType != entity.TokenType_ERC20 {
		return nil, errors.New("Unknown token type")
	}
	var payID string
	var err error

	sdkConditions := make([]*celersdk.Condition, len(conditions))
	for i, condition := range conditions {
		sdkConditions[i] = &celersdk.Condition{
			OnChainDeployed: condition.OnChainDeployed,
			ContractAddress: ctype.Hex2Bytes(condition.ContractAddress),
			IsFinalizedArgs: condition.IsFinalizedArgs,
			GetOutcomeArgs:  condition.GetOutcomeArgs,
		}
	}
	payID, err = s.apiClient.SendConditionalPayment(
		&celersdk.TokenInfo{
			TokenType:    celersdk.TokenType(int32(tokenInfo.TokenType)),
			TokenAddress: tokenInfo.TokenAddress,
		},
		request.Destination,
		request.Amount,
		celersdk.TransferLogicType(int32(request.TransferLogicType)),
		sdkConditions,
		int64(request.Timeout),
		request.Note)
	if err != nil {
		return nil, err
	}
	return &rpc.PaymentID{PaymentId: payID}, nil
}

func (s *ApiServer) ConfirmOutgoingPayment(
	context context.Context, request *rpc.PaymentID) (*empty.Empty, error) {
	err := s.apiClient.ConfirmPay(request.PaymentId)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) RejectIncomingPayment(
	context context.Context, request *rpc.PaymentID) (*empty.Empty, error) {
	err := s.apiClient.RejectPay(request.PaymentId)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) SettleOnChainResolvedIncomingPayment(
	context context.Context, request *rpc.PaymentID) (*empty.Empty, error) {
	err := s.apiClient.SettleOnChainResolvedIncomingPayment(request.PaymentId)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) ResolveIncomingPaymentOnChain(
	context context.Context, request *rpc.PaymentID) (*empty.Empty, error) {
	err := s.apiClient.ResolveIncomingPaymentOnChain(request.PaymentId)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) GetOnChainPaymentInfo(
	context context.Context, request *rpc.PaymentID) (*rpc.OnChainPaymentInfo, error) {
	info, err := s.apiClient.GetOnChainPaymentInfo(request.PaymentId)
	if err != nil {
		return nil, err
	}
	return &rpc.OnChainPaymentInfo{Amount: info.Amount, ResolveDeadline: info.ResolveDeadline}, nil
}

func (s *ApiServer) SubscribeIncomingPayments(
	empty *empty.Empty, stream rpc.WebApi_SubscribeIncomingPaymentsServer) error {
	writeToStream := func(payment *celersdkintf.Payment) error {
		tokenAddr := ctype.Hex2Addr(payment.TokenAddr)
		var tokenType entity.TokenType
		if tokenAddr == ctype.Hex2Addr(ctype.EthTokenAddrStr) {
			tokenType = entity.TokenType_ETH
		} else {
			tokenType = entity.TokenType_ERC20
		}
		return stream.Send(&rpc.PaymentInfo{
			PaymentId: payment.UID,
			Sender:    payment.Sender,
			Receiver:  payment.Receiver,
			TokenInfo: &rpc.TokenInfo{
				TokenType:    tokenType,
				TokenAddress: payment.TokenAddr,
			},
			Amount:      payment.AmtWei,
			PaymentJson: payment.PayJSON,
			Status:      uint32(payment.Status),
		})
	}
	callbackImpl := s.callbackImpl
	for {
		select {
		case payment := <-callbackImpl.recvStart:
			err := writeToStream(payment)
			if err != nil {
				return err
			}
		case payment := <-callbackImpl.recvDone:
			err := writeToStream(payment)
			if err != nil {
				return err
			}
		}
	}
}

func (s *ApiServer) SubscribeOutgoingPayments(
	empty *empty.Empty, stream rpc.WebApi_SubscribeOutgoingPaymentsServer) error {
	writeToStream := func(payment *celersdkintf.Payment, errInfo *celersdkintf.E) error {
		tokenAddr := ctype.Hex2Addr(payment.TokenAddr)
		var tokenType entity.TokenType
		if tokenAddr == ctype.Hex2Addr(ctype.EthTokenAddrStr) {
			tokenType = entity.TokenType_ETH
		} else {
			tokenType = entity.TokenType_ERC20
		}
		var errReason string
		var errCode int64
		if errInfo != nil {
			errReason = errInfo.Reason
			errCode = int64(errInfo.Code)
		}
		return stream.Send(&rpc.OutgoingPaymentInfo{
			Payment: &rpc.PaymentInfo{
				PaymentId: payment.UID,
				Sender:    payment.Sender,
				Receiver:  payment.Receiver,
				TokenInfo: &rpc.TokenInfo{
					TokenType:    tokenType,
					TokenAddress: payment.TokenAddr,
				},
				Amount:      payment.AmtWei,
				PaymentJson: payment.PayJSON,
				Status:      uint32(payment.Status),
			},
			ErrorReason: errReason,
			ErrorCode:   errCode,
		})
	}
	callbackImpl := s.callbackImpl
	for {
		select {
		case payment := <-callbackImpl.sendComplete:
			err := writeToStream(payment, nil)
			if err != nil {
				return err
			}
		case pair := <-callbackImpl.sendErr:
			err := writeToStream(pair.pay, pair.e)
			if err != nil {
				return err
			}
		}
	}
}

func (s *ApiServer) GetIncomingPaymentStatus(
	context context.Context, request *rpc.PaymentID) (*rpc.PaymentStatus, error) {
	status := s.apiClient.GetIncomingPaymentStatus(request.PaymentId)
	return &rpc.PaymentStatus{Status: uint32(status)}, nil
}

func (s *ApiServer) GetOutgoingPaymentStatus(
	context context.Context, request *rpc.PaymentID) (*rpc.PaymentStatus, error) {
	status := s.apiClient.GetOutgoingPaymentStatus(request.PaymentId)
	return &rpc.PaymentStatus{Status: uint32(status)}, nil
}

func (s *ApiServer) ConfirmOnChainResolvedPayments(
	context context.Context, request *rpc.TokenInfo) (*empty.Empty, error) {
	var ercType string
	if request.TokenType == entity.TokenType_ETH {
		ercType = ""
	} else {
		ercType = "ERC20"
	}
	err := s.apiClient.ConfirmOnChainResolvedPays(
		&celersdk.Token{
			Erctype: ercType,
			Addr:    request.TokenAddress,
		})
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) SettleExpiredPayments(
	context context.Context, request *rpc.TokenInfo) (*empty.Empty, error) {
	err := s.apiClient.SettleExpiredPayments(
		&celersdk.TokenInfo{
			TokenType:    celersdk.TokenType(int32(request.TokenType)),
			TokenAddress: request.TokenAddress,
		})
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) IntendWithdraw(
	context context.Context, in *rpc.DepositOrWithdrawRequest) (*empty.Empty, error) {
	tokenInfo := in.TokenInfo
	err := s.apiClient.IntendWithdraw(
		&celersdk.TokenInfo{
			TokenType:    celersdk.TokenType(int32(tokenInfo.TokenType)),
			TokenAddress: tokenInfo.TokenAddress,
		}, in.Amount)
	return &empty.Empty{}, err
}

func (s *ApiServer) ConfirmWithdraw(
	context context.Context, in *rpc.TokenInfo) (*empty.Empty, error) {
	err := s.apiClient.ConfirmWithdraw(
		&celersdk.TokenInfo{
			TokenType:    celersdk.TokenType(int32(in.TokenType)),
			TokenAddress: in.TokenAddress,
		})
	return &empty.Empty{}, err
}

func (s *ApiServer) IntendSettlePaymentChannel(
	context context.Context, request *rpc.TokenInfo) (*empty.Empty, error) {
	err := s.apiClient.IntendSettlePaymentChannel(
		&celersdk.TokenInfo{
			TokenType:    celersdk.TokenType(int32(request.TokenType)),
			TokenAddress: request.TokenAddress,
		})
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) ConfirmSettlePaymentChannel(
	context context.Context, request *rpc.TokenInfo) (*empty.Empty, error) {
	err := s.apiClient.ConfirmSettlePaymentChannel(
		&celersdk.TokenInfo{
			TokenType:    celersdk.TokenType(int32(request.TokenType)),
			TokenAddress: request.TokenAddress,
		})
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) GetSettleFinalizedTimeForPaymentChannel(
	context context.Context, request *rpc.TokenInfo) (*rpc.BlockNumber, error) {
	time, err := s.apiClient.GetSettleFinalizedTimeForPaymentChannel(
		&celersdk.TokenInfo{
			TokenType:    celersdk.TokenType(int32(request.TokenType)),
			TokenAddress: request.TokenAddress,
		})
	if err != nil {
		return nil, err
	}
	return &rpc.BlockNumber{BlockNumber: uint64(time)}, nil
}

// GetPayHistory returns paginated pay history.
// This method is **NOT IDEMPOTENT**.
func (s *ApiServer) GetPayHistory(
	context context.Context, request *rpc.GetPayHistoryRequest) (*rpc.GetPayHistoryResponse, error) {
	if request.GetFromStart() || s.payIter == nil {
		payIter, err := s.apiClient.GetPayHistoryIterator()
		if err != nil {
			return nil, err
		}
		s.payIter = payIter
	}
	itemsPerPage := request.GetItemsPerPage()
	paysJSONStr, err := s.payIter.NextPage(itemsPerPage)
	if err != nil {
		return nil, err
	}

	var pays []*msgrpc.OneHistoricalPay
	if err = json.Unmarshal([]byte(paysJSONStr), &pays); err != nil {
		return nil, err
	}

	resp := &rpc.GetPayHistoryResponse{
		Pays:          pays,
		HasMoreResult: s.payIter.HasMoreResult(),
	}
	return resp, nil
}

func (s *ApiServer) SyncOnChainPaymentChannelStatus(
	context context.Context, request *rpc.TokenInfo) (*empty.Empty, error) {
	var ercType string
	if request.TokenType == entity.TokenType_ETH {
		ercType = ""
	} else {
		ercType = "ERC20"
	}
	err := s.apiClient.SyncOnChainChannelStates(
		&celersdk.Token{
			Erctype: ercType,
			Addr:    request.TokenAddress,
		})
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) SyncStateWithPeer(
	context context.Context, request *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

type appSessionCallback struct {
	seqNumChan chan int
}

func (cb *appSessionCallback) OnDispute(seqNum int) {
	select {
	case cb.seqNumChan <- seqNum:
	default:
	}
}

func (s *ApiServer) CreateAppSessionOnVirtualContract(
	context context.Context,
	request *rpc.CreateAppSessionOnVirtualContractRequest) (
	*rpc.SessionID, error) {
	callback := &appSessionCallback{seqNumChan: make(chan int)}
	session, err :=
		s.apiClient.CreateAppSessionOnVirtualContract(
			request.ContractBin,
			request.ContractConstructor,
			request.Nonce,
			request.OnChainTimeout,
			callback)
	if err != nil {
		return nil, err
	}
	sessionID := session.ID
	s.appSessionMapLock.Lock()
	s.appSessionMap[sessionID] = session
	s.appSessionMapLock.Unlock()
	s.appSessionCallbackMapLock.Lock()
	s.appSessionCallbackMap[sessionID] = callback
	s.appSessionCallbackMapLock.Unlock()
	return &rpc.SessionID{SessionId: sessionID}, nil
}

func (s *ApiServer) CreateAppSessionOnDeployedContract(
	context context.Context,
	request *rpc.CreateAppSessionOnDeployedContractRequest) (
	*rpc.SessionID, error) {
	callback := &appSessionCallback{seqNumChan: make(chan int)}
	participants := strings.Join(request.Participants, ",")
	session, err :=
		s.apiClient.CreateAppSessionOnDeployedContract(
			request.ContractAddress,
			request.Nonce,
			request.OnChainTimeout,
			participants,
			callback)
	if err != nil {
		return nil, err
	}
	sessionID := session.ID
	s.appSessionMapLock.Lock()
	s.appSessionMap[sessionID] = session
	s.appSessionMapLock.Unlock()
	s.appSessionCallbackMapLock.Lock()
	s.appSessionCallbackMap[sessionID] = callback
	s.appSessionCallbackMapLock.Unlock()
	return &rpc.SessionID{SessionId: sessionID}, nil
}

func (s *ApiServer) SubscribeAppSessionDispute(
	request *rpc.SessionID, stream rpc.WebApi_SubscribeAppSessionDisputeServer) error {
	sessionID := request.SessionId
	s.appSessionCallbackMapLock.Lock()
	callback := s.appSessionCallbackMap[sessionID]
	s.appSessionCallbackMapLock.Unlock()
	for {
		seqNum := <-callback.seqNumChan
		err := stream.Send(&rpc.DisputeInfo{
			SessionId: sessionID,
			SeqNum:    uint64(seqNum),
		})
		if err != nil {
			return err
		}
	}
}

func (s *ApiServer) SignOutgoingState(
	context context.Context, request *rpc.SignOutgoingStateRequest) (*rpc.SignedState, error) {
	session := s.getAppSession(request.SessionId)
	signed, err := session.SignAppData(request.State)
	if err != nil {
		return nil, err
	}
	return &rpc.SignedState{SignedState: signed}, nil
}

func (s *ApiServer) ValidateAck(
	context context.Context, request *rpc.ValidateAckRequest) (*rpc.BoolValue, error) {
	session := s.getAppSession(request.SessionId)
	_, err := session.HandleMatchData(celersdk.OPCODE_ACK, request.Envelope)
	valid := true
	if err != nil {
		valid = false
	}
	return &rpc.BoolValue{Value: valid}, nil
}

func (s *ApiServer) ProcessReceivedState(
	context context.Context,
	request *rpc.ProcessReceivedStateRequest) (*rpc.ProcessReceivedStateResponse, error) {
	session := s.getAppSession(request.SessionId)
	appData, err := session.HandleMatchData(celersdk.OPCODE_NEWSTATE, request.Envelope)
	if err != nil {
		return nil, err
	}
	return &rpc.ProcessReceivedStateResponse{
		DecodedState: appData.Received, PreparedAck: appData.AckMsg}, nil
}

func (s *ApiServer) SignData(context context.Context, request *rpc.Data) (*rpc.Signature, error) {
	signature, err := s.apiClient.SignData(request.Data)
	if err != nil {
		return nil, err
	}
	return &rpc.Signature{Signature: signature}, nil
}

func (s *ApiServer) SettleAppSession(
	context context.Context,
	request *rpc.SettleAppSessionRequest) (*empty.Empty, error) {
	session := s.getAppSession(request.SessionId)
	err := session.SwitchToOnchain(request.StateProof)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) SettleAppSessionBySigTimeout(
	context context.Context,
	request *rpc.SettleAppSessionByTimeoutRequest) (*empty.Empty, error) {
	session := s.getAppSession(request.SessionId)
	err := session.SettleBySigTimeout(request.OracleProof)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) SettleAppSessionByMoveTimeout(
	context context.Context,
	request *rpc.SettleAppSessionByTimeoutRequest) (*empty.Empty, error) {
	session := s.getAppSession(request.SessionId)
	err := session.SettleByMoveTimeout(request.OracleProof)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) SettleAppSessionByInvalidTurn(
	context context.Context,
	request *rpc.SettleAppSessionByInvalidityRequest) (*empty.Empty, error) {
	session := s.getAppSession(request.SessionId)
	err := session.SettleByInvalidTurn(request.OracleProof, request.CosignedStateProof)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) SettleAppSessionByInvalidState(
	context context.Context,
	request *rpc.SettleAppSessionByInvalidityRequest) (*empty.Empty, error) {
	session := s.getAppSession(request.SessionId)
	err := session.SettleByInvalidState(request.OracleProof, request.CosignedStateProof)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) DeleteAppSession(
	context context.Context, request *rpc.SessionID) (*empty.Empty, error) {
	sessionID := request.SessionId
	err := s.apiClient.EndAppSession(sessionID)
	if err != nil {
		return nil, err
	}
	s.appSessionMapLock.Lock()
	s.appSessionMap[sessionID] = nil
	s.appSessionMapLock.Unlock()
	return &empty.Empty{}, nil
}

func (s *ApiServer) GetDeployedAddressForAppSession(
	context context.Context, request *rpc.SessionID) (*rpc.Address, error) {
	session := s.getAppSession(request.SessionId)
	address, err := session.GetDeployedAddress()
	if err != nil {
		return nil, err
	}
	return &rpc.Address{Address: address}, nil
}

func (s *ApiServer) GetBooleanOutcomeForAppSession(
	context context.Context,
	request *rpc.GetBooleanOutcomeForAppSessionRequest) (*rpc.BooleanOutcome, error) {
	session := s.getAppSession(request.SessionId)
	res, err := session.OnChainGetBooleanOutcome(request.Query)
	if err != nil {
		return nil, err
	}
	return &rpc.BooleanOutcome{Finalized: res.Finalized, Outcome: res.Outcome}, nil
}

func (s *ApiServer) ApplyActionForAppSession(
	context context.Context, request *rpc.ApplyActionForAppSessionRequest) (*empty.Empty, error) {
	session := s.getAppSession(request.SessionId)
	err := session.OnChainApplyAction(request.Action)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) FinalizeOnActionTimeoutForAppSession(
	context context.Context, request *rpc.SessionID) (*empty.Empty, error) {
	session := s.getAppSession(request.SessionId)
	err := session.OnChainFinalizeOnActionTimeout()
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ApiServer) GetSettleFinalizedTimeForAppSession(
	context context.Context, request *rpc.SessionID) (*rpc.BlockNumber, error) {
	session := s.getAppSession(request.SessionId)
	time, err := session.OnChainGetSettleFinalizedTime()
	if err != nil {
		return nil, err
	}
	return &rpc.BlockNumber{BlockNumber: uint64(time)}, nil
}

func (s *ApiServer) GetActionDeadlineForAppSession(
	context context.Context, request *rpc.SessionID) (*rpc.BlockNumber, error) {
	session := s.getAppSession(request.SessionId)
	deadline, err := session.OnChainGetActionDeadline()
	if err != nil {
		return nil, err
	}
	return &rpc.BlockNumber{BlockNumber: uint64(deadline)}, err
}

func (s *ApiServer) GetStatusForAppSession(
	context context.Context, request *rpc.SessionID) (*rpc.AppSessionStatus, error) {
	session := s.getAppSession(request.SessionId)
	status, err := session.OnChainGetStatus()
	if err != nil {
		return nil, err
	}
	return &rpc.AppSessionStatus{Status: uint32(status)}, err
}

func (s *ApiServer) GetStateForAppSession(
	context context.Context,
	request *rpc.GetStateForAppSessionRequest) (*rpc.AppSessionState, error) {
	session := s.getAppSession(request.SessionId)
	state, err := session.OnChainGetState(request.Key)
	if err != nil {
		return nil, err
	}
	return &rpc.AppSessionState{State: state}, nil
}

func (s *ApiServer) GetSeqNumForAppSession(
	context context.Context, request *rpc.SessionID) (*rpc.AppSessionSeqNum, error) {
	session := s.getAppSession(request.SessionId)
	seqNum, err := session.OnChainGetSeqNum()
	if err != nil {
		return nil, err
	}
	return &rpc.AppSessionSeqNum{SeqNum: uint64(seqNum)}, err
}

func (s *ApiServer) GetBlockNumber(
	context context.Context, request *empty.Empty) (*rpc.BlockNumber, error) {
	return &rpc.BlockNumber{BlockNumber: uint64(s.apiClient.GetCurrentBlockNumber())}, nil
}

func (s *ApiServer) SetMsgDropper(context context.Context, req *rpc.SetMsgDropReq) (*empty.Empty, error) {
	s.apiClient.SetMsgDropper(req.DropRecv, req.DropSend)
	return new(empty.Empty), nil
}

func (s *ApiServer) getAppSession(sessionID string) *celersdk.AppSession {
	s.appSessionMapLock.Lock()
	session := s.appSessionMap[sessionID]
	s.appSessionMapLock.Unlock()
	return session
}

func stripPort(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}
	if i := strings.IndexByte(hostport, ']'); i != -1 {
		return strings.TrimPrefix(hostport[:i], "[")
	}
	return hostport[:colon]
}
