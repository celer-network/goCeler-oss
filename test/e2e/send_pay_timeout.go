// Copyright 2018-2020 Celer Network

package e2e

import (
	"testing"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/testing/testapp"
	"github.com/celer-network/goutils/log"
)

func sendEthPayTimeout(t *testing.T) {
	log.Info("============== start test sendEthPayTimeout ==============")
	defer log.Info("============== end test sendEthPayTimeout ==============")
	t.Parallel()
	sendPayTimeout(t, entity.TokenType_ETH, tokenAddrEth)
}

func sendPayTimeout(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for sendPayTimeout token", tokenAddr, addrs)
	if tokenAddr != tokenAddrEth {
		err = tf.FundAccountsWithErc20(tokenAddr, addrs, accountBalance)
		if err != nil {
			t.Error(err)
			return
		}
	}
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

	_, err = c1.OpenChannel(c1EthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = c2.OpenChannel(c2EthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	constructor := testapp.GetSingleSessionConstructor(
		[]ctype.Addr{
			ctype.Hex2Addr(c1EthAddr),
			ctype.Hex2Addr(c2EthAddr),
		})
	appChanID, err := c1.NewAppChannelOnVirtualContract(
		testapp.AppCode,
		constructor,
		testapp.Nonce.Uint64(),
		testapp.Timeout.Uint64())
	if err != nil {
		t.Error(err)
		return
	}

	c1Cond1 := &entity.Condition{
		ConditionType:          entity.ConditionType_VIRTUAL_CONTRACT,
		VirtualContractAddress: ctype.Hex2Bytes(appChanID),
		ArgsQueryFinalization:  []byte{},
		ArgsQueryOutcome:       []byte{2},
	}
	c1Cond2 := &entity.Condition{
		ConditionType:          entity.ConditionType_VIRTUAL_CONTRACT,
		VirtualContractAddress: ctype.Hex2Bytes(appChanID),
		ArgsQueryFinalization:  []byte{},
		ArgsQueryOutcome:       []byte{3},
	}
	c2Cond := &entity.Condition{
		ConditionType:          entity.ConditionType_VIRTUAL_CONTRACT,
		VirtualContractAddress: ctype.Hex2Bytes(appChanID),
		ArgsQueryFinalization:  []byte{},
		ArgsQueryOutcome:       []byte{2},
	}

	// source pay in full
	timeout := uint64(3)
	_, err = c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{c1Cond1}, timeout)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{c1Cond2}, timeout)
	if err != nil {
		t.Error(err)
		return
	}
	p3, err := c2.SendPaymentWithBooleanConditions(
		c1EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{c2Cond}, timeout)
	if err != nil {
		t.Error(err)
		return
	}
	payTime, err := c1.GetCurrentBlockNumber()
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentPending(p3, c2, c1)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-2"),
		"2",
		tf.AddAmtStr(initialBalance, "-1"))
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-1"),
		"1",
		tf.AddAmtStr(initialBalance, "-2"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.WaitUntilDeadline(payTime + timeout + 10)
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("c1 settle expired pays")
	err = c1.SettleExpiredPays(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p3, nil, c1)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		initialBalance,
		"0",
		initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-1"),
		"1",
		tf.AddAmtStr(initialBalance, "-2"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("c2 settle expired pays -")
	err = c2.SettleExpiredPays(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p3, c2, nil)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		initialBalance,
		"0",
		initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.AssertBalance(
		tokenAddr,
		initialBalance,
		"0",
		initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	p4, err := c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("c1 settle expired pays")
	err = c1.SettleExpiredPays(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("c2 settle expired pays")
	err = c2.SettleExpiredPays(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p4, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-1"),
		"0",
		tf.AddAmtStr(initialBalance, "1"))
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "1"),
		"0",
		tf.AddAmtStr(initialBalance, "-1"))
	if err != nil {
		t.Error(err)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
}
