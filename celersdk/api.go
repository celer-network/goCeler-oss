// Copyright 2018-2020 Celer Network

// interface for celer sdk

package celersdk

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/celersdkintf"
	"github.com/celer-network/goCeler/client"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

// Celer mobile client. must define before methods!
// gobind has a bug to clear method doc if struct is defined after methods
type Client struct {
	c       *client.CelerClient
	datadir string // dataPath when InitClient
}

// InitClient creates a celer client
// cfg is the content of profile json file.
// dataPath is the dir that holds celer data.
func InitClient(acnt *Account, cfg, dataPath string, cb ClientCallback) {
	go func() {
		p := common.Bytes2Profile([]byte(cfg))
		p.StoreDir = dataPath
		c, err := newClient(acnt.Keystore, acnt.Password, p, cb)
		if err != nil {
			cb.HandleClientInitErr(&celersdkintf.E{Reason: err.Error(), Code: -1})
		} else {
			cb.HandleClientReady(c)
		}
	}()
}

// InitClientWithSigner creates celer client with external signer
// addr is hex string of ETH address eg. 0x1234...
func InitClientWithSigner(addr, cfg, dataPath string, cb ClientCallback, signcb ExternalSignerCallback) {
	myaddr := ctype.Hex2Addr(addr)
	if myaddr == ctype.ZeroAddr {
		log.Errorln("invalid addr:", addr)
		cb.HandleClientInitErr(&celersdkintf.E{Reason: "invalid addr: " + addr, Code: -1})
		return
	}
	go func() {
		p := common.Bytes2Profile([]byte(cfg))
		p.StoreDir = dataPath
		cc, err := client.NewCelerClientWithExternalSigner(myaddr, newExtSignerMgr(signcb), *p, cb)
		if err != nil {
			cb.HandleClientInitErr(&celersdkintf.E{Reason: err.Error(), Code: -1})
			return
		}
		// Note RegisterStream will trigger sign so ext signer will get callback before client ready callback
		err = cc.RegisterStream()
		if err != nil {
			cc.Close()
			cb.HandleClientInitErr(&celersdkintf.E{Reason: err.Error(), Code: -1})
			return
		}
		cb.HandleClientReady(&Client{
			c:       cc,
			datadir: dataPath,
		})
	}()
}

// GetDataDir returns dataPath when InitClient, so cxc can share same folder
func (mc *Client) GetDataDir() string {
	return mc.datadir
}

func (mc *Client) SetDelegation(tks []*Token, duration int64) error {
	tokenInfos := make([]*entity.TokenInfo, 0, len(tks))
	for _, tk := range tks {
		tokenInfo := utils.GetTokenInfoFromAddress(ctype.Hex2Addr(tk.Addr))
		tokenInfos = append(tokenInfos, tokenInfo)
	}
	return mc.c.SetDelegation(tokenInfos, duration)
}

// Destroy tries best to do clean up of current client
// note after this returns, all API calls to same client will crash.
// This is by design to catch invalid call flow.
// So caller should be careful about the lifecycle management.
// There is no need to call Destroy if client isn't init correctly
func (mc *Client) Destroy() {
	mc.c.Close()
	mc.c.ClearCallbacks()
	mc.c = nil
}

func (mc *Client) OpenETHChannel(dep *Deposit, cb ClientCallback) {
	mc.c.OpenChannel(&entity.TokenInfo{
		TokenType:    entity.TokenType_ETH,
		TokenAddress: ctype.EthTokenAddr.Bytes(),
	}, utils.Wei2BigInt(dep.Myamtwei), utils.Wei2BigInt(dep.Peeramtwei), cb)
}

// TODO(erctype): use proper enum based on tk.Erctype string
func (mc *Client) OpenTokenChannel(tk *Token, dep *Deposit, cb ClientCallback) {
	mc.c.OpenChannel(&entity.TokenInfo{
		TokenType:    entity.TokenType_ERC20,
		TokenAddress: ctype.Hex2Bytes(tk.Addr),
	}, utils.Wei2BigInt(dep.Myamtwei), utils.Wei2BigInt(dep.Peeramtwei), cb)
}

func (mc *Client) TcbOpenETHChannel(peerAmtWei string, cb ClientCallback) {
	mc.c.TcbOpenChannel(&entity.TokenInfo{
		TokenType: entity.TokenType_ETH,
	}, utils.Wei2BigInt(peerAmtWei), cb)
}

func (mc *Client) TcbOpenTokenChannel(tk *Token, peerAmtWei string, cb ClientCallback) {
	mc.c.TcbOpenChannel(&entity.TokenInfo{
		TokenType:    entity.TokenType_ERC20,
		TokenAddress: ctype.Hex2Bytes(tk.Addr),
	}, utils.Wei2BigInt(peerAmtWei), cb)
}
func (mc *Client) InstantiateChannelForToken(tk *Token, cb ClientCallback) {
	mc.c.InstantiateChannelForToken(&entity.TokenInfo{
		TokenType:    entity.TokenType_ERC20,
		TokenAddress: ctype.Hex2Bytes(tk.Addr),
	}, cb)
}

func (mc *Client) DepositETH(amount string, callback DepositCallback) (string, error) {
	amtInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", common.ErrInvalidArg
	}
	return mc.c.Deposit(ctype.EthTokenAddr, amtInt, callback)
}

func (mc *Client) DepositERC20(
	token *Token, amount string, callback DepositCallback) (string, error) {
	address, err := utils.ValidateAndFormatAddress(token.Addr)
	if err != nil {
		return "", err
	}
	amtInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", common.ErrInvalidArg
	}
	return mc.c.Deposit(address, amtInt, callback)
}

func (mc *Client) MonitorDepositJob(jobID string, callback DepositCallback) {
	mc.c.MonitorDepositJob(jobID, callback)
}

func (mc *Client) RemoveDepositJob(jobID string) {
	mc.c.RemoveDepositJob(jobID)
}

func (mc *Client) WithdrawETH(amount string, callback CooperativeWithdrawCallback) (string, error) {
	amtInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", common.ErrInvalidArg
	}
	return mc.c.CooperativeWithdraw(ctype.EthTokenAddr, amtInt, callback)
}

func (mc *Client) WithdrawERC20(
	token *Token, amount string, callback CooperativeWithdrawCallback) (string, error) {
	tokenAddr, err := utils.ValidateAndFormatAddress(token.Addr)
	if err != nil {
		return "", err
	}
	amtInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", common.ErrInvalidArg
	}
	return mc.c.CooperativeWithdraw(tokenAddr, amtInt, callback)
}

func (mc *Client) MonitorCooperativeWithdrawJob(
	withdrawHash string, callback CooperativeWithdrawCallback) {
	mc.c.MonitorCooperativeWithdrawJob(withdrawHash, callback)
}

func (mc *Client) RemoveCooperativeWithdrawJob(withdrawHash string) {
	mc.c.RemoveCooperativeWithdrawJob(withdrawHash)
}

// GetChannelState returns a string for channel state for given token
func (mc *Client) GetChannelState(tk *Token) string {
	tokenAddr, err := utils.ValidateAndFormatAddress(tk.Addr)
	if err != nil {
		return "INVALID_TOKEN_ADDRESS"
	}
	return mc.c.GetChannelState(tokenAddr)
}

// HasPendingOpenChanRequest is expected to be called when GetChannelState returns NOT_FOUND
// in this case, it's helpful to know whether client has a pending onchain open channel request so app
// can update UI and/or try open again.
// Note pending true is only possible for client iniated onchain openchannel, not TCB
//
// Internally we save the blockNum when start openchan request, if current blockNum <= saved+OpenChannelTimeout
// we return true. The value is set to 0 in openchan callback so future calls will
// return false (assume real blkNum is much larger than OpenChannelTimeout)
func (mc *Client) HasPendingOpenChanRequest(tk *Token) bool {
	return mc.c.HasPendingOpenChanRequest(sdkToken2entityToken(tk))
}

// QueryReceivingCapacity Check whether address has also joined Celer and
// returns its join status and free balance in a decimal string.  It is useful for
// checking the state of the intended receiver.
// If the given address has not joined Celer, an empty string will
// be returned.
func (mc *Client) QueryReceivingCapacity(addr string) (*CelerStatus, error) {
	joinStatus, freeBalance, err := mc.c.IsConnectedToCeler(ctype.EthTokenAddrStr, addr)
	return &CelerStatus{
		JoinStatus:  int32(joinStatus),
		FreeBalance: freeBalance}, err
}

// QueryReceivingCapacityOnToken Check whether address has also joined Celer on tokenAddr and
// returns its join status and free balance in a decimal string.  It is useful for
// checking the state of the intended receiver.
// If the given address has not joined Celer, an empty string will
// be returned.
func (mc *Client) QueryReceivingCapacityOnToken(tokenAddr string, addr string) (*CelerStatus, error) {
	if addr == "" {
		return &CelerStatus{
			JoinStatus:  int32(rpc.JoinCelerStatus_NOT_JOIN),
			FreeBalance: ""}, errors.New("Invalid input addr")
	}
	joinStatus, freeBalance, err := mc.c.IsConnectedToCeler(tokenAddr, addr)
	return &CelerStatus{
		JoinStatus:  int32(joinStatus),
		FreeBalance: freeBalance}, err
}

// Get celer offchain ETH balance
func (mc *Client) GetBalance() (*Balance, error) {
	return mc.GetBalanceERC20(ctype.EthTokenAddrStr)
}

// GetBalanceERC20 gets celer offchain tokenAddr balance
func (mc *Client) GetBalanceERC20(tokenAddr string) (*Balance, error) {
	token, err := utils.ValidateAndFormatAddress(tokenAddr)
	if err != nil {
		return nil, err
	}
	a, p, r, err := mc.c.GetTokenBalance(token)
	if err != nil {
		return nil, err
	}
	return &Balance{
		a.String(),
		p.String(),
		r.String(),
	}, nil
}

func (mc *Client) SyncOnChainChannelStates(tk *Token) error {
	token := sdkToken2entityToken(tk)
	return mc.c.SyncOnChainChannelStates(token)
}

func (mc *Client) IntendWithdraw(tokenInfo *TokenInfo, amount string) error {
	amt := new(big.Int)
	_, success := amt.SetString(amount, 10)
	if !success {
		return errors.New("Invalid withdraw amount")
	}
	return mc.c.IntendWithdraw(
		&entity.TokenInfo{
			TokenType:    entity.TokenType(int32(tokenInfo.TokenType)),
			TokenAddress: ctype.Hex2Bytes(tokenInfo.TokenAddress),
		},
		amt)
}

func (mc *Client) ConfirmWithdraw(tokenInfo *TokenInfo) error {
	return mc.c.ConfirmWithdraw((&entity.TokenInfo{
		TokenType:    entity.TokenType(int32(tokenInfo.TokenType)),
		TokenAddress: ctype.Hex2Bytes(tokenInfo.TokenAddress),
	}))
}

func (mc *Client) IntendSettlePaymentChannel(tokenInfo *TokenInfo) error {
	return mc.c.IntendSettlePaymentChannel(&entity.TokenInfo{
		TokenType:    entity.TokenType(int32(tokenInfo.TokenType)),
		TokenAddress: ctype.Hex2Bytes(tokenInfo.TokenAddress),
	})
}

func (mc *Client) ConfirmSettlePaymentChannel(tokenInfo *TokenInfo) error {
	return mc.c.ConfirmSettlePaymentChannel(&entity.TokenInfo{
		TokenType:    entity.TokenType(int32(tokenInfo.TokenType)),
		TokenAddress: ctype.Hex2Bytes(tokenInfo.TokenAddress),
	})
}

func (mc *Client) GetSettleFinalizedTimeForPaymentChannel(tokenInfo *TokenInfo) (int64, error) {
	time, err := mc.c.GetSettleFinalizedTime(&entity.TokenInfo{
		TokenType:    entity.TokenType(int32(tokenInfo.TokenType)),
		TokenAddress: ctype.Hex2Bytes(tokenInfo.TokenAddress),
	})
	if err != nil {
		return 0, err
	}
	return time.Int64(), nil
}

// SignData signs arbitrary data and returns the signature
func (mc *Client) SignData(data []byte) ([]byte, error) {
	return mc.c.SignState(data), nil
}

func (mc *Client) GetCurrentBlockNumber() int64 {
	return int64(mc.c.GetCurrentBlockNumberUint64())
}

// SetMsgDropper will drop grpc msgs, used for testing only
func (mc *Client) SetMsgDropper(dropRecv, dropSend bool) {
	mc.c.SetMsgDropper(dropRecv, dropSend)
}

// PayHistoryIterator is an iterator to get pay history.
type PayHistoryIterator struct {
	myAddr        string
	beforeTs      int64
	rpcClient     rpc.RpcClient
	hasMoreResult bool
	smallestPayID string
	// signFunc signs param using myAddr private key.
	signFunc func([]byte) []byte
}

// HasMoreResult returns false if the iterator has gone over all pays in history.
func (iter *PayHistoryIterator) HasMoreResult() bool {
	return iter.hasMoreResult
}

// NextPage returns next (earlier) batch of pay history in JSON string and error.
// itemsPerPage specifies the max number of pays in the response.
// History entries returned in json encoding string and reverse-chronological order.
// If there is more to retrieve, hasMoreResult filed in iterator is set to true .
// Note that hasMoreResult could be true when #(entries left for retrival before calling the func) happens to be same as itemsPerPage.
// In this case, the next time calling NextPage will return empty history set and set hasMoreResult field in iterator to false.
func (iter *PayHistoryIterator) NextPage(itemsPerPage int32) (string, error) {
	if !iter.hasMoreResult {
		return "[]", nil
	}
	if iter.beforeTs == 0 {
		iter.beforeTs = time.Now().Unix()
	}
	if iter.smallestPayID == "" {
		iter.smallestPayID = ctype.PayID2Hex(ctype.ZeroPayID)
	}
	ts, tsSig := utils.GetTsAndSig(iter.signFunc)
	req := &rpc.GetPayHistoryRequest{
		Peer:          iter.myAddr,
		BeforeTs:      iter.beforeTs,
		ItemsPerPage:  itemsPerPage,
		SmallestPayId: iter.smallestPayID,
		Ts:            ts,
		TsSig:         tsSig,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	resp, err := iter.rpcClient.GetPayHistory(ctx, req)
	if err != nil {
		return "[]", err
	}
	pays := resp.GetPays()
	// if array length is 0 or array is nil, return emtpy array json string
	if len(pays) == 0 {
		return "[]", nil
	}
	// Pays in response is sorted in reverse chronological order.
	// The last pay in the list will be the earliest one in the batch.
	// For pays with same time stamp, it's sorted by payID in ascending order.
	// Update iterator internal state so that it will fetch older pays next time.
	// Update smallestPayID to solve multiple-pay-same-ts issue.
	if len(pays) > 0 {
		iter.beforeTs = pays[len(pays)-1].GetCreateTs()
		iter.smallestPayID = pays[len(pays)-1].GetPayId()
	}
	iter.hasMoreResult = int32(len(pays)) == itemsPerPage

	paysBytes, err := json.Marshal(pays)
	if err != nil {
		return "[]", err
	}
	return string(paysBytes), nil
}

// GetPayHistoryIterator returns a celer crypto pay history iterator. Caller can call "NextPage" to get
// paginated pay history.
func (mc *Client) GetPayHistoryIterator() (*PayHistoryIterator, error) {
	conn, err := mc.c.GetRpcClientToOsp()
	if err != nil {
		return nil, err
	}
	return &PayHistoryIterator{
		myAddr:        ctype.Addr2Hex(mc.c.GetMyEthAddr()),
		beforeTs:      0,
		rpcClient:     conn,
		signFunc:      mc.c.SignState,
		hasMoreResult: true,
	}, nil
}
