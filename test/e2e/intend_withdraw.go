// Copyright 2018-2020 Celer Network

package e2e

import (
	"testing"

	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

func ospIntendWithdrawErc20(t *testing.T) {
	log.Info("============== start test ospIntendWithdrawErc20 ==============")
	defer log.Info("============== end test ospIntendWithdrawErc20 ==============")
	t.Parallel()
	ospIntendWithdraw(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func clientIntendWithdrawErc20(t *testing.T) {
	log.Info("============== start test clientIntendWithdrawErc20 ==============")
	defer log.Info("============== end test clientIntendWithdrawErc20 ==============")
	t.Parallel()
	clientIntendWithdraw(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func ospIntendWithdraw(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
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

	// comment out when run with others, as the server onchain balance may be changed by other tests
	/*
		cEthClient, err := getEthClient(cEthAddr)
		if err != nil {
			t.Error(err)
			return
		}
		sAmtBefore, err := c.GetAccountBalance(tokenAddr, ospEthAddr, cEthClient)
	*/
	tf.StartProcess(outRootDir+"ospcli",
		"-ks", ospKeystore,
		"-nopassword",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-intendwithdraw",
		"-cid", channel.ChannelId,
		"-amount", "1.1",
		"-logprefix", "cli").Wait()
	tf.AdvanceBlocks(10)

	tf.StartProcess(outRootDir+"ospcli",
		"-ks", ospKeystore,
		"-nopassword",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-confirmwithdraw",
		"-cid", channel.ChannelId,
		"-logprefix", "cli").Wait()
	tf.AdvanceBlocks(5)

	err = c.SyncOnChainChannelStates(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	err = c.AssertBalance(
		tokenAddr, initialBalance, "0", tf.AddAmtStr(initialBalance, "-1100000000000000000"))
	if err != nil {
		t.Error(err)
		return
	}

	// comment out when run with others, as the server onchain balance may be changed by other tests
	/*
		sAmtAfter, err := c.GetAccountBalance(tokenAddr, ospEthAddr, cEthClient)
		target := big.NewInt(0)
		target.SetString("1100000000000000000", 10)
		target = target.Add(target, sAmtBefore)
		if sAmtAfter.Cmp(target) != 0 {
			t.Errorf("wrong client on-chain balance after confirm withdrawal: expect %v, got %v", target, sAmtAfter)
		}
	*/
}

func clientIntendWithdraw(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for clientIntendWithdraw token", tokenAddr, addrs)
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

	err = c.IntendWithdraw(tokenType, tokenAddr, "1000000000000000000")
	if err != nil {
		t.Error(err)
		return
	}

	err = c.ConfirmWithdraw(tokenType, tokenAddr)
	if err == nil {
		t.Error("confirm withdraw should fail")
		return
	}

	err = c.SyncOnChainChannelStates(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = c.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
}
