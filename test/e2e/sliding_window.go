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

func slidingWindowEth(t *testing.T) {
	log.Info("============== start test slidingWindow ==============")
	defer log.Info("============== end test slidingWindow ==============")
	t.Parallel()
	slidingWindow(t, entity.TokenType_ETH, tokenAddrEth)
}

func slidingWindow(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for slidingWindow token", tokenAddr, addrs)
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

	// construct payment condition
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
	cond := &entity.Condition{
		ConditionType:          entity.ConditionType_VIRTUAL_CONTRACT,
		VirtualContractAddress: ctype.Hex2Bytes(appChanID),
		ArgsQueryFinalization:  []byte{},
		ArgsQueryOutcome:       []byte{1},
	}
	conds := []*entity.Condition{cond}
	var timeout uint64
	timeout = 100

	log.Info("============ start sending pay ============")

	log.Info("------------ c1 send cond pay p1 --------")
	p1, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p2 --------")
	p2, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentPending(p2, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-2"),
		"2",
		initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		initialBalance,
		"0",
		tf.AddAmtStr(initialBalance, "-2"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c2 drop send ============")
	err = c2.SetMsgDropper(false, true)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p3 --------")
	p3, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p4 --------")
	p4, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send confirm p1 --------")
	err = c1.ConfirmBooleanPay(p1)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p1, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-4"),
		"3",
		tf.AddAmtStr(initialBalance, "1"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "1"),
		"0",
		tf.AddAmtStr(initialBalance, "-4"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c2 resume send ============")
	err = c2.SetMsgDropper(false, false)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send confirm p2 --------")
	err = c1.ConfirmBooleanPay(p2)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c2 send reject p3 --------")
	err = c2.RejectBooleanPay(p3)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send confirm p4 --------")
	err = c1.ConfirmBooleanPay(p4)
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
		tf.AddAmtStr(initialBalance, "-3"),
		"0",
		tf.AddAmtStr(initialBalance, "3"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "3"),
		"0",
		tf.AddAmtStr(initialBalance, "-3"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c1 drop send ============")
	err = c1.SetMsgDropper(false, true)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p5 --------")
	p5, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p6 --------")
	p6, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	sleep(2)
	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-5"),
		"2",
		tf.AddAmtStr(initialBalance, "3"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "3"),
		"0",
		tf.AddAmtStr(initialBalance, "-3"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c1 resume send ============")
	err = c1.SetMsgDropper(false, false)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send confirm p5 --------")
	err = c1.ConfirmBooleanPay(p5)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send confirm p6 --------")
	err = c1.ConfirmBooleanPay(p6)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p6, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-5"),
		"0",
		tf.AddAmtStr(initialBalance, "5"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "5"),
		"0",
		tf.AddAmtStr(initialBalance, "-5"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c1 drop send ============")
	err = c1.SetMsgDropper(false, true)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cpay p7 --------")
	_, err = c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cpay p8 --------")
	_, err = c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	sleep(2)
	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-7"),
		"2",
		tf.AddAmtStr(initialBalance, "5"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "5"),
		"0",
		tf.AddAmtStr(initialBalance, "-5"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c1 resume send ============")
	err = c1.SetMsgDropper(false, false)

	log.Info("============ c1 disconnect ============")
	c1.KillWithoutRemovingKeystore()

	log.Info("============ c1 reconnect ============")
	c1, err = tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	sleep(2)
	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-7"),
		"0",
		tf.AddAmtStr(initialBalance, "7"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "7"),
		"0",
		tf.AddAmtStr(initialBalance, "-7"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c2 drop send ============")
	err = c2.SetMsgDropper(false, true)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p9 --------")
	p9, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p10 --------")
	p10, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p11 --------")
	p11, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 confirms pay p9 --------")
	err = c1.ConfirmBooleanPay(p9)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p9, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-10"),
		"2",
		tf.AddAmtStr(initialBalance, "8"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "8"),
		"0",
		tf.AddAmtStr(initialBalance, "-10"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c2 resume send ============")
	err = c2.SetMsgDropper(false, false)

	log.Info("============ c2 disconnect ============")
	c2.KillWithoutRemovingKeystore()

	log.Info("============ c2 reconnect ============")
	c2, err = tf.StartC2WithoutProxy(c2KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c2.Kill()

	log.Info("------------ c1 confirms pay p10 --------")
	err = c1.ConfirmBooleanPay(p10)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 confirms pay p11 --------")
	err = c1.ConfirmBooleanPay(p11)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p11, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-10"),
		"0",
		tf.AddAmtStr(initialBalance, "10"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "10"),
		"0",
		tf.AddAmtStr(initialBalance, "-10"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c1 drop send ============")
	err = c1.SetMsgDropper(false, true)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("-- c1 send cond pay p12 with too large timeout --")
	p12, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, 2000)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p13 --------")
	_, err = c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cond pay p14 --------")
	_, err = c1.SendPayment(ospEthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	sleep(2)
	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-13"),
		"2",
		tf.AddAmtStr(initialBalance, "11"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "10"),
		"0",
		tf.AddAmtStr(initialBalance, "-10"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("============ c1 resume send ============")
	err = c1.SetMsgDropper(false, false)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cpay p15 --------")
	p15, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p12, c1, nil)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentPending(p15, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-13"),
		"2",
		tf.AddAmtStr(initialBalance, "11"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "10"),
		"0",
		tf.AddAmtStr(initialBalance, "-12"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cpay p16 --------")
	_, err = c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ c1 send cpay p17 --------")
	p17, err := c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p17, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-15"),
		"2",
		tf.AddAmtStr(initialBalance, "13"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "12"),
		"0",
		tf.AddAmtStr(initialBalance, "-14"))
	if err != nil {
		t.Error(err)
		return
	}

}
