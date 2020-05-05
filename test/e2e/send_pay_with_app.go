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

func sendEthOnVirtualContractCondition(t *testing.T) {
	log.Info("============== start test sendEthOnVirtualContractCondition ==============")
	defer log.Info("============== end test sendEthOnVirtualContractCondition ==============")
	t.Parallel()
	sendPayOnVirtualContractCondition(t, entity.TokenType_ETH, tokenAddrEth)
}

func sendPayOnVirtualContractCondition(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for sendPayOnVirtualContractCondition token", tokenAddr, addrs)
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

	c1Cond := &entity.Condition{
		ConditionType:          entity.ConditionType_VIRTUAL_CONTRACT,
		VirtualContractAddress: ctype.Hex2Bytes(appChanID),
		ArgsQueryFinalization:  []byte{},
		ArgsQueryOutcome:       []byte{2},
	}
	// source pay in full
	payID, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{c1Cond}, 100)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentPending(payID, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-1"),
		"1",
		initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		initialBalance,
		"0",
		tf.AddAmtStr(initialBalance, "-1"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.ConfirmBooleanPay(payID)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(payID, c1, c2)
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

	// destination cancel
	payID, err = c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{c1Cond}, 100)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentPending(payID, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-2"),
		"1",
		tf.AddAmtStr(initialBalance, "1"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "1"),
		"0",
		tf.AddAmtStr(initialBalance, "-2"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.RejectBooleanPay(payID)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(payID, c1, c2)
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
}
