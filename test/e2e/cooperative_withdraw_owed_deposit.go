// Copyright 2018-2020 Celer Network

package e2e

import (
	"testing"

	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

func cooperativeWithdrawOwedDeposit(t *testing.T) {
	log.Info("============== start test cooperativeWithdrawOwedDeposit ==============")
	defer log.Info("============== end test cooperativeWithdrawOwedDeposit ==============")
	t.Parallel()
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for cooperativeWithdrawOwedDeposit", addrs)
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
	defer c2.Kill()

	initialBalance := "9"
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

	sendAmt := "3"
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

	resp, err := c2.CooperativeWithdraw(entity.TokenType_ETH, tokenAddrEth, "10")
	if err != nil {
		t.Error(err)
		return
	}
	if resp.TxHash == "" {
		t.Error("CooperativeWithdraw TxHash empty")
	}
	err = c2.AssertBalance(tokenAddrEth, "2", "0", "6")
	if err != nil {
		t.Error(err)
		return
	}
}
