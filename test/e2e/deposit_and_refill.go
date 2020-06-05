// Copyright 2018-2020 Celer Network

package e2e

import (
	"math/big"
	"testing"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/rtconfig"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

func clientDepositEth(t *testing.T) {
	log.Info("============== start test clientDepositEth ==============")
	defer log.Info("============== end test clientDepositEth ==============")
	t.Parallel()
	clientDeposit(t, entity.TokenType_ETH, tokenAddrEth)
}

func clientDepositErc20WithRestart(t *testing.T) {
	log.Info("============== start test clientDepositErc20WithRestart ==============")
	defer log.Info("============== end test clientDepositErc20WithRestart ==============")
	t.Parallel()
	clientDepositWithRestart(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func ospDepositAndRefill(t *testing.T) {
	log.Info("============== start test ospDepositAndRefill ==============")
	defer log.Info("============== end test ospDepositAndRefill ==============")
	t.Parallel()

	err := rtconfig.Init(rtConfig)
	if err != nil {
		t.Error("init runtime config failed:", err)
		return
	}

	ethRefillThreshold := rtconfig.GetRefillThreshold(tokenAddrEth)
	ethRefillAmount, RefillMaxWait := rtconfig.GetRefillAmountAndMaxWait(tokenAddrEth)
	log.Infoln("ETH refill threshold", ethRefillThreshold, "amount", ethRefillAmount, "maxWait", RefillMaxWait)

	erc20RefillThreshold := rtconfig.GetRefillThreshold(tokenAddrErc20)
	erc20RefillAmount, RefillMaxWait := rtconfig.GetRefillAmountAndMaxWait(tokenAddrErc20)
	log.Infoln("Erc20 refill threshold", erc20RefillThreshold, "amount", erc20RefillAmount, "maxWait", RefillMaxWait)

	pollingInterval := rtconfig.GetDepositPollingInterval()
	minBatchSize := rtconfig.GetDepositMinBatchSize()
	maxBatchSize := rtconfig.GetDepositMaxBatchSize()
	log.Info("Deposit pooling Interval", pollingInterval, "batch size min", minBatchSize, "max", maxBatchSize)

	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for ospDepositAndRefill", addrs)

	err = tf.FundAccountsWithErc20(tokenAddrErc20, addrs, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}

	c1KeyStore := ks[0]
	c2KeyStore := ks[1]
	c1EthAddr := addrs[0]
	c2EthAddr := addrs[1]
	log.Info("client addresses", addrs)

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

	ethInitBalance := ethRefillThreshold.String()
	cid, err := c1.OpenChannel(c1EthAddr, entity.TokenType_ETH, tokenAddrEth, ethInitBalance, ethInitBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("ETH channel with c1:", cid.ChannelId)
	cid, err = c2.OpenChannel(c2EthAddr, entity.TokenType_ETH, tokenAddrEth, ethInitBalance, ethInitBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("ETH channel with c2:", cid.ChannelId)

	erc20InitBalance := erc20RefillThreshold.String()
	cid, err = c1.OpenChannel(c1EthAddr, entity.TokenType_ERC20, tokenAddrErc20, erc20InitBalance, erc20InitBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("ERC20 channel with c1:", cid.ChannelId)
	cid, err = c2.OpenChannel(c2EthAddr, entity.TokenType_ERC20, tokenAddrErc20, erc20InitBalance, erc20InitBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("ERC20 channel with c2:", cid.ChannelId)

	log.Info("--------------- deposit eth to channel with c1 ---------------")
	depositID, err := requestSvrDeposit(c1EthAddr, tokenAddrEth, "1", true, 5)
	if err != nil {
		t.Error(err)
		return
	}

	res, err := querySvrDeposit(depositID)
	if err != nil {
		t.Error(err)
		return
	}
	if res.Error != "" {
		t.Error(res.Error)
		return
	}
	if res.DepositState != rpc.DepositState_Deposit_QUEUED {
		t.Error("invalid deposit state")
		return
	}

	depositID, err = requestSvrDeposit(c1EthAddr, tokenAddrEth, "1", false, 4)
	if err != nil {
		t.Error(err)
		return
	}

	depositID, err = requestSvrDeposit(c1EthAddr, tokenAddrEth, "1", false, 0)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("--------------- deposit erc20 to channel with c1 ---------------")
	depositID, err = requestSvrDeposit(c1EthAddr, tokenAddrErc20, "1", true, 0)
	if err != nil {
		t.Error(err)
		return
	}

	depositID, err = requestSvrDeposit(c1EthAddr, tokenAddrErc20, "1", false, 1)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("--------------- deposit eth to channel with c2 ---------------")
	depositID, err = requestSvrDeposit(c2EthAddr, tokenAddrEth, "1", true, 2)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("--------------- send eth to c2, trigger refill ---------------")
	_, err = requestSvrSendToken(c2EthAddr, "1", tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = requestSvrSendToken(c2EthAddr, "1", tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = requestSvrSendToken(c2EthAddr, "1", tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("--------------- deposit erc20 to channel with c2 ---------------")
	depositID, err = requestSvrDeposit(c2EthAddr, tokenAddrErc20, "1", true, 1)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("--------------- send erc20 to c2, trigger refill ---------------")
	_, err = requestSvrSendToken(c2EthAddr, "10", tokenAddrErc20)
	if err != nil {
		t.Error(err)
		return
	}

	depositID, err = requestSvrDeposit(c2EthAddr, tokenAddrErc20, "1", false, 0)
	if err != nil {
		t.Error(err)
		return
	}

	depositID, err = requestSvrDeposit(c2EthAddr, tokenAddrErc20, "1", false, 1)
	if err != nil {
		t.Error(err)
		return
	}

	sleep(5)
	err = syncOnChainStates(c1)
	if err != nil {
		t.Error(err)
		return
	}
	err = syncOnChainStates(c2)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(ethInitBalance, "1"),
		"0",
		tf.AddAmtStr(ethInitBalance, "2"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr(erc20InitBalance, "1"),
		"0",
		tf.AddAmtStr(erc20InitBalance, "1"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(ethInitBalance, "4"),
		"0",
		tf.AddAmtStr(ethInitBalance, "-3", ethRefillAmount.String()))
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr(erc20InitBalance, "11"),
		"0",
		tf.AddAmtStr(erc20InitBalance, "-10", ethRefillAmount.String()))
	if err != nil {
		t.Error(err)
		return
	}

	res, err = querySvrDeposit(depositID)
	if err != nil {
		t.Error(err)
		return
	}
	if res.Error != "" {
		t.Error(res.Error)
		return
	}
	if res.DepositState != rpc.DepositState_Deposit_SUCCEEDED {
		t.Error("invalid deposit state")
		return
	}

	log.Info("--------------- send eth and erc20 to c1, trigger refill ---------------")
	_, err = requestSvrSendToken(c1EthAddr, "10", tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = requestSvrSendToken(c1EthAddr, "5", tokenAddrErc20)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = requestSvrSendToken(c1EthAddr, "5", tokenAddrErc20)
	if err != nil {
		t.Error(err)
		return
	}

	sleep(5)

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(ethInitBalance, "11"),
		"0",
		tf.AddAmtStr(ethInitBalance, "-8"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr(erc20InitBalance, "11"),
		"0",
		tf.AddAmtStr(erc20InitBalance, "-9"))
	if err != nil {
		t.Error(err)
		return
	}

	sleep(4)
	err = syncOnChainStates(c1)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(ethInitBalance, "11"),
		"0",
		tf.AddAmtStr(ethInitBalance, "-8", ethRefillAmount.String()))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr(erc20InitBalance, "11"),
		"0",
		tf.AddAmtStr(erc20InitBalance, "-9", ethRefillAmount.String()))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("--------------- deposit eth to channel with c1 ---------------")
	ospFreeEth := tf.AddAmtStr(ethInitBalance, "-8", ethRefillAmount.String())
	_, err = requestSvrDeposit(c1EthAddr, tokenAddrEth, "10", false, 0)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(5)

	log.Info("--------------- send large amount of eth to c1, trigger c1 sync onchain ---------------")
	_, err = requestSvrSendToken(c1EthAddr, tf.AddAmtStr(ospFreeEth, "5"), tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(3)
	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(ethInitBalance, "11", ospFreeEth, "5"),
		"0",
		"5")
	if err != nil {
		t.Error(err)
		return
	}
}

func syncOnChainStates(c *tf.ClientController) error {
	err := c.SyncOnChainChannelStates(entity.TokenType_ETH, tokenAddrEth)
	if err != nil {
		return err
	}
	err = c.SyncOnChainChannelStates(entity.TokenType_ERC20, tokenAddrErc20)
	if err != nil {
		return err
	}
	return nil
}

func requestSvrDeposit(peerAddr, tokenAddr, amt string, toPeer bool, maxWait uint64) (string, error) {
	amtInt, ok := new(big.Int).SetString(amt, 10)
	if !ok {
		return "", common.ErrInvalidArg
	}
	return utils.RequestDeposit(sAdminWeb, ctype.Hex2Addr(peerAddr), ctype.Hex2Addr(tokenAddr), amtInt, toPeer, maxWait)
}

func querySvrDeposit(depositID string) (*rpc.QueryDepositResponse, error) {
	return utils.QueryDeposit(sAdminWeb, depositID)
}

func clientDeposit(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(1, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for clientDeposit token", tokenAddr, addrs)
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

	resp, err := c.Deposit(tokenType, tokenAddr, "100")
	if err != nil {
		t.Error(err)
		return
	}
	if resp.TxHash == "" {
		t.Error("Deposit TxHash empty")
	}
	err = c.AssertBalance(tokenAddr, tf.AddAmtStr(initialBalance, "100"), "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
}

func clientDepositWithRestart(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(1, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for clientDepositWithRestart token", tokenAddr, addrs)
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

	resp, err := c.DepositNonBlocking(tokenType, tokenAddr, "100")
	if err != nil {
		t.Error(err)
		return
	}
	jobID := resp.GetJobId()
	log.Infoln("submitted deposit job", jobID)
	c.KillWithoutRemovingKeystore()

	cnew, err := tf.StartC1WithoutProxy(cKeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer cnew.Kill()

	log.Infoln("restart and monitor deposit job", jobID)
	resp, err = cnew.MonitorDepositJob(jobID)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.TxHash == "" {
		t.Error("Deposit TxHash empty")
	}
	err = cnew.AssertBalance(tokenAddr, tf.AddAmtStr(initialBalance, "100"), "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
}
