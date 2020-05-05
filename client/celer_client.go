// Copyright 2018-2020 Celer Network

package client

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/celersdkintf"
	"github.com/celer-network/goCeler/cnode"
	"github.com/celer-network/goCeler/cnode/cooperativewithdraw"
	"github.com/celer-network/goCeler/common"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/deposit"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/route"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"golang.org/x/net/context"
)

// subset of celersdk async.go ClientCallback so we can call app callback
type clientCallbackAdapter interface {
	HandleChannelOpened(token, cid string)
	HandleOpenChannelError(token, reason string)
	HandleRecvStart(pay *celersdkintf.Payment)
	HandleRecvDone(pay *celersdkintf.Payment)
	HandleSendComplete(pay *celersdkintf.Payment)
	HandleSendErr(pay *celersdkintf.Payment, e *celersdkintf.E)
}

// CelerClient implements main functionalities
type CelerClient struct {
	cNode         *cnode.CNode
	sc            common.StateCallback // callback for java
	svr           string               // OSP gRPC endpoint
	svrEth        ctype.Addr           // OSP ETH address
	dal           *storage.DAL         // database
	onClientEvent clientCallbackAdapter
}

func condPayToPayment(
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	payNote *any.Any,
	status int) *celersdkintf.Payment {
	// doulbe check here, caller should avoid to pass nil pay
	if pay == nil {
		return &celersdkintf.Payment{}
	}
	payJSON, _ := utils.PbToJSONString(pay)
	payNoteJSON, _ := utils.PbToJSONString(payNote)
	// ignore the error since if there is an error, it would return empty string
	payNoteType, _ := ptypes.AnyMessageName(payNote)
	maxTransfer := pay.TransferFunc.MaxTransfer
	payTimestamp := int64(pay.PayTimestamp / uint64(time.Millisecond))

	p := &celersdkintf.Payment{
		Sender:   ctype.Bytes2Hex(pay.GetSrc()),
		Receiver: ctype.Bytes2Hex(pay.GetDest()),
		// return decimal amt.
		AmtWei:       new(big.Int).SetBytes(maxTransfer.Receiver.Amt).String(),
		TokenAddr:    ctype.Bytes2Hex(maxTransfer.Token.TokenAddress),
		UID:          ctype.PayID2Hex(payID),
		PayJSON:      payJSON,
		Status:       status,
		PayNoteType:  payNoteType,
		PayNoteJSON:  payNoteJSON,
		PayTimestamp: payTimestamp,
	}
	if maxTransfer.Token.TokenType == entity.TokenType_ETH {
		p.TokenAddr = ""
	}
	return p
}

func (c *CelerClient) HandleReceivingStart(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any) {
	if c.onClientEvent != nil {
		c.onClientEvent.HandleRecvStart(condPayToPayment(payID, pay, note, celersdkintf.PAY_STATUS_PENDING))
	}
}

func settleReasonToPayStatus(reason rpc.PaymentSettleReason) int {
	var status int
	switch reason {
	case rpc.PaymentSettleReason_PAY_EXPIRED:
		status = celersdkintf.PAY_STATUS_UNPAID_EXPIRED
	case rpc.PaymentSettleReason_PAY_REJECTED:
		status = celersdkintf.PAY_STATUS_UNPAID_REJECTED
	case rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN:
		status = celersdkintf.PAY_STATUS_PAID_RESOLVED_ONCHAIN
	case rpc.PaymentSettleReason_PAY_PAID_MAX:
		status = celersdkintf.PAY_STATUS_PAID
	case rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE:
		status = celersdkintf.PAY_STATUS_UNPAID_DEST_UNREACHABLE
	default:
		status = celersdkintf.PAY_STATUS_INVALID
	}
	return status
}

func (c *CelerClient) HandleReceivingDone(
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	note *any.Any,
	reason rpc.PaymentSettleReason) {
	if c.onClientEvent != nil {
		status := settleReasonToPayStatus(reason)
		c.onClientEvent.HandleRecvDone(condPayToPayment(payID, pay, note, status))
	}
}

func (r *CelerClient) HandleSendComplete(
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	note *any.Any,
	reason rpc.PaymentSettleReason) {
	if r.onClientEvent != nil {
		status := settleReasonToPayStatus(reason)
		r.onClientEvent.HandleSendComplete(condPayToPayment(payID, pay, note, status))
	}
}

func (r *CelerClient) HandleDestinationUnreachable(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any) {
	if r.onClientEvent != nil {
		r.onClientEvent.HandleSendErr(
			condPayToPayment(payID, pay, note, celersdkintf.PAY_STATUS_UNPAID_DEST_UNREACHABLE),
			&celersdkintf.E{Reason: "Unreachable to " + ctype.Bytes2Hex(pay.GetDest()), Code: -1})
	}
}

func (r *CelerClient) HandleSendFail(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any, errMsg string) {
	if r.onClientEvent != nil {
		r.onClientEvent.HandleSendErr(
			condPayToPayment(payID, pay, note, celersdkintf.PAY_STATUS_UNPAID),
			&celersdkintf.E{Reason: errMsg + ": Send token failed to " + ctype.Bytes2Hex(pay.GetDest()), Code: -1})
	}
}

func NewCelerClient(
	keyStore string, passPhrase string, profile common.CProfile, clientCallback clientCallbackAdapter) (*CelerClient, error) {
	c := &CelerClient{}
	var err error
	masterTxConfig := transactor.NewTransactorConfig(keyStore, passPhrase)
	c.cNode, err = cnode.NewCNode(
		masterTxConfig,
		nil, /* depositTxConfig */
		nil, /* transactorConfigs */
		profile,
		route.GateWayPolicy,
		nil /* routingData */)
	if err != nil {
		return nil, fmt.Errorf("cNode init failed: %w", err)
	}
	c.initialize(profile, clientCallback)
	log.Infoln("Finishing NewCelerClient")
	return c, nil
}

func NewCelerClientWithExternalSigner(
	address ctype.Addr, signer common.Signer, profile common.CProfile,
	clientCallback clientCallbackAdapter) (*CelerClient, error) {
	c := &CelerClient{}
	var err error
	c.cNode, err = cnode.NewCNodeWithExternalSigner(address, signer, profile)
	if err != nil {
		return nil, fmt.Errorf("cNode init failed: %w", err)
	}
	c.initialize(profile, clientCallback)
	log.Infoln("Finishing NewCelerClient with external signer", ctype.Addr2Hex(address))
	return c, nil
}

func (c *CelerClient) initialize(profile common.CProfile, clientCallback clientCallbackAdapter) {
	c.onClientEvent = clientCallback
	c.dal = c.cNode.GetDAL()
	c.svr = profile.SvrRPC
	c.svrEth = ctype.Hex2Addr(profile.SvrETHAddr)
	c.cNode.OnReceivingToken(c)
	c.cNode.OnSendToken(c)
}

// Close tries to close db and networking then set c.cNode to nil
// so all future code to access c.cNode.xxx will panic
// TODO: a cleaner solution is to have a close only (ie no data) signal chan
// all components must honor and exit cleanly
func (c *CelerClient) Close() {
	if c.cNode != nil {
		c.cNode.Close()
		c.cNode = nil
	}
}

// ClearCallbacks set c.onClientEvent to nil so no more callbacks will be
// triggered. This has to be a different func instead of within Close is in
// Init if client failed, we call close but still want the init failed callback
// to trigger so mobile knows. But in celersdk Destroy API, we want to cleanup
// everything
func (c *CelerClient) ClearCallbacks() {
	// this is unprotected and is bad.
	// we decide to do this with assumption that by the time ClearCallbacks is called,
	// all other stuff should have been closed properly so no code will try to trigger
	// callback
	c.onClientEvent = nil
}

// Return the FeeMgr's Eth address.
//
// Note: for now this returns the Eth address of the OSP server that the
// client is connected to.  In the future when multiple OSP Eth addresses
// could be used, and assuming a single FeeMgr across all OSPs, then this
// function should be changed to return the FeeMgr's Eth address which
// would be configured separately in the profile.
func (c *CelerClient) GetFeeMgrEth() string {
	return ctype.Addr2Hex(c.svrEth)
}

// RegisterStream establishes gRPC streaming connections with OSP
func (c *CelerClient) RegisterStream() error {
	err := c.cNode.RegisterStream(c.svrEth, c.svr)
	if err != nil {
		log.Errorln("RegisterStream failed:", c.svrEth.Hex(), c.svr, err)
		return fmt.Errorf("RegisterStream failed: %w", err)
	}

	// Register the callback to handle stream errors and try to reconnect.
	// TODO: note that such a two-step API has a tiny race-condition window
	// in case the stream quickly disconnects after a successful call to
	// RegisterStream() and before RegisterStreamErrCallback() is done.
	// The next design should either allow them both to be done atomically
	// or allow RegisterStreamErrCallback() before RegisterStream().
	streamRetryCb := func(addr ctype.Addr, streamErr error) {
		log.Infoln("streamRetryCb triggered for", addr.Hex(), streamErr)
		delay := time.Second * 10
		maxdelay := time.Minute

		for {
			log.Debugln("streamRetryCb: try to register again", addr.Hex())
			err := c.cNode.RegisterStream(c.svrEth, c.svr)
			if err == nil {
				break
			}
			log.Errorln("streamRetryCb: register failed", addr.Hex(), err)
			time.Sleep(delay)
			delay += time.Second * 10
			if delay > maxdelay {
				delay = maxdelay
			}
		}

		log.Infoln("streamRetry:Cb successful re-register", addr.Hex())
	}

	c.cNode.RegisterStreamErrCallback(c.svrEth, streamRetryCb)
	return nil
}

// IntendSettlePaymentChannel starts payment channel settling process
func (c *CelerClient) IntendSettlePaymentChannel(token *entity.TokenInfo) error {
	cid, exist := c.getCidFromTokenInfo(token)
	if !exist {
		return errors.New("PSC_NOT_OPEN_" + utils.GetTokenAddrStr(token))
	}
	err := c.cNode.IntendSettlePaymentChannel(cid)
	return err
}

// ConfirmSettlePaymentChannel confirm settle and close a payment channel on-chain
func (c *CelerClient) ConfirmSettlePaymentChannel(token *entity.TokenInfo) error {
	cid, exist := c.getCidFromTokenInfo(token)
	if !exist {
		return errors.New("PSC_NOT_OPEN_" + utils.GetTokenAddrStr(token))
	}
	return c.cNode.ConfirmSettlePaymentChannel(cid)
}

func (c *CelerClient) GetSettleFinalizedTime(token *entity.TokenInfo) (*big.Int, error) {
	cid, exist := c.getCidFromTokenInfo(token)
	if !exist {
		return big.NewInt(0), errors.New("PSC_NOT_OPEN_" + utils.GetTokenAddrStr(token))
	}
	return c.cNode.GetSettleFinalizedTime(cid)
}

func (c *CelerClient) IntendWithdraw(token *entity.TokenInfo, amount *big.Int) error {
	cid, exist := c.getCidFromTokenInfo(token)
	if !exist {
		return errors.New("PSC_NOT_OPEN_" + utils.GetTokenAddrStr(token))
	}
	err := c.cNode.IntendWithdraw(cid, amount, ctype.ZeroCid)
	return err
}

func (c *CelerClient) ConfirmWithdraw(token *entity.TokenInfo) error {
	cid, exist := c.getCidFromTokenInfo(token)
	if !exist {
		return errors.New("PSC_NOT_OPEN_" + utils.GetTokenAddrStr(token))
	}
	err := c.cNode.ConfirmWithdraw(cid)
	return err
}

func (c *CelerClient) VetoWithdraw(token *entity.TokenInfo) error {
	cid, exist := c.getCidFromTokenInfo(token)
	if !exist {
		return errors.New("PSC_NOT_OPEN_" + utils.GetTokenAddrStr(token))
	}
	err := c.cNode.VetoWithdraw(cid)
	return err
}

func (c *CelerClient) SetDelegation(tokens []*entity.TokenInfo, timeout int64) error {
	tks := make([]ctype.Addr, 0, len(tokens))
	for _, tk := range tokens {
		tks = append(tks, ctype.Bytes2Addr(tk.GetTokenAddress()))
	}
	return c.cNode.SetDelegation(tks, timeout)
}

func (c *CelerClient) Deposit(
	tokenAddr ctype.Addr, amt *big.Int, cb deposit.DepositCallback) (string, error) {
	cid, exist := c.getCidFromToken(tokenAddr)
	if !exist {
		return "", errors.New("PSC_NOT_OPEN_" + ctype.Addr2Hex(tokenAddr))
	}
	return c.cNode.DepositWithCallback(amt, cid, cb)
}

func (c *CelerClient) MonitorDepositJob(jobID string, cb deposit.DepositCallback) {
	c.cNode.MonitorDepositJobWithCallback(jobID, cb)
}

func (c *CelerClient) RemoveDepositJob(jobID string) {
	c.cNode.RemoveDepositJob(jobID)
}

func (c *CelerClient) GetBalance(cid ctype.CidType) (*common.ChannelBalance, error) {
	return c.cNode.GetBalance(cid)
}

func (c *CelerClient) GetChannelState(tkAddr ctype.Addr) string {
	cid, exist := c.getCidFromToken(tkAddr)
	if !exist {
		return "NOT_FOUND"
	}
	state, found, err := c.dal.GetChanState(cid)
	if err != nil {
		return err.Error()
	}
	if !found {
		return common.ErrChannelNotFound.Error()
	}
	return fsm.ChanStateName(state)
}

func (c *CelerClient) HasPendingOpenChanRequest(tk *entity.TokenInfo) bool {
	blk, found, err := c.dal.GetDestTokenOpenChanBlkNum(c.svrEth, tk)
	if err != nil || !found {
		// not found, never requested, no pending
		return false
	}
	curBlk := c.GetCurrentBlockNumberUint64()
	if curBlk <= blk+uint64(config.OpenChannelTimeout) {
		return true
	}
	return false
}

// GetTokenBalance returns ERC20 avaialble, locked and maxReceiving capacity
// Note that the returned balance is based on local knowledge
func (c *CelerClient) GetTokenBalance(tokenAddr ctype.Addr) (*big.Int, *big.Int, *big.Int, error) {
	cid, exist := c.getCidFromToken(tokenAddr)
	if !exist {
		return nil, nil, nil, errors.New("PSC_NOT_OPEN_" + ctype.Addr2Hex(tokenAddr))
	}
	balance, err := c.GetBalance(cid)
	if err != nil {
		return nil, nil, nil, err
	}
	return balance.MyFree, balance.MyLocked, balance.PeerFree, nil
}

// IsConnectedToCeler checks if the given peer has connected to Celer
// and returns its join status(LOCAL or REMOTE) and its free balance as a decimal string value.
// If the peer has not joined Celer, it returns join status(NOT JOIN) and an empty string
func (c *CelerClient) IsConnectedToCeler(tokenaddr string, addr string) (rpc.JoinCelerStatus, string, error) {
	conn, err := c.cNode.GetConnManager().GetClient(c.svrEth)
	if err != nil {
		return rpc.JoinCelerStatus_NOT_JOIN, "", fmt.Errorf("fail to get peer status: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	response, err := conn.CelerGetPeerStatus(
		ctx,
		&rpc.PeerAddress{
			Address:   addr,
			TokenAddr: tokenaddr,
		},
	)
	if err != nil {
		return rpc.JoinCelerStatus_NOT_JOIN, "", fmt.Errorf("fail to get peer status: %w", err)
	}
	return response.GetJoinStatus(), response.GetFreeBalance(), nil
}

// GetCurrentBlockNumber returns the current block number
func (c *CelerClient) GetCurrentBlockNumber() *big.Int {
	return c.cNode.GetCurrentBlockNumber()
}

// GetCurrentBlockNumber returns the current block number
func (c *CelerClient) GetCurrentBlockNumberUint64() uint64 {
	return c.cNode.GetCurrentBlockNumber().Uint64()
}

func (c *CelerClient) CooperativeWithdraw(
	tokenAddr ctype.Addr, amount *big.Int, callback cooperativewithdraw.Callback) (string, error) {
	cid, exist := c.getCidFromToken(tokenAddr)
	if !exist {
		return "", errors.New("PSC_NOT_OPEN_" + ctype.Addr2Hex(tokenAddr))
	}
	return c.cNode.CooperativeWithdraw(cid, amount, callback)
}

func (c *CelerClient) MonitorCooperativeWithdrawJob(
	withdrawHash string, callback cooperativewithdraw.Callback) {
	c.cNode.MonitorCooperativeWithdrawJob(withdrawHash, callback)
}

func (c *CelerClient) RemoveCooperativeWithdrawJob(withdrawHash string) {
	c.cNode.RemoveCooperativeWithdrawJob(withdrawHash)
}

// GetDAL is simple helper so sdk layer can call GetSerializedDatabase
func (c *CelerClient) GetDAL() *storage.DAL {
	return c.dal
}

func (c *CelerClient) SetMsgDropper(dropRecv, dropSend bool) {
	c.cNode.SetMsgDropper(dropRecv, dropSend)
}

// get incoming payment sdk level status
func (c *CelerClient) GetIncomingPaymentStatus(payID ctype.PayIDType) int {
	inState, _ := c.cNode.GetPaymentState(payID)
	return payFsmStateToSdkStatus(inState)
}

// get outgoing payment sdk level status
func (c *CelerClient) GetOutgoingPaymentStatus(payID ctype.PayIDType) int {
	_, outState := c.cNode.GetPaymentState(payID)
	return payFsmStateToSdkStatus(outState)
}

// GetPayment returns the related payment info of a specified payment ID
func (c *CelerClient) GetPayment(payID ctype.PayIDType) (*celersdkintf.Payment, error) {
	pay, note, _, inState, _, outState, _, found, err := c.dal.GetPaymentInfo(payID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, common.ErrPayNotFound
	}

	state := inState
	if inState == enums.PayState_NULL {
		state = outState
	}

	sdkState := payFsmStateToSdkStatus(state)

	return condPayToPayment(payID, pay, note, sdkState), nil
}

// GetAllPayments returns all payments info
func (c *CelerClient) GetAllPayments() ([]*celersdkintf.Payment, error) {
	allPayIDs, err := c.dal.GetAllPayIDs()
	if err != nil {
		return nil, err
	}

	var allpays []*celersdkintf.Payment
	for _, payID := range allPayIDs {
		pay, err := c.GetPayment(payID)
		if err != nil {
			continue
		}
		allpays = append(allpays, pay)
	}

	return allpays, nil
}

func payFsmStateToSdkStatus(state int) int {
	switch state {
	case enums.PayState_ONESIG_PENDING, enums.PayState_COSIGNED_PENDING:
		return celersdkintf.PAY_STATUS_INITIALIZING
	case enums.PayState_SECRET_REVEALED, enums.PayState_ONESIG_PAID, enums.PayState_ONESIG_CANCELED, enums.PayState_INGRESS_REJECTED:
		return celersdkintf.PAY_STATUS_PENDING
	case enums.PayState_COSIGNED_PAID:
		return celersdkintf.PAY_STATUS_PAID
	case enums.PayState_COSIGNED_CANCELED, enums.PayState_NACKED:
		return celersdkintf.PAY_STATUS_UNPAID
	default:
		return celersdkintf.PAY_STATUS_INVALID
	}
}
