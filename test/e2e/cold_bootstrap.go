// Copyright 2018-2020 Celer Network

package e2e

import (
	"math/big"
	"testing"
	"time"

	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

func coldBootstrap(t *testing.T) {
	log.Info("============== start test coldBootstrap ==============")
	defer log.Info("============== end test coldBootstrap ==============")
	t.Parallel()
	const c1OnChainBalance = "50000000000000000000"
	const c2OnChainBalance = "0"
	ks, addrs, err := tf.CreateAccountsWithBalance(1, c1OnChainBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for coldBootstrap 1", addrs)
	c1KeyStore := ks[0]
	c1EthAddr := addrs[0]

	// Client c2 does a cold bootstrap.
	ks, addrs, err = tf.CreateAccountsWithBalance(1, c2OnChainBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for coldBootstrap 2", addrs)
	c2KeyStore := ks[0]
	c2EthAddr := addrs[0]

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

	_, err = c1.OpenChannel(c1EthAddr, entity.TokenType_ETH, tokenAddrEth, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	const c2PeerAmt = "800000000000000000"
	_, err = c2.OpenChannel(c2EthAddr, entity.TokenType_ETH, tokenAddrEth, "0", c2PeerAmt)
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

	const c2BalanceBefore = "800000000000000000"
	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(initialBalance, "-1"),
		"0",
		tf.AddAmtStr(initialBalance, "1"))
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.AssertBalance(tokenAddrEth, "1", "0", tf.AddAmtStr(c2BalanceBefore, "-1"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== DONE: c2 cold bootstrap and get payment PASSED =====")
}

func tcbOpenChannel(t *testing.T) {
	log.Info("============== start test tcbOpenChannel ==============")
	defer log.Info("============== end test tcbOpenChannel ==============")
	t.Parallel()
	// Full lifecycle of tcb
	// 1. create c1 with ERC20 token and c2 with 0 ERC20 token
	// 2. c1 opens ERC20 channel in standard way and c2 opens a TCB channel
	// 3. c1 sends all tokens to c2. c2 should now have non-zero balance.
	// 4. c2 instantiate the channel on-chain
	// 5. c2 sends some tokens to c1. c2 should see balance reduction
	// 6. c2 settles (intentSettle+confirmSettle) the channel
	// 7. c2 should have non-zero ERC20 token on-chain
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for tcbOpenChannel", addrs)
	tokenAddr := tokenAddrErc20
	tokenType := entity.TokenType_ERC20
	err = tf.FundAccountsWithErc20(tokenAddr, []string{addrs[0]}, accountBalance)
	if err != nil {
		t.Error(err)
		return
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
	_, err = c2.TcbOpenChannel(c2EthAddr, tokenType, tokenAddr, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== Client-1 sends payment to client-2 =====")

	p1, err := c1.SendPayment(c2EthAddr, "2000000000000000000", tokenType, tokenAddr)
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
		tf.AddAmtStr(initialBalance, "-2000000000000000000"),
		"0",
		tf.AddAmtStr(initialBalance, "2000000000000000000"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		"2000000000000000000",
		"0",
		tf.AddAmtStr(initialBalance, "-2000000000000000000"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== DONE: c2 tcb and get payment PASSED =====")
	log.Info("===== c2 tcb instantiating channel =====")
	_, err = c2.InstantiateChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.AssertBalance(
		tokenAddr,
		"2000000000000000000",
		"0",
		tf.AddAmtStr(initialBalance, "-2000000000000000000"))
	if err != nil {
		t.Error(err)
		return
	}
	log.Info("===== c2 send back =====")

	p2, err := c2.SendPayment(c1EthAddr, "1000000000000000000", tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p2, c2, c1)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddr,
		"1000000000000000000",
		"0",
		tf.AddAmtStr(initialBalance, "-1000000000000000000"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== c2 tcb intend to settle =====")
	err = c2.IntendSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	finalizedTime, err := c2.GetSettleFinalizedTime(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.WaitUntilDeadline(finalizedTime)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== c2 tcb confirm settle =====")
	err = c2.ConfirmSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	c2EthClient, err := getEthClient(c2EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	c2Amt, err := c2.GetAccountBalance(tokenAddr, c2EthAddr, c2EthClient)
	if err != nil {
		t.Error(err)
		return
	}

	c2TargetAmt, _ := new(big.Int).SetString("1000000000000000000", 10)

	log.Infoln("c2 on chain balance", c2Amt)
	if c2Amt.Cmp(c2TargetAmt) != 0 {
		t.Errorf("wrong c2 on-chain balance: expect %v, got %v", c2TargetAmt, c2Amt)
	}
}

func concurrentOpenChannel(t *testing.T) {
	log.Info("============== start test concurrentOpenChannel ==============")
	defer log.Info("============== end test concurrentOpenChannel ==============")
	t.Parallel()
	const c1OnChainBalance = "0"

	// Client c1 does a cold bootstrap.
	ks, addrs, err := tf.CreateAccountsWithBalance(1, c1OnChainBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for concurrentOpenChannel", addrs)
	c1KeyStore := ks[0]
	c1EthAddr := addrs[0]

	c1, err := tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	errCh := make(chan error)
	go func() {
		time.Sleep(100 * time.Millisecond)
		_, err2 := c1.TcbOpenChannel(c1EthAddr, entity.TokenType_ETH, tokenAddrEth, initialBalance)
		errCh <- err2
	}()
	_, err = c1.OpenChannel(c1EthAddr, entity.TokenType_ETH, tokenAddrEth, "0", initialBalance)
	err2 := <-errCh
	if err == nil && err2 == nil {
		// Cannot both succeed.
		t.Error(err, err2, "Both open went through")
		return
	}
	log.Info("===== DONE: concurrent open channel PASSED =====")
}

/*
func tcbOverCommitOpenChannel(t *testing.T) {
	log.Info("============== start test tcbOverCommitOpenChannel ==============")
	defer log.Info("============== end test tcbOverCommitOpenChannel ==============")
	t.Parallel()
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for tcbOverCommitOpenChannel", addrs)
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

	fakeToken := "1111111111111111111111111111111111111111" // token not on-chain at all, defined in rt_config.json
	_, err = c1.TcbOpenChannel(c1EthAddr, entity.TokenType_ERC20, fakeToken, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = c2.TcbOpenChannel(c2EthAddr, entity.TokenType_ERC20, fakeToken, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("===== osp sends payment to client using fake token =====")
	p1, err := requestSvrSendToken(c2EthAddr, "1", fakeToken)
	if err != nil {
		t.Error(err)
		return
	}
	err = waitForPaymentCompletion(p1, nil, c2)
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.AssertBalance(fakeToken, "1", "0", tf.AddAmtStr(initialBalance, "-1"))
	if err != nil {
		t.Error(err)
		return
	}
	log.Info("===== DONE: c2 tcb and get payment PASSED =====")
	log.Info("===== c2 sending to c1 using fake token =====")
	p2, err := c2.SendPayment(c1EthAddr, "1", entity.TokenType_ERC20, fakeToken)
	if err != nil {
		t.Error(err)
		return
	}
	err = waitForPaymentCompletion(p2, c2, c1)
	err = c2.AssertBalance(fakeToken, "0", "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	err = c1.AssertBalance(fakeToken, "1", "0", tf.AddAmtStr(initialBalance, "-1"))
	if err != nil {
		t.Error(err)
		return
	}
	log.Info("===== DONE: c2 sending to c1 using fake token PASSED =====")
}
*/
