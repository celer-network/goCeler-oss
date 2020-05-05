// Copyright 2020 Celer Network

package e2e

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/rpc"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"google.golang.org/grpc"
)

func multiOspOpenChannelTest(t *testing.T) {
	log.Info("============== start test multiOspOpenChannelTest ==============")
	defer log.Info("============== end test multiOspOpenChannelTest ==============")
	// Let osp2 initiate openning channel with osp1.
	err := requestOpenChannel(o2AdminWeb, osp1EthAddr, initOspToOspBalance, initOspToOspBalance, tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("done open channel, waiting")
	time.Sleep(2 * time.Second)
	log.Infoln("sending token")
	// requestSvrSendToken is defined in admin.go. It will ask osp1 to send 1 token to osp2EthAddr.
	requestSvrSendToken(osp2EthAddr, "1", "")
	log.Infoln("done sending token, waiting")
	time.Sleep(1 * time.Second)
	// Ask osp1 balance
	free, err := getEthBalance(localhost+o1Port, osp2EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	// osp1 sends osp2 1 token. Remaining capacity is 1.
	if free != tf.AddAmtStr(initOspToOspBalance, "-1") {
		t.Errorf("expect %s sending capacity %s, got %s", osp1EthAddr, tf.AddAmtStr(initOspToOspBalance, "-1"), free)
		return
	}
	// Ask osp2 balance
	free, err = getEthBalance(localhost+o2Port, osp1EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	// osp2 got 1 token. Remaining capacity is 3.
	if free != tf.AddAmtStr(initOspToOspBalance, "1") {
		t.Errorf("expect %s sending capacity %s, got %s", osp2EthAddr, tf.AddAmtStr(initOspToOspBalance, "1"), free)
		return
	}

	log.Infoln("sending back")
	requestSendToken(o2AdminWeb, osp1EthAddr, "1", "")
}

func multiOspOpenChannelPolicyTest(t *testing.T) {
	log.Info("============== start test multiOspOpenChannelPolicyTest ==============")
	defer log.Info("============== end test multiOspOpenChannelPolicyTest ==============")
	tf.FundAccountsWithErc20(tokenAddrErc20, []string{osp2EthAddr}, accountBalance)
	// Let osp2 initiate openning channel with osp1 using bad deposit combination.
	err := requestOpenChannel(o2AdminWeb, osp1EthAddr, "20000000000000000000", "20000000000000000000", tokenAddrEth)
	if err == nil {
		t.Error("Expect to break policy")
		return
	}
	// ask osp1 to deposit 8. This should break ratio policy which is set to 1.0 in rt_config.json
	err = requestOpenChannel(o2AdminWeb, osp1EthAddr, "8", "1", tokenAddrEth)
	if err == nil {
		t.Error("Expect to break matching ratio policy")
		return
	}
}

func multiOspOpenChannelPolicyFallbackTest(t *testing.T) {
	log.Info("============== start test multiOspOpenChannelPolicyFallbackTest ==============")
	defer log.Info("============== end test multiOspOpenChannelPolicyFallbackTest ==============")
	tf.FundAccountsWithErc20(tokenAddrErc20, []string{osp2EthAddr}, accountBalance)
	// Let osp2 initiate openning channel with osp1 using erc20. This should fallback to client-osp policy
	// and fail because it doesn't meet the fallback policy.
	err := requestOpenChannel(o2AdminWeb, osp1EthAddr, "2" /*peerDeposit*/, "2" /*selfDeposit*/, tokenAddrErc20)
	if err == nil {
		t.Error("Expect to fail due to exceeding deposit in fallback")
		return
	}
	// Let osp2 initiate openning channel with osp1 using erc20. Deposit meets fallback aka client-osp policy
	err = requestOpenChannel(o2AdminWeb, osp1EthAddr, "1" /*peerDeposit*/, "1" /*selfDeposit*/, tokenAddrErc20)
	if err != nil {
		t.Error("Unable to fallback", err)
		return
	}
}

func requestOpenChannel(adminWebAddr, peerAddr, peerDeposit, selfDeposit, tokenAddr string) error {
	peerDepositInt, ok := new(big.Int).SetString(peerDeposit, 10)
	if !ok {
		return common.ErrInvalidArg
	}
	selfDepositInt, ok := new(big.Int).SetString(selfDeposit, 10)
	if !ok {
		return common.ErrInvalidArg
	}
	return utils.RequestOpenChannel(adminWebAddr, ctype.Hex2Addr(peerAddr), ctype.Hex2Addr(tokenAddr), peerDepositInt, selfDepositInt)
}

func getEthBalance(ospHTTPTarget string, osp2Addr string) (string, error) {
	conn, err := grpc.Dial(ospHTTPTarget, utils.GetClientTlsOption(), grpc.WithBlock(),
		grpc.WithTimeout(4*time.Second), grpc.WithKeepaliveParams(config.KeepAliveClientParams))
	if err != nil {
		return "", fmt.Errorf("fail to get peer status: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	response, err := rpc.NewRpcClient(conn).CelerGetPeerStatus(
		ctx,
		&rpc.PeerAddress{
			Address:   osp2Addr,
			TokenAddr: tokenAddrEth,
		},
	)
	if err != nil {
		return "", fmt.Errorf("fail to get peer status: %s", err)
	}
	return response.GetFreeBalance(), nil
}
