// Copyright 2018-2020 Celer Network

package e2e

import (
	"math/big"
	"testing"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

// adminSendToken combines two cases.
// 1. client first opens a channel and get offline. admin sending token should returns error
// 2. client gets online and the admin API should execute correctly.
func adminSendToken(t *testing.T) {
	log.Info("============== start test adminSendToken ==============")
	defer log.Info("============== end test adminSendToken ==============")
	t.Parallel()
	const c1OnChainBalance = "0"

	// Client c1 does a cold bootstrap.
	ks, addrs, err := tf.CreateAccountsWithBalance(1, c1OnChainBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for adminSendToken", addrs)
	c1KeyStore := ks[0]
	c1EthAddr := addrs[0]

	c1, err := tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	const c1PeerAmt = "800000000000000000"
	_, err = c1.OpenChannel(c1EthAddr, entity.TokenType_ETH, tokenAddrEth, "0", c1PeerAmt)
	if err != nil {
		t.Error(err)
		return
	}
	log.Info("killing c1")
	c1.KillWithoutRemovingKeystore()

	log.Info("===== OSP sends payment to OFFLINE client-1 =====")
	requestSvrSendToken(c1EthAddr, "1", "")

	log.Info("restarting c1")
	c1, err = tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	// Expect the payment to be delivered to c1 after it restarts.
	const c1BalanceBefore = "800000000000000000"
	err = c1.AssertBalance(tokenAddrEth, "1", "0", tf.AddAmtStr(c1BalanceBefore, "-1"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== OSP sends payment to ONLINE client-1 =====")
	c1ReceiveDoneChan := make(chan string, 1)
	c1.SubscribeIncomingPayments(c1ReceiveDoneChan)
	payID, err := requestSvrSendToken(c1EthAddr, "1", "")
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(payID, nil, c1)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(tokenAddrEth, "2", "0", tf.AddAmtStr(c1BalanceBefore, "-2"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== DONE: OSP sends payment PASSED =====")
}

func requestSvrSendToken(receiver, amt, tokenAddr string) (string, error) {
	amtInt, ok := new(big.Int).SetString(amt, 10)
	if !ok {
		return "", common.ErrInvalidArg
	}
	payId, err := utils.RequestSendToken(sAdminWeb, ctype.Hex2Addr(receiver), ctype.Hex2Addr(tokenAddr), amtInt)
	return ctype.PayID2Hex(payId), err
}

func requestSendToken(adminWebAddr, receiver, amt, tokenAddr string) (string, error) {
	amtInt, ok := new(big.Int).SetString(amt, 10)
	if !ok {
		return "", common.ErrInvalidArg
	}
	payId, err := utils.RequestSendToken(adminWebAddr, ctype.Hex2Addr(receiver), ctype.Hex2Addr(tokenAddr), amtInt)
	return ctype.PayID2Hex(payId), err
}
