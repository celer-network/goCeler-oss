// Copyright 2018-2020 Celer Network

package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

func authSync(t *testing.T) {
	log.Info("============== start test authSync ==============")
	defer log.Info("============== end test authSync ==============")
	t.Parallel()
	authsynctest(t)
}

// open eth channel and erc20 tcb
func openchannel(c *tf.ClientController, eth string) error {
	_, err := c.OpenChannel(eth, entity.TokenType_ETH, tokenAddrEth, initialBalance, initialBalance)
	if err != nil {
		return err
	}
	_, err = c.TcbOpenChannel(eth, entity.TokenType_ERC20, tokenAddrErc20, initialBalance)
	return err
}

// fsm pay state from db directly
func checkOutPayState(c *tf.ClientController, payid string, exp int) bool {
	b, err := c.RunSQL(fmt.Sprintf(`SELECT outstate from payments where payid = "%s"`, payid))
	if err != nil {
		log.Error("runsql err: ", err)
		return false
	}
	ss := strings.Trim(string(b), "\n")
	s, _ := strconv.Atoi(ss)
	if s != exp {
		log.Errorf("pay %s state %s exp %s", payid, fsm.PayStateName(s), fsm.PayStateName(exp))
	}
	return s == exp
}

func startNewC1(c1KeyStore string) (*tf.ClientController, error) {
	return tf.StartClientController(
		tf.GetNextClientPort(), c1KeyStore, outRootDir+"profile.json", outRootDir+"c1Store2", "c1")
}

func authsynctest(t *testing.T) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for authsynctest", addrs)

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

	err = openchannel(c1, c1EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	err = openchannel(c2, c2EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(1)
	log.Info("🔥🔥🔥🔥🔥 authsync test: seq 0 wipe data 🔥🔥🔥🔥🔥")
	c1.KillAndRemoveDB()
	sleep(1)
	c1, err = tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()
	err = c1.AssertBalance(tokenAddrEth, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	// fund c1 2wei erc20
	requestSvrSendToken(c1EthAddr, "1", tokenAddrErc20)
	requestSvrSendToken(c1EthAddr, "9", tokenAddrErc20)

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr("10"),
		"0",
		tf.AddAmtStr(initialBalance, "-10"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("🔥🔥🔥🔥🔥 authsync test: lost resp 🔥🔥🔥🔥🔥")
	// drop recv so cosign lost
	c1.SetMsgDropper(true, false)
	// direct pays
	p1, err := c1.SendPayment(ospEthAddr, "1", entity.TokenType_ETH, tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	p2, err := c1.SendPayment(ospEthAddr, "1", entity.TokenType_ERC20, tokenAddrErc20)
	if err != nil {
		t.Error(err)
		return
	}
	// with condition
	p3, err := c1.SendPayment(c2EthAddr, "1", entity.TokenType_ERC20, tokenAddrErc20)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(1) // wait long enough for normal msg exchange
	// low level pay state should be onesig paid for direct pay and onsig pending for p3
	if !checkOutPayState(c1, p1, structs.PayState_ONESIG_PAID) {
		t.Error("wrong paystate. payid:", p1)
		return
	}
	if !checkOutPayState(c1, p2, structs.PayState_ONESIG_PAID) {
		t.Error("wrong paystate. payid:", p2)
		return
	}
	if !checkOutPayState(c1, p3, structs.PayState_ONESIG_PENDING) {
		t.Error("wrong paystate. payid:", p3)
		return
	}
	// restart c1 so cosign is synced
	c1.Kill()
	c1, err = tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()
	// now both pays should be paid
	if !checkOutPayState(c1, p1, structs.PayState_COSIGNED_PAID) {
		t.Error("wrong paystate. payid:", p1)
		return
	}
	if !checkOutPayState(c1, p2, structs.PayState_COSIGNED_PAID) {
		t.Error("wrong paystate. payid:", p2)
		return
	}
	if !checkOutPayState(c1, p3, structs.PayState_COSIGNED_PENDING) {
		t.Error("wrong paystate. payid:", p3)
		return
	}

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(initialBalance, "-1"),
		"0",
		tf.AddAmtStr(initialBalance, "1"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr("10", "-2"),
		"1",
		tf.AddAmtStr(initialBalance, "-10", "1"))
	if err != nil {
		t.Error(err)
		return
	}

	c1.KillWithoutRemovingKeystore()

	log.Info("------------ restart c1 on device 2 (using a different storeDir) --------")
	c1, err = startNewC1(c1KeyStore)
	defer c1.Kill()

	p4, err := c1.SendPayment(ospEthAddr, "1", entity.TokenType_ETH, tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	err = waitForPaymentCompletion(p4, c1, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if !checkOutPayState(c1, p4, structs.PayState_COSIGNED_PAID) {
		t.Error("wrong paystate. payid: ", p4)
		return
	}

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(initialBalance, "-1", "-1"),
		"0",
		tf.AddAmtStr(initialBalance, "1", "1"))
	if err != nil {
		t.Error(err)
		return
	}

	requestSvrSendToken(c1EthAddr, "1", tokenAddrErc20)

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr("10", "-2", "1"),
		"1",
		tf.AddAmtStr(initialBalance, "-10", "1", "-1"))
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("------------ set c1 to drop incomoing and outgoing payments --------")
	c1.SetMsgDropper(true, true)

	requestSvrSendToken(c1EthAddr, "10", tokenAddrEth)
	requestSvrSendToken(c1EthAddr, "10", tokenAddrEth)

	p5, err := c1.SendPayment(ospEthAddr, "99", entity.TokenType_ETH, tokenAddrEth)
	log.Infoln("send p5", p5, "which should be eventually canceled")
	if !checkOutPayState(c1, p5, structs.PayState_ONESIG_PAID) {
		t.Error("wrong paystate. payid: ", p5)
		return
	}

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(initialBalance, "-1", "-1", "-99"),
		"0",
		tf.AddAmtStr(initialBalance, "1", "1", "99"))
	if err != nil {
		t.Error(err)
		return
	}

	requestSvrSendToken(c1EthAddr, "1", tokenAddrErc20)

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr("10", "-2", "1"),
		"1",
		tf.AddAmtStr(initialBalance, "-10", "1", "-1"))
	if err != nil {
		t.Error(err)
		return
	}

	c1.KillWithoutRemovingKeystore()

	log.Info("------------ restart c1 on device 1 (using the default storeDir) --------")
	c1, err = tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(initialBalance, "-1", "-1", "20"),
		"0",
		tf.AddAmtStr(initialBalance, "1", "1", "-20"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr("10", "-2", "1", "1"),
		"1",
		tf.AddAmtStr(initialBalance, "-10", "1", "-1", "-1"))
	if err != nil {
		t.Error(err)
		return
	}

	requestSvrSendToken(c1EthAddr, "2", tokenAddrEth)

	p6, err := c1.SendPayment(ospEthAddr, "5", entity.TokenType_ETH, tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	err = waitForPaymentCompletion(p6, c1, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if !checkOutPayState(c1, p6, structs.PayState_COSIGNED_PAID) {
		t.Error("wrong paystate. payid: ", p6)
		return
	}

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(initialBalance, "-1", "-1", "20", "2", "-5"),
		"0",
		tf.AddAmtStr(initialBalance, "1", "1", "-20", "-2", "5"))
	if err != nil {
		t.Error(err)
		return
	}

	p7, err := c1.SendPayment(ospEthAddr, "1", entity.TokenType_ERC20, tokenAddrErc20)
	if err != nil {
		t.Error(err)
		return
	}
	err = waitForPaymentCompletion(p7, c1, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if !checkOutPayState(c1, p7, structs.PayState_COSIGNED_PAID) {
		t.Error("wrong paystate. payid: ", p7)
		return
	}

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr("10", "-2", "1", "1", "-1"),
		"1",
		tf.AddAmtStr(initialBalance, "-10", "1", "-1", "-1", "1"))
	if err != nil {
		t.Error(err)
		return
	}
	c1.KillWithoutRemovingKeystore()

	log.Info("------------ restart c1 on device 2 (using the different storeDir) --------")
	c1, err = startNewC1(c1KeyStore)
	defer c1.Kill()

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(initialBalance, "-1", "-1", "20", "2", "-5"),
		"0",
		tf.AddAmtStr(initialBalance, "1", "1", "-20", "-2", "5"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.AssertBalance(
		tokenAddrErc20,
		tf.AddAmtStr("10", "-2", "1", "1", "-1"),
		"1",
		tf.AddAmtStr(initialBalance, "-10", "1", "-1", "-1", "1"))
	if err != nil {
		t.Error(err)
		return
	}

	// NOTE: p5 should be COSIGNED_CANCELED by the current stage, but is now set to COSIGNED_PAID due to
	// current implementation limitation. see the WARNING comment at the begining of cnode/auth.go

	p8, err := c1.SendPayment(ospEthAddr, "10", entity.TokenType_ETH, tokenAddrEth)
	if err != nil {
		t.Error(err)
		return
	}
	err = waitForPaymentCompletion(p8, c1, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if !checkOutPayState(c1, p8, structs.PayState_COSIGNED_PAID) {
		t.Error("wrong paystate. payid: ", p8)
		return
	}

	err = c1.AssertBalance(
		tokenAddrEth,
		tf.AddAmtStr(initialBalance, "-1", "-1", "20", "2", "-5", "-10"),
		"0",
		tf.AddAmtStr(initialBalance, "1", "1", "-20", "-2", "5", "10"))
	if err != nil {
		t.Error(err)
		return
	}
}
