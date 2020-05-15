// Copyright 2018-2020 Celer Network

package e2e

import (
	"math/big"
	"os"
	"os/exec"
	"testing"

	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

const (
	emptyChannel = iota
	oneSimplexChannel
	fullDuplexChannel
)

func settleErc20ChannelEmpty(t *testing.T) {
	log.Info("============== start test settleErc20ChannelEmpty ==============")
	defer log.Info("============== end test settleErc20ChannelEmpty ==============")
	t.Parallel()
	settleChannel(t, entity.TokenType_ERC20, tokenAddrErc20, emptyChannel)
}

func settleErc20ChannelOneSimplex(t *testing.T) {
	log.Info("============== start test settleErc20ChannelOneSimplex ==============")
	defer log.Info("============== end test settleErc20ChannelOneSimplex ==============")
	t.Parallel()
	settleChannel(t, entity.TokenType_ERC20, tokenAddrErc20, oneSimplexChannel)
}

func settleErc20ChannelFullDuplex(t *testing.T) {
	log.Info("============== start test settleErc20ChannelFullDuplex ==============")
	defer log.Info("============== end test settleErc20ChannelFullDuplex ==============")
	t.Parallel()
	settleChannel(t, entity.TokenType_ERC20, tokenAddrErc20, fullDuplexChannel)
}

func settleErc20ChannelWithDispute(t *testing.T) {
	log.Info("============== start test settleErc20ChannelWithDispute ==============")
	defer log.Info("============== end test settleErc20ChannelWithDispute ==============")
	t.Parallel()
	settleWithDispute(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func settleErc20ChannelWithReopen(t *testing.T) {
	log.Info("============== start test settleErc20ChannelWithReopen ==============")
	defer log.Info("============== end test settleErc20ChannelWithReopen ==============")
	t.Parallel()
	settleChannelWithReopen(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func ospIntendSettleErc20Channel(t *testing.T) {
	log.Info("============== start test ospIntendSettleErc20Channel ==============")
	defer log.Info("============== end test ospIntendSettleErc20Channel ==============")
	t.Parallel()
	ospIntendSettleChannel(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func settleChannel(t *testing.T, tokenType entity.TokenType, tokenAddr string, mode int) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for settleChannel token", tokenAddr, addrs)
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

	if mode == oneSimplexChannel || mode == fullDuplexChannel {
		p1, err2 := c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
		if err2 != nil {
			t.Error(err2)
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

	if mode == fullDuplexChannel {
		p2, err2 := c2.SendPayment(c1EthAddr, "2", tokenType, tokenAddr)
		if err2 != nil {
			t.Error(err2)
			return
		}

		err = waitForPaymentCompletion(p2, c2, c1)
		if err != nil {
			t.Error(err)
			return
		}

		err = c1.AssertBalance(
			tokenAddr,
			tf.AddAmtStr(initialBalance, "1"),
			"0",
			tf.AddAmtStr(initialBalance, "-1"))
		if err != nil {
			t.Error(err)
			return
		}

		err = c2.AssertBalance(
			tokenAddr,
			tf.AddAmtStr(initialBalance, "-1"),
			"0",
			tf.AddAmtStr(initialBalance, "1"))
		if err != nil {
			t.Error(err)
			return
		}
	}

	err = c1.IntendSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.IntendSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	finalizedTime, err := c1.GetSettleFinalizedTime(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	finalizedTime2, err := c2.GetSettleFinalizedTime(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if finalizedTime2 > finalizedTime {
		finalizedTime = finalizedTime2
	}

	err = c1.WaitUntilDeadline(finalizedTime)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.ConfirmSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.ConfirmSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	c1EthClient, err := getEthClient(c1EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	c2EthClient, err := getEthClient(c2EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	c1Amt, err := c1.GetAccountBalance(tokenAddr, c1EthAddr, c1EthClient)
	if err != nil {
		t.Error(err)
		return
	}
	c2Amt, err := c2.GetAccountBalance(tokenAddr, c2EthAddr, c2EthClient)
	if err != nil {
		t.Error(err)
		return
	}
	c1TargetAmt := big.NewInt(0)
	c2TargetAmt := big.NewInt(0)

	if mode == emptyChannel {
		c1TargetAmt.SetString(accountBalance, 10)
		c2TargetAmt.SetString(accountBalance, 10)
	} else if mode == oneSimplexChannel {
		c1TargetAmt.SetString(tf.AddAmtStr(accountBalance, "-1"), 10)
		c2TargetAmt.SetString(tf.AddAmtStr(accountBalance, "1"), 10)
	} else if mode == fullDuplexChannel {
		c1TargetAmt.SetString(tf.AddAmtStr(accountBalance, "1"), 10)
		c2TargetAmt.SetString(tf.AddAmtStr(accountBalance, "-1"), 10)
	}

	if c1Amt.Cmp(c1TargetAmt) != 0 {
		t.Errorf("wrong c1 on-chain balance after settlement: expect %v, got %v", c1TargetAmt, c1Amt)
	}
	if c2Amt.Cmp(c2TargetAmt) != 0 {
		t.Errorf("wrong c2 on-chain balance after settlement: expect %v, got %v", c2TargetAmt, c2Amt)
	}
}

func settleWithDispute(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	// To test case of dispute, we need an state that is valid but staled in client.
	// To achieve this, the test does followings:
	// 1. c1, c2 open channel with osp respectively
	// 2. kill c1 and backup storedir of c1
	// 3. bring up c1 and let c1 to send one token to c2. seqNum of payment channel will increase.
	// 4. kill c1 and restore the back-up storedir of c1
	// 5. bring up c1 again and let c1 settle payment channel. As the seq number in back-up is lower,
	//    this will trigger OSP to dispute with newer channel state.
	// 6. Check c1 on-chain balance that should align with view of osp channel state.
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for settleWithDispute token", tokenAddr, addrs)
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

	log.Infoln("Starting c1 first time")
	c1, err := tf.StartClientController(
		tf.GetNextClientPort(), c1KeyStore, noProxyProfile, c1StoreSettleDisputeDir, "c1")
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
	log.Infoln("Killing c1")
	c1.KillWithoutRemovingKeystore()

	c1StoreDirStale := "/tmp/c1StoreStale"
	defer os.Remove(c1StoreDirStale)
	log.Infoln("Backing up", c1StoreSettleDisputeDir, "to", c1StoreDirStale)
	copyFile(c1StoreSettleDisputeDir, c1StoreDirStale)

	log.Infoln("Restarting c1 after backup")
	c1, err = tf.StartClientController(
		tf.GetNextClientPort(), c1KeyStore, noProxyProfile, c1StoreSettleDisputeDir, "c1")
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	err = c1.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("Sending pay to c2")

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

	log.Infoln("Checking c1 balance")
	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-1"),
		"0",
		tf.AddAmtStr(initialBalance, "1"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Infoln("Killing c1 second time for restoring")
	c1.KillWithoutRemovingKeystore()
	log.Infoln("Restoring", c1StoreSettleDisputeDir, "from", c1StoreDirStale)
	copyFile(c1StoreDirStale, c1StoreSettleDisputeDir)
	log.Infoln("Restarting c1 AFTER restoring")
	c1, err = tf.StartClientController(
		tf.GetNextClientPort(), c1KeyStore, noProxyProfile, c1StoreSettleDisputeDir, "c1")
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	err = c1.IntendSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	finalizedTime, err := c1.GetSettleFinalizedTime(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.WaitUntilDeadline(finalizedTime + 1)
	if err != nil {
		t.Error(err)
		return
	}

	// check finalized time again as it may change due to server response
	finalizedTime2, err := c1.GetSettleFinalizedTime(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if finalizedTime2 != finalizedTime {
		err = c1.WaitUntilDeadline(finalizedTime2 + 1)
		if err != nil {
			t.Error(err)
			return
		}
	}

	err = c1.ConfirmSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(1)

	c1EthClient, err := getEthClient(c1EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	c1Amt, err := c1.GetAccountBalance(tokenAddr, c1EthAddr, c1EthClient)
	if err != nil {
		t.Error(err)
		return
	}
	c1TargetAmt := big.NewInt(0)

	c1TargetAmt.SetString(tf.AddAmtStr(accountBalance, "-1"), 10)

	if c1Amt.Cmp(c1TargetAmt) != 0 {
		t.Errorf("wrong c1 on-chain balance after settlement: expect %v, got %v", c1TargetAmt, c1Amt)
	}
}

func copyFile(src string, dst string) {
	rm := exec.Command("rm", "-rf", dst)
	err := rm.Run()
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
	cp := exec.Command("cp", "-a", src, dst)
	err = cp.Run()
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
}

// Test case of reopen channel after settle.
func settleChannelWithReopen(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	// The test does followings:
	// 1. c1, c2 open channel with OSP respectively.
	// 2. c1 send 1 token to c2.
	// 3. check c1, c2 offchain balance. c2 should get 1 token.
	// 4. c1 settle the channel with OSP.
	// 5. c1 reopen channel with OSP.
	// 6. c1 send 1 token to c2.
	// 7. check c1, c2 balance. c1 should 1 token reduced, c2 should have 2 more token than initial deposit.
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for settleChannelWithReopen token", tokenAddr, addrs)
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

	err = c1.IntendSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	finalizedTime, err := c1.GetSettleFinalizedTime(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.WaitUntilDeadline(finalizedTime)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.ConfirmSettlePaymentChannel(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(1)

	log.Infoln("c1 reopening channel")
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

	log.Infoln("c1 sending to c2 again")
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
		tf.AddAmtStr(initialBalance, "2"),
		"0",
		tf.AddAmtStr(initialBalance, "-2"))
	if err != nil {
		t.Error(err)
		return
	}
}

func ospIntendSettleChannel(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(1, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for ospIntendWithdraw token", tokenAddr, addrs)
	if tokenAddr != tokenAddrEth {
		err = tf.FundAccountsWithErc20(tokenAddr, addrs, accountBalance)
		if err != nil {
			t.Error(err)
			return
		}
	}
	cKeyStore := ks[0]
	cEthAddr := addrs[0]

	c, err := tf.StartC1WithoutProxy(cKeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Kill()

	channel, err := c.OpenChannel(cEthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	err = c.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	p1, err := c.SendPayment(ospEthAddr, "10", tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p1, c, nil)
	if err != nil {
		t.Error(err)
		return
	}

	tf.StartProcess(outRootDir+"ospcli",
		"-ks", ospKeystore,
		"-nopassword",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-intendsettle",
		"-cid", channel.ChannelId,
		"-logprefix", "cli").Wait()

	var finalizedTime uint64
	for {
		log.Infoln("Wait for intendSettle tx")
		finalizedTime, err = c.GetSettleFinalizedTime(tokenType, tokenAddr)
		if err != nil {
			t.Error(err)
			return
		}
		if finalizedTime > 0 {
			break
		}
		sleep(1)
	}

	err = c.WaitUntilDeadline(finalizedTime)
	if err != nil {
		t.Error(err)
		return
	}

	tf.StartProcess(outRootDir+"ospcli",
		"-ks", ospKeystore,
		"-nopassword",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-confirmsettle",
		"-cid", channel.ChannelId,
		"-logprefix", "cli").Wait()

	cEthClient, err := getEthClient(cEthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	cAmt, err := c.GetAccountBalance(tokenAddr, cEthAddr, cEthClient)
	if err != nil {
		t.Error(err)
		return
	}
	cTargetAmt := big.NewInt(0)
	cTargetAmt.SetString(tf.AddAmtStr(accountBalance, "-10"), 10)

	if cAmt.Cmp(cTargetAmt) != 0 {
		t.Errorf("wrong c1 on-chain balance after settlement: expect %v, got %v", cTargetAmt, cAmt)
	}
}
