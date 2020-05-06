// Copyright 2018-2020 Celer Network

package e2e

import (
	"testing"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	ta "github.com/celer-network/goCeler/testing/testapp"
	"github.com/celer-network/goutils/log"
	ec "github.com/ethereum/go-ethereum/common"
)

func sendPaySettleWithEthDstReconnect(t *testing.T) {
	log.Info("============== start test sendPaySettleWithEthDstReconnect ==============")
	defer log.Info("============== end test sendPaySettleWithEthDstReconnect ==============")
	t.Parallel()
	sendPaySettleDstReconnect(t, entity.TokenType_ETH, tokenAddrEth)
}

func sendPaySettleDstReconnect(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for sendPaySettleDstReconnect token", tokenAddr, addrs)

	if tokenAddr != tokenAddrEth {
		err = tf.FundAccountsWithErc20(tokenAddr, addrs, accountBalance)
		if err != nil {
			t.Error(err)
			return
		}
	}

	c1KS := ks[0] // keystore
	c2KS := ks[1]
	c1Addr := addrs[0]
	c2Addr := addrs[1]

	c1, err := tf.StartC1WithoutProxy(c1KS)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	c2, err := tf.StartC2WithoutProxy(c2KS)
	if err != nil {
		t.Error(err)
		return
	}
	defer c2.Kill()

	_, err = c1.OpenChannel(c1Addr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = c2.OpenChannel(c2Addr, tokenType, tokenAddr, initialBalance, initialBalance)
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
	constructor := ta.GetSingleSessionConstructor(
		[]ec.Address{
			ctype.Hex2Addr(c1Addr),
			ctype.Hex2Addr(c2Addr),
		},
	)
	appChanId, err := c1.NewAppChannelOnVirtualContract(
		ta.AppCode,
		constructor,
		ta.Nonce.Uint64(),
		ta.Timeout.Uint64(),
	)
	if err != nil {
		t.Error(err)
		return
	}

	cond := &entity.Condition{
		ConditionType:          entity.ConditionType_VIRTUAL_CONTRACT,
		VirtualContractAddress: ctype.Hex2Bytes(appChanId),
		ArgsQueryFinalization:  []byte{},
		ArgsQueryOutcome:       []byte{1},
	}
	conds := []*entity.Condition{cond}
	timeout := uint64(100)

	sleep(2)
	log.Info("================ Start sending payment with boolean condition  =====================")

	log.Info("----------------- C1 sends cond payment to C2 ------------------")

	payID, err := c1.SendPaymentWithBooleanConditions(c2Addr, sendAmt, tokenType, tokenAddr, conds, timeout)
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

	log.Info("-------------- C2 disconnects ---------------")
	c2.KillWithoutRemovingKeystore()
	sleep(2)

	log.Info("-------------- C1 confirms cond payment to C2 ----------------")
	err = c1.ConfirmBooleanPay(payID)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(payID, c1, nil)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-1"),
		"0",
		tf.AddAmtStr(initialBalance, "1"),
	)

	log.Info("------------- C2 reconnects ---------------")
	c2, err = tf.StartC2WithoutProxy(c2KS)
	if err != nil {
		t.Error(err)
		return
	}
	defer c2.Kill()

	err = waitForPaymentCompletion(payID, nil, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "1"),
		"0",
		tf.AddAmtStr(initialBalance, "-1"),
	)
	if err != nil {
		t.Error(err)
		return
	}

}
