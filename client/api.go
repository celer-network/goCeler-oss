// Copyright 2019-2020 Celer Network

package client

import (
	"errors"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/ptypes/any"
)

// mycb implements event.OpenChannel to receive callback from cnode
type mycb struct {
	tokenAddr ctype.Addr
	appcb     clientCallbackAdapter
	dal       *storage.DAL // only used in onchain openchannel callback
	svrEth    ctype.Addr
}

func (cb *mycb) HandleOpenChannelFinish(cid ctype.CidType) {
	tokenAddr := cb.tokenAddr
	log.Infoln("Opened channel for tokenAddr", tokenAddr.Hex(), "cid", cid.Hex())
	if cb.dal != nil {
		cb.setBlkNumTo0()
	}
	if cb.appcb != nil {
		go cb.appcb.HandleChannelOpened(ctype.Addr2Hex(cb.tokenAddr), ctype.Cid2Hex(cid))
	}
}

func (cb *mycb) HandleOpenChannelErr(e *common.E) {
	log.Error("Openchannel err:", *e)
	if cb.dal != nil {
		cb.setBlkNumTo0()
	}
	if cb.appcb != nil {
		go cb.appcb.HandleOpenChannelError(ctype.Addr2Hex(cb.tokenAddr), e.Reason)
	}
}

// set lastOpenChanReqBlkNum to 0
func (cb *mycb) setBlkNumTo0() {
	err := cb.dal.UpsertDestTokenOpenChanBlkNum(cb.svrEth, utils.GetTokenInfoFromAddress(cb.tokenAddr), 0)
	if err != nil {
		log.Warnln("setBlkNumTo0 err:", err)
	}
}

// GetRpcClientToOsp returns rpc client to osp.
func (c *CelerClient) GetRpcClientToOsp() (rpc.RpcClient, error) {
	return c.cNode.GetConnManager().GetClient(c.svrEth)
}

// GetMyEthAddr returns eth addr of this client.
func (c *CelerClient) GetMyEthAddr() ctype.Addr {
	return c.cNode.EthAddress
}

// TODO: if we ever want to enforce only one openchannel request is pending, we can trylock and
// unlock in cb
func (c *CelerClient) OpenChannel(
	token *entity.TokenInfo, myAmt, peerAmt *big.Int, appcb clientCallbackAdapter) error {
	cid, ok := c.getCidFromTokenInfo(token)
	if ok { // already opened channel for this token
		if appcb != nil {
			go appcb.HandleChannelOpened(ctype.Bytes2Hex(token.TokenAddress), cid.Hex())
		}
		return nil
	}
	err := c.dal.UpsertDestTokenOpenChanBlkNum(c.svrEth, token, c.GetCurrentBlockNumberUint64())
	if err != nil {
		log.Warnln("OpenChannel: cannot save block number:", err)
	}
	return c.cNode.OpenChannel(
		c.svrEth,
		myAmt,
		peerAmt,
		token,
		false, /*ospToOspOpen*/
		&mycb{
			tokenAddr: utils.GetTokenAddr(token),
			appcb:     appcb,
			dal:       c.dal,
			svrEth:    c.svrEth,
		})
}
func (c *CelerClient) TcbOpenChannel(
	token *entity.TokenInfo, peerAmt *big.Int, appcb clientCallbackAdapter) error {
	cid, ok := c.getCidFromTokenInfo(token)
	if ok { // already opened channel for this token
		if appcb != nil {
			go appcb.HandleChannelOpened(ctype.Addr2Hex(utils.GetTokenAddr(token)), cid.Hex())
		}
		return nil
	}
	return c.cNode.TcbOpenChannel(
		c.svrEth,
		peerAmt,
		token,
		&mycb{
			tokenAddr: utils.GetTokenAddr(token),
			appcb:     appcb,
		})
}
func (c *CelerClient) InstantiateChannelForToken(token *entity.TokenInfo, appcb clientCallbackAdapter) error {
	cid, ok := c.getCidFromTokenInfo(token)
	if !ok {
		return common.ErrNoChannel
	}
	return c.cNode.InstantiateChannel(cid, &mycb{
		tokenAddr: ctype.Bytes2Addr(token.TokenAddress),
		appcb:     appcb,
	})
}

// AddBooleanPay creates a condpay based on args, and call cnode to send CondPayRequest
// returns payId or err
func (c *CelerClient) AddBooleanPay(
	xfer *entity.TokenTransfer, conds []*entity.Condition, resolveDeadline uint64, note *any.Any, dstNetId uint64) (ctype.PayIDType, error) {

	if xfer == nil || xfer.Receiver == nil || xfer.Receiver.Account == nil {
		return ctype.ZeroPayID, common.ErrInvalidArg
	}
	if resolveDeadline <= c.GetCurrentBlockNumber().Uint64() {
		return ctype.ZeroPayID, common.ErrDeadlinePassed
	}

	// Create a new condpay object
	pay := &entity.ConditionalPay{
		Src:        c.cNode.EthAddress.Bytes(),
		Dest:       xfer.Receiver.Account,
		Conditions: conds,
		TransferFunc: &entity.TransferFunction{
			LogicType:   entity.TransferFunctionType_BOOLEAN_AND,
			MaxTransfer: xfer,
		},
		ResolveDeadline: resolveDeadline,
		ResolveTimeout:  config.PayResolveTimeout,
	}

	var payID ctype.PayIDType
	var cnoderr error
	for i := 0; i < 10; i++ {
		payID, cnoderr = c.cNode.AddBooleanPay(pay, note, dstNetId)
		if cnoderr != common.ErrPendingSimplex {
			break
		}
		log.Warn("pending simplexstate, retry: ", i)
		time.Sleep(200 * time.Millisecond)
	}
	return payID, cnoderr
}

func (c *CelerClient) ConfirmBooleanPay(payID ctype.PayIDType) error {
	return c.cNode.ConfirmBooleanPay(payID)
}

func (c *CelerClient) RejectBooleanPay(payID ctype.PayIDType) error {
	return c.cNode.RejectBooleanPay(payID)
}

func (c *CelerClient) SettleOnChainResolvedPay(payID ctype.PayIDType) error {
	return c.cNode.SettleOnChainResolvedPay(payID)
}

func (c *CelerClient) ConfirmOnChainResolvedPays(token *entity.TokenInfo) error {
	cid, exist := c.getCidFromTokenInfo(token)
	if !exist {
		return errors.New("PSC_NOT_OPEN_" + utils.GetTokenAddrStr(token))
	}
	return c.cNode.ConfirmOnChainResolvedPays(cid)
}

func (c *CelerClient) SettleExpiredPays(token *entity.TokenInfo) error {
	cid, exist := c.getCidFromTokenInfo(token)
	if !exist {
		return errors.New("PSC_NOT_OPEN_" + utils.GetTokenAddrStr(token))
	}
	return c.cNode.SettleExpiredPays(cid)
}

func (c *CelerClient) OnReceivingToken(callback event.OnReceivingTokenCallback) {
	c.cNode.OnReceivingToken(callback)
}
func (c *CelerClient) OnSendingToken(callback event.OnSendingTokenCallback) {
	c.cNode.OnSendToken(callback)
}

// legacy helper, should be deprecated after all args are typed
func (c *CelerClient) getCidFromToken(tokenAddr ctype.Addr) (ctype.CidType, bool) {
	tk := utils.GetTokenInfoFromAddress(tokenAddr)
	cid, exist := c.getCidFromTokenInfo(tk)
	return cid, exist
}

// getCidFromTokenInfo reads peer:token->cid from database (peer is svrEth)
// return channel id and whether it exists. note cid 0 is invalid
func (c *CelerClient) getCidFromTokenInfo(token *entity.TokenInfo) (ctype.CidType, bool) {
	cid, found, err := c.dal.GetCidByPeerToken(c.svrEth, token)
	if err != nil {
		log.Error(err, utils.GetTokenAddrStr(token))
	}
	return cid, found
}

func (c *CelerClient) SignState(in []byte) []byte {
	return c.cNode.SignState(in)
}

// ResolveCondPayOnChain tries to resolve a payment onchain in the PayRegistry
func (c *CelerClient) ResolveCondPayOnChain(payID ctype.PayIDType) error {
	return c.cNode.Disputer.SettleConditionalPay(payID)
}

func (c *CelerClient) GetCondPayInfoFromRegistry(payID ctype.PayIDType) (*big.Int, uint64, error) {
	return c.cNode.Disputer.GetCondPayInfoFromRegistry(payID)
}

func (c *CelerClient) SyncOnChainChannelStates(token *entity.TokenInfo) error {
	cid, exist := c.getCidFromTokenInfo(token)
	if !exist {
		return errors.New("PSC_NOT_OPEN_" + utils.GetTokenAddrStr(token))
	}
	_, err := c.cNode.SyncOnChainChannelStates(cid)
	if err != nil {
		return err
	}
	return nil
}

// MyAddress returns ctype.Addr from c.cNode.EthAddress
func (c *CelerClient) MyAddress() ctype.Addr {
	return c.cNode.EthAddress
}
