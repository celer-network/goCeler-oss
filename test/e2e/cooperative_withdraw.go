// Copyright 2018-2020 Celer Network

package e2e

import (
	"fmt"
	"testing"

	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

func cooperativeWithdrawEth(t *testing.T) {
	log.Info("============== start test cooperativeWithdrawEth ==============")
	defer log.Info("============== end test cooperativeWithdrawEth ==============")
	t.Parallel()
	cooperativeWithdraw(t, entity.TokenType_ETH, tokenAddrEth)
}

func cooperativeWithdrawErc20(t *testing.T) {
	log.Info("============== start test cooperativeWithdrawErc20 ==============")
	defer log.Info("============== end test cooperativeWithdrawErc20 ==============")
	t.Parallel()
	cooperativeWithdraw(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func cooperativeWithdrawEthWithRestart(t *testing.T) {
	log.Info("============== start test cooperativeWithdrawEthWithRestart ==============")
	defer log.Info("============== end test cooperativeWithdrawEthWithRestart ==============")
	t.Parallel()
	cooperativeWithdrawWithRestart(t, entity.TokenType_ETH, tokenAddrEth)
}

func cooperativeWithdraw(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(1, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for cooperativeWithdraw token", tokenAddr, addrs)
	cKeyStore := ks[0]
	cEthAddr := addrs[0]

	if tokenAddr != tokenAddrEth {
		err = tf.FundAccountsWithErc20(tokenAddr, addrs, accountBalance)
		if err != nil {
			t.Error(err)
			return
		}
	}

	c, err := tf.StartC1WithoutProxy(cKeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Kill()

	_, err = c.OpenChannel(cEthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	err = c.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := c.CooperativeWithdraw(tokenType, tokenAddr, "123")
	if err != nil {
		t.Error(err)
		return
	}
	if resp.TxHash == "" {
		t.Error("CooperativeWithdraw TxHash empty")
		return
	}
	err = c.AssertBalance(
		tokenAddr, tf.AddAmtStr(initialBalance, "-123"), "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
}

func cooperativeWithdrawWithRestart(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(1, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for cooperativeWithdrawWithRestart token", tokenAddr, addrs)
	cKeyStore := ks[0]
	cEthAddr := addrs[0]

	if tokenAddr != tokenAddrEth {
		err = tf.FundAccountsWithErc20(tokenAddr, addrs, accountBalance)
		if err != nil {
			t.Error(err)
			return
		}
	}

	c, err := tf.StartC1WithoutProxy(cKeyStore)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = c.OpenChannel(cEthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := c.CooperativeWithdrawNonBlocking(tokenType, tokenAddr, "123")
	if err != nil {
		t.Error(err)
		return
	}
	jobID := resp.GetJobId()
	log.Infoln("submitted withdraw job", jobID)
	sleep(1)
	c.KillWithoutRemovingKeystore()

	cnew, err := tf.StartC1WithoutProxy(cKeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer cnew.Kill()

	log.Infoln("restart and monitor withdraw job", jobID)
	resp, err = cnew.MonitorCooperativeWithdrawJob(jobID)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.TxHash == "" {
		t.Error("CooperativeWithdraw TxHash empty")
		return
	}
	err = cnew.AssertBalance(
		tokenAddr, tf.AddAmtStr(initialBalance, "-123"), "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
}

func cooperativeWithdrawAfterSendPay(t *testing.T) {
	log.Info("============== start test cooperativeWithdrawAfterSendPay ==============")
	defer log.Info("============== end test cooperativeWithdrawAfterSendPay ==============")
	t.Parallel()
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for cooperativeWithdrawAfterSendPay", addrs)
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

	initialBalance := "900000000000000000"
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

	sendAmt := "300000000000000000"
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

	resp, err := c2.CooperativeWithdraw(entity.TokenType_ETH, tokenAddrEth, "1000000000000000000")
	if err != nil {
		t.Error(err)
		return
	}
	if resp.TxHash == "" {
		t.Error("CooperativeWithdraw TxHash empty")
	}
	err = c2.AssertBalance(tokenAddrEth, "200000000000000000", "0", "600000000000000000")
	if err != nil {
		t.Error(err)
		return
	}
}

func cooperativeWithdrawAndSendInvalidPay(t *testing.T) {
	log.Info("============== start test cooperativeWithdrawAndSendInvalidPay ==============")
	defer log.Info("============== end test cooperativeWithdrawAndSendInvalidPay ==============")
	t.Parallel()
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for cooperativeWithdrawAndSendInvalidPay", addrs)
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

	initialBalance := "900000000000000000"
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

	_, err = c1.CooperativeWithdrawNonBlocking(entity.TokenType_ETH, tokenAddrEth, "600000000000000000")
	if err != nil {
		t.Error(err)
		return
	}
	sendAmt := "400000000000000000"
	_, err = c1.SendPayment(c2EthAddr, sendAmt, entity.TokenType_ETH, tokenAddrEth)
	if err == nil {
		err2 := fmt.Errorf("should not able to send")
		t.Error(err2)
		return
	}

	err = c1.AssertBalance(tokenAddrEth, "300000000000000000", "0", "900000000000000000")
	if err != nil {
		t.Error(err)
		return
	}
}

func cooperativeWithdrawInsufficient(t *testing.T) {
	log.Info("============== start test cooperativeWithdrawInsufficient ==============")
	defer log.Info("============== end test cooperativeWithdrawInsufficient ==============")
	t.Parallel()
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for cooperativeWithdrawInsufficient", addrs)
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

	initialBalance := "900000000000000000"
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

	sendAmt := "900000000000000000"
	_, err = c1.SendPayment(c2EthAddr, sendAmt, entity.TokenType_ETH, tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(2)
	err = c1.AssertBalance(tokenAddrEth, "0", "0", "1800000000000000000")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = c1.CooperativeWithdraw(entity.TokenType_ETH, tokenAddrEth, "600000000000000000")
	if err == nil {
		err2 := fmt.Errorf("Should not able to withdraw")
		t.Error(err2)
		return
	}
}
