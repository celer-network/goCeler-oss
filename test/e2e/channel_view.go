// Copyright 2018-2020 Celer Network

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/testing/testapp"
	"github.com/celer-network/goutils/log"
)

func ethChannelView(t *testing.T) {
	log.Info("============== start test ethChannelView ==============")
	defer log.Info("============== end test ethChannelView ==============")
	t.Parallel()
	channelView(t, entity.TokenType_ETH, tokenAddrEth)
}

func erc20ChannelView(t *testing.T) {
	log.Info("============== start test erc20ChannelView ==============")
	defer log.Info("============== end test erc20ChannelView ==============")
	t.Parallel()
	channelView(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func channelView(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for channelView token", tokenAddr, addrs)
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

	cid1, err := c1.OpenChannel(c1EthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = c2.OpenChannel(c2EthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("--------- Channels Opened --------")

	p1, err := c1.SendPayment(c2EthAddr, "10", tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p1, c1, c2)
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
	appChanID, err := c2.NewAppChannelOnVirtualContract(
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

	p2, err := c2.SendPaymentWithBooleanConditions(
		c1EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	p3, err := c2.SendPaymentWithBooleanConditions(
		c1EthAddr, sendAmt, tokenType, tokenAddr, conds, timeout)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentPending(p3, c2, c1)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("--------- c1 drop recv ---------")
	err = c1.SetMsgDropper(true, false)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.ConfirmBooleanPay(p2)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentCompletion(p2, c2, c1)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = c2.SendPayment(c1EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(2)
	ts := time.Now()
	sleep(2)

	p4, err := requestSvrSendToken(c1EthAddr, "1", tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(2)

	fmt.Println()
	fmt.Println("-------------------------------------- channel cid")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "channel",
		"-cid", cid1.GetChannelId(),
		"-payhistory",
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
	fmt.Println("-------------------------------------- channel peer token")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "channel",
		"-peer", c2EthAddr,
		"-token", tokenAddr,
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
	fmt.Println("-------------------------------------- list all channel details")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "channel",
		"-list",
		"-detail",
		"-token", tokenAddr,
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
	fmt.Println("-------------------------------------- list inactive channels")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "channel",
		"-list",
		"-detail",
		"-token", tokenAddr,
		"-inactivesec", fmt.Sprintf("%d", int(time.Now().Sub(ts).Seconds())),
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
	fmt.Println("-------------------------------------- list inactive channels")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "channel",
		"-list",
		"-detail",
		"-token", tokenAddr,
		"-inactivesec", "1",
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
	fmt.Println("-------------------------------------- list channel IDs")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "channel",
		"-token", tokenAddr,
		"-list",
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
	fmt.Println("-------------------------------------- balance")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "channel",
		"-token", tokenAddr,
		"-balance",
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
	fmt.Println("-------------------------------------- route")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "route",
		"-dest", c2EthAddr,
		"-token", tokenAddr,
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
	fmt.Println("-------------------------------------- pay")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "pay",
		"-payid", p2,
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
	fmt.Println("-------------------------------------- pay")
	tf.StartProcess(outRootDir+"ospcli",
		"-profile", noProxyProfile,
		"-storedir", sStoreDir+"/"+ospEthAddr,
		"-dbview", "pay",
		"-payid", p4,
		"-logcolor",
		"-logprefix", "cli").Wait()

	fmt.Println()
}
