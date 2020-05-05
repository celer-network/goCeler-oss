// Copyright 2018-2020 Celer Network

package e2e

import (
	"testing"

	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

func clientRecovery(t *testing.T) {
	log.Info("============== start test clientRecovery ==============")
	defer log.Info("============== end test clientRecovery ==============")
	t.Parallel()
	const c1OnChainBalance = "50000000000000000000"
	ks, addrs, err := tf.CreateAccountsWithBalance(2, c1OnChainBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for clientRecovery", addrs)
	c1KeyStore := ks[0]
	c2KeyStore := ks[1]
	c1EthAddr := addrs[0]
	c2EthAddr := addrs[1]

	c1, err := tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	c2, err := tf.StartC2WithoutProxy(c2KeyStore)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = c1.OpenChannel(c1EthAddr, entity.TokenType_ETH, tokenAddrEth, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = c2.OpenChannel(c2EthAddr, entity.TokenType_ETH, tokenAddrEth, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== Client-1 sends payment to client-2 =====")
	p1, err := c1.SendPayment(c2EthAddr, sendAmt, entity.TokenType_ETH, tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p1, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	const balanceBefore = "5000000000000000000"
	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(balanceBefore, "-1"),
		"0",
		tf.AddAmtStr(balanceBefore, "1"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(balanceBefore, "1"),
		"0",
		tf.AddAmtStr(balanceBefore, "-1"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== Kill and restart client-2 =====")
	c2.KillWithoutRemovingKeystore()

	c2New, err := tf.StartC2WithoutProxy(c2KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c2New.Kill()

	_, err = c2New.OpenChannel(c2EthAddr, entity.TokenType_ETH, tokenAddrEth, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== Client-2 restarted, sending payment to client-1 =====")
	backAmt := "3"
	p2, err := c2New.SendPayment(c1EthAddr, backAmt, entity.TokenType_ETH, tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p2, c2New, c1)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2New.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(balanceBefore, "-2"),
		"0",
		tf.AddAmtStr(balanceBefore, "2"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(balanceBefore, "2"),
		"0",
		tf.AddAmtStr(balanceBefore, "-2"))
	if err != nil {
		t.Error(err)
		return
	}
	log.Info("===== DONE: recovery and back payment PASSED =====")
}
