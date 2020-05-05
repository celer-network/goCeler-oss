// Copyright 2018-2020 Celer Network

package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

func sendCondPayWithEth(t *testing.T) {
	log.Info("============== start test sendCondPayWithEth ==============")
	defer log.Info("============== end test sendCondPayWithEth ==============")
	t.Parallel()
	sendCondPay(t, entity.TokenType_ETH, tokenAddrEth)
}
func sendCondPayWithErc20(t *testing.T) {
	log.Info("============== start test sendCondPayWithErc20 ==============")
	defer log.Info("============== end test sendCondPayWithErc20 ==============")
	t.Parallel()
	sendCondPay(t, entity.TokenType_ERC20, tokenAddrErc20)
}
func sendCondPayWithEthDstOffline(t *testing.T) {
	log.Info("============== start test sendCondPayWithEthDstOffline ==============")
	defer log.Info("============== end test sendCondPayWithEthDstOffline ==============")
	t.Parallel()
	sendCondPayDstOffline(t, entity.TokenType_ETH, tokenAddrEth)
}
func delegateSendEth(t *testing.T) {
	log.Info("============== start test delegateSendEth ==============")
	defer log.Info("============== end test delegateSendEth ==============")
	t.Parallel()
	delegateSendCondPay(t, entity.TokenType_ETH, tokenAddrEth)
}
func delegateSendErc20(t *testing.T) {
	log.Info("============== start test delegateSendErc20 ==============")
	defer log.Info("============== end test delegateSendErc20 ==============")
	t.Parallel()
	delegateSendCondPay(t, entity.TokenType_ERC20, tokenAddrErc20)
}
func sendCondPayWithEthToOSP(t *testing.T) {
	log.Info("============== start test sendCondPayWithEthToOSP ==============")
	defer log.Info("============== end test sendCondPayWithEthToOSP ==============")
	t.Parallel()
	sendCondPayToOSP(t, entity.TokenType_ETH, tokenAddrEth)
}
func sendCondPayNoEnoughErc20AtSrc(t *testing.T) {
	log.Info("============== start test sendCondPayNoEnoughErc20AtSrc ==============")
	defer log.Info("============== end test sendCondPayNoEnoughErc20AtSrc ==============")
	t.Parallel()
	sendCondPayNoEnoughFundAtSrc(t, entity.TokenType_ERC20, tokenAddrErc20)
}
func sendCondPayNoEnoughErc20AtOsp(t *testing.T) {
	log.Info("============== start test sendCondPayNoEnoughErc20AtOsp ==============")
	defer log.Info("============== end test sendCondPayNoEnoughErc20AtOsp ==============")
	t.Parallel()
	sendCondPayNoEnoughFundAtOsp(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func sendCondPayNoEnoughFundAtSrc(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	// 1. Bring up c1 and c2 with sending capacity 1.
	// 2. c1 tries to send 2 (which exceeds sending capacity of c1) to c2.
	//    This should return error straight when we call send pay.
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for sendCondPayNoEnoughFundAtSrc token", tokenAddr, addrs)
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

	smallDeposit := "100000000000000000"
	_, err = c1.OpenChannel(c1EthAddr, tokenType, tokenAddr, smallDeposit, smallDeposit)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = c2.OpenChannel(c2EthAddr, tokenType, tokenAddr, smallDeposit, smallDeposit)
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("================== C1 sending to C2 exceeding sending capacity =======================")
	amtToSend := "200000000000000000"
	_, err = c1.SendPayment(c2EthAddr, amtToSend, tokenType, tokenAddr)
	if err == nil {
		t.Error("Sending more than c1 has but no err", err)
		return
	}
}

func sendCondPayNoEnoughFundAtOsp(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	// 1. Bring up c1 and c2 with sending capacity 2 and receiving capacity 1 respectively.
	// 2. c1 tries to send 2 (which is within c1 sending capacity but exceeds c2 receiving capacity) to c2.
	//    This should not return error straight when we call send pay. OSP, however, will errors due to
	//    insufficient fund and will tell c1 asynchronously triggering DstUnreachable callback of c1.
	// 3. Check for the "unreachable" error on c1.
	// 4. c1 sends 1 (which within all capacities) to c2.
	// 5. Check sending is successful.
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for sendCondPayNoEnoughFundAtOsp token", tokenAddr, addrs)
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

	smallDeposit := "100000000000000000"
	largeDeposit := "200000000000000000"
	_, err = c1.OpenChannel(c1EthAddr, tokenType, tokenAddr, largeDeposit, smallDeposit)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = c2.OpenChannel(c2EthAddr, tokenType, tokenAddr, largeDeposit, smallDeposit)
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("================== C1 sending to C2 exceeding C2 receiving capacity =======================")
	// Send while c2 is offline, should not change balance.
	c1SendCompleteChan := make(chan string, 1)
	c1SendErrChan := make(chan string, 1)
	c1.SubscribeOutgoingPayments(c1SendCompleteChan, c1SendErrChan)
	p1, err := c1.SendPayment(c2EthAddr, largeDeposit, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	c1Err := <-c1SendErrChan
	if !strings.HasPrefix(c1Err, "Unreachable") {
		t.Error("Unreachable callback not triggered")
		return
	}

	err = waitForPaymentCompletion(p1, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(tokenAddr, largeDeposit, "0", smallDeposit)
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("================== C1 sending to C2 within C2 receiving capacity =======================")
	p2, err := c1.SendPayment(c2EthAddr, smallDeposit, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p2, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(tokenAddr, "100000000000000000", "0", "200000000000000000")
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(tokenAddr, "300000000000000000", "0", "0")
	if err != nil {
		t.Error(err)
		return
	}
}

func sendCondPayDstOffline(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for sendCondPayDstOffline token", tokenAddr, addrs)
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
	c2.KillWithoutRemovingKeystore()

	log.Infoln("================== C1 sending to C2 when C2 is offline =======================")
	c1SendCompleteChan := make(chan string, 1)
	c1SendErrChan := make(chan string, 1)
	c1.SubscribeOutgoingPayments(c1SendCompleteChan, c1SendErrChan)
	// Send while c2 is offline, should not change balance.
	p1, err := c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	c1Err := <-c1SendErrChan
	if !strings.HasPrefix(c1Err, "Unreachable") {
		t.Error("Unreachable callback not triggered, err:", c1Err)
		return
	}

	err = waitForPaymentCompletion(p1, c1, c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	// Restart c2
	log.Infoln("================== Restarting C2 =======================")
	c2, err = tf.StartC2WithoutProxy(c2KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c2.Kill()
	err = c2.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("================== Resending to C2 =======================")

	p2, err := c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p2, c1, c2)
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

	sleep(1)
	// Check pay history from osp.
	// There should be two pays in c1 history.
	pays, hasMoreResult, err := c1.GetPayHistory(true /*fromStart*/, 1 /*itemsPerPage*/)
	if err != nil {
		t.Error(err)
		return
	}
	if len(pays) != 1 {
		t.Errorf("wrong c1 history first batch. pays: %v", pays)
		return
	}
	if !hasMoreResult || int(pays[0].GetState()) != structs.PayState_COSIGNED_PAID {
		t.Errorf("wrong c1 history first batch. pays: %v, hasMoreResult: %t, state: %d", pays, hasMoreResult, pays[0].GetState())
		return
	}
	pays, hasMoreResult, err = c1.GetPayHistory(false /*fromStart*/, 2 /*itemsPerPage*/)
	if err != nil {
		t.Error(err)
		return
	}
	if len(pays) != 1 {
		t.Errorf("wrong c1 history second batch. pays: %v", pays)
		return
	}
	if hasMoreResult || int(pays[0].GetState()) != structs.PayState_COSIGNED_CANCELED {
		t.Errorf("wrong c1 history second batch. pays: %v, hasMoreResult: %t, state: %d", pays, hasMoreResult, pays[0].GetState())
		return
	}
	pays, hasMoreResult, err = c2.GetPayHistory(true /*fromStart*/, 1)
	if err != nil {
		t.Error(err)
		return
	}
	if len(pays) != 1 {
		t.Errorf("wrong c1 history first batch. pays: %v", pays)
		return
	}
	if !hasMoreResult || int(pays[0].GetState()) != structs.PayState_COSIGNED_PAID {
		t.Errorf("wrong c2 history first batch. pays: %v, hasMoreResult: %t, state: %d", pays, hasMoreResult, pays[0].GetState())
		return
	}
	pays, hasMoreResult, err = c2.GetPayHistory(false /*fromStart*/, 2)
	if err != nil {
		t.Error(err)
		return
	}
	if len(pays) != 1 {
		t.Errorf("wrong c1 history second batch. pays: %v", pays)
		return
	}
	if hasMoreResult || int(pays[0].GetState()) != structs.PayState_COSIGNED_CANCELED {
		t.Errorf("wrong c2 history second batch. pays: %v, hasMoreResult: %t, state: %d", pays, hasMoreResult, pays[0].GetState())
		return
	}
}

func delegateSendCondPay(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for delegateSendCondPay token", tokenAddr, addrs)
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

	log.Infoln("Openning channel for c1")
	_, err = c1.OpenChannel(c1EthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("Openning channel for c2")
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
	log.Infoln("---------------- Authorizing delegation ----------------")
	c2.SetDelegation([]string{tokenAddr}, 500)
	// shut down c2 so dst will be offline in send pay below.
	c2.KillWithoutRemovingKeystore()
	sleep(1)

	p1, err := c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	p2, err := c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p1, c1, nil)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p2, c1, nil)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-2"),
		"0",
		tf.AddAmtStr(initialBalance, "2"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("--------------- Restarting c2 -----------------")
	c2, err = tf.StartC2WithoutProxy(c2KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c2.Kill()

	time.Sleep(time.Second)

	// c2 should get the money automatically after connecting to osp.
	err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "2"),
		"0",
		tf.AddAmtStr(initialBalance, "-2"))
	if err != nil {
		t.Error(err)
		return
	}
}

func sendCondPay(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for sendCondPay token", tokenAddr, addrs)
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

	sleep(2)
	log.Info("------------ start sending pay --------")
	p1, err := c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
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

func sendCondPayToOSP(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(1, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for sendCondPayToOSP token", tokenAddr, addrs)
	if tokenAddr != tokenAddrEth {
		err = tf.FundAccountsWithErc20(tokenAddr, addrs, accountBalance)
		if err != nil {
			t.Error(err)
			return
		}
	}
	c1KeyStore := ks[0]
	c1EthAddr := addrs[0]

	c1, err := tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	_, err = c1.OpenChannel(c1EthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	err = c1.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	p1, err := c1.SendPayment(ospEthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p1, c1, nil)
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
}
