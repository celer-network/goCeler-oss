// Copyright 2018-2020 Celer Network

package e2e

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/celer-network/goCeler/app"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/testing/testapp"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
)

func disputeEthPayWithVirtualContract(t *testing.T) {
	log.Info("============== start test disputeEthPayWithVirtualContract ==============")
	defer log.Info("============== end test disputeEthPayWithVirtualContract ==============")
	t.Parallel()
	disputePayWithVirtualContract(t, entity.TokenType_ETH, tokenAddrEth)
}

func disputeEthPayWithDeployedContract(t *testing.T) {
	log.Info("============== start test disputeEthPayWithDeployedContract ==============")
	defer log.Info("============== end test disputeEthPayWithDeployedContract ==============")
	t.Parallel()
	disputePayWithDeployedContract(t, entity.TokenType_ETH, tokenAddrEth)
}

func disputeEthPayWithDeployedGomoku(t *testing.T) {
	log.Info("============== start test disputeEthPayWithDeployedGomoku ==============")
	defer log.Info("============== end test disputeEthPayWithDeployedGomoku ==============")
	t.Parallel()
	disputePayWithDeployedGomoku(t, entity.TokenType_ETH, tokenAddrEth)
}

func disputeEthPaySrcOffline(t *testing.T) {
	log.Info("============== start test disputeEthPaySrcOffline ==============")
	defer log.Info("============== end test disputeEthPaySrcOffline ==============")
	t.Parallel()
	disputePaySrcOffline(t, entity.TokenType_ETH, tokenAddrEth)
}

func disputePayWithVirtualContract(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for disputePayWithVirtualContract token", tokenAddr, addrs)
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

	appChanID2, err := c2.NewAppChannelOnVirtualContract(
		testapp.AppCode,
		constructor,
		testapp.Nonce.Uint64(),
		testapp.Timeout.Uint64())
	if err != nil {
		t.Error(err)
		return
	}
	if appChanID != appChanID2 {
		err = fmt.Errorf("appChanID does not match")
		if err != nil {
			t.Error(err)
			return
		}
	}

	c1Cond := &entity.Condition{
		ConditionType:          entity.ConditionType_VIRTUAL_CONTRACT,
		VirtualContractAddress: ctype.Hex2Bytes(appChanID),
		ArgsQueryFinalization:  []byte{},
		ArgsQueryOutcome:       []byte{2},
	}

	payID, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{c1Cond}, 100)
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

	appState := testapp.GetAppState(2, testapp.Nonce.Uint64())
	c1Sig, _ := c1.SignData(appState)
	c2Sig, _ := c2.SignData(appState)
	stateProof := &app.StateProof{AppState: appState}
	stateProof.Sigs = append(stateProof.Sigs, c1Sig)
	stateProof.Sigs = append(stateProof.Sigs, c2Sig)
	serializedStateProof, err := proto.Marshal(stateProof)
	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)
	go tf.AdvanceBlocksUntilDone(done)

	err = c2.SettleAppChannel(appChanID, serializedStateProof)
	if err != nil {
		t.Error(err)
		return
	}

	finalized, result, err := c2.GetAppChannelBooleanOutcome(appChanID, []byte{2})
	if err != nil {
		t.Error(err)
		return
	}
	if !finalized {
		err = fmt.Errorf("condition not finalized")
		if err != nil {
			t.Error(err)
			return
		}
	}
	if !result {
		err = fmt.Errorf("result not satisfied")
		if err != nil {
			t.Error(err)
			return
		}
	}
	sleep(1)

	amount, _, err := c2.SettleConditionalPayOnChain(payID)
	if err != nil {
		t.Error(err)
		return
	}
	if amount != "1" {
		err = fmt.Errorf("pay result not match. expect 1 got %s", amount)
		if err != nil {
			t.Error(err)
			return
		}
	}

	err = c2.SettleOnChainResolvedPay(payID)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(1)
	done <- true

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

func disputePayWithDeployedContract(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for disputePayWithDeployedContract token", tokenAddr, addrs)
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

	players := []string{c1EthAddr, c2EthAddr}
	appChanID, err := c1.NewAppChannelOnDeployedContract(
		testapp.ContractAddr,
		testapp.Nonce.Uint64(),
		players,
		testapp.Timeout.Uint64())
	if err != nil {
		t.Error(err)
		return
	}

	appChanID2, err := c2.NewAppChannelOnDeployedContract(
		testapp.ContractAddr,
		testapp.Nonce.Uint64(),
		players,
		testapp.Timeout.Uint64())
	if err != nil {
		t.Error(err)
		return
	}
	if appChanID != appChanID2 {
		err = fmt.Errorf("appChanID does not match")
		if err != nil {
			t.Error(err)
			return
		}
	}

	sessionQuery := &app.SessionQuery{
		Session: ctype.Hex2Bytes(appChanID),
		Query:   []byte{2},
	}
	serializedSessionQuery, err := proto.Marshal(sessionQuery)
	c1Cond := &entity.Condition{
		ConditionType:           entity.ConditionType_DEPLOYED_CONTRACT,
		DeployedContractAddress: ctype.Hex2Bytes(testapp.ContractAddr),
		ArgsQueryFinalization:   ctype.Hex2Bytes(appChanID),
		ArgsQueryOutcome:        serializedSessionQuery,
	}
	payID, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{c1Cond}, 100)
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

	log.Info("============ Test generic channel onchain API =============")
	appState := testapp.GetAppState(3, testapp.Nonce.Uint64())
	c1Sig, _ := c1.SignData(appState)
	c2Sig, _ := c2.SignData(appState)
	stateProof := &app.StateProof{AppState: appState}
	stateProof.Sigs = append(stateProof.Sigs, c1Sig)
	stateProof.Sigs = append(stateProof.Sigs, c2Sig)
	serializedStateProof, err := proto.Marshal(stateProof)
	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)
	go tf.AdvanceBlocksUntilDone(done)

	err = c2.SettleAppChannel(appChanID, serializedStateProof)
	if err != nil {
		t.Error(err)
		return
	}

	state, err := c2.GetAppChannelState(appChanID, 0)
	if err != nil {
		t.Error(err)
		return
	}
	if state == nil || big.NewInt(0).SetBytes(state).Cmp(big.NewInt(3)) != 0 {
		err = fmt.Errorf("incorrect state %x", state)
		if err != nil {
			t.Error(err)
			return
		}
	}

	finalizedTime, err := c1.GetAppChannelSettleFinalizedTime(appChanID)
	if err != nil {
		t.Error(err)
		return
	}
	err = c1.WaitUntilDeadline(finalizedTime)
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.ApplyAppChannelAction(appChanID, []byte{2})
	if err != nil {
		t.Error(err)
		return
	}

	finalized, result, err := c2.GetAppChannelBooleanOutcome(appChanID, []byte{2})
	if err != nil {
		t.Error(err)
		return
	}
	if !finalized {
		err = fmt.Errorf("condition not finalized")
		if err != nil {
			t.Error(err)
			return
		}
	}
	if !result {
		err = fmt.Errorf("result not satisfied")
		if err != nil {
			t.Error(err)
			return
		}
	}

	amount, _, err := c2.SettleConditionalPayOnChain(payID)
	if err != nil {
		t.Error(err)
		return
	}
	if amount != "1" {
		err = fmt.Errorf("pay result not match. expect 1 got %s", amount)
		if err != nil {
			t.Error(err)
			return
		}
	}

	err = c2.SettleOnChainResolvedPay(payID)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(1)
	done <- true

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

func disputePaySrcOffline(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for disputePaySrcOffline token", tokenAddr, addrs)
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

	players := []string{c1EthAddr, c2EthAddr}

	log.Infoln("==================  First cond pay and app settle =======================")
	appChanID, err := c1.NewAppChannelOnDeployedContract(
		testapp.ContractAddr,
		666,
		players,
		testapp.Timeout.Uint64())
	if err != nil {
		t.Error(err)
		return
	}

	sessionQuery := &app.SessionQuery{
		Session: ctype.Hex2Bytes(appChanID),
		Query:   []byte{2},
	}
	serializedSessionQuery, err := proto.Marshal(sessionQuery)
	cond := &entity.Condition{
		ConditionType:           entity.ConditionType_DEPLOYED_CONTRACT,
		DeployedContractAddress: ctype.Hex2Bytes(testapp.ContractAddr),
		ArgsQueryFinalization:   ctype.Hex2Bytes(appChanID),
		ArgsQueryOutcome:        serializedSessionQuery,
	}
	payID1, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{cond}, 100)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentPending(payID1, c1, c2)
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

	appState := testapp.GetAppState(2, 666)
	c1Sig, _ := c1.SignData(appState)
	c2Sig, _ := c2.SignData(appState)
	stateProof := &app.StateProof{AppState: appState}
	stateProof.Sigs = append(stateProof.Sigs, c1Sig)
	stateProof.Sigs = append(stateProof.Sigs, c2Sig)
	serializedStateProof, err := proto.Marshal(stateProof)
	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)
	go tf.AdvanceBlocksUntilDone(done)

	err = c1.SettleAppChannel(appChanID, serializedStateProof)
	if err != nil {
		t.Error(err)
		return
	}
	done <- true

	finalized, result, err := c1.GetAppChannelBooleanOutcome(appChanID, []byte{2})
	if err != nil {
		t.Error(err)
		return
	}
	if !finalized {
		err = fmt.Errorf("condition not finalized")
		if err != nil {
			t.Error(err)
			return
		}
	}
	if !result {
		err = fmt.Errorf("result not satisfied")
		if err != nil {
			t.Error(err)
			return
		}
	}
	c1.DeleteAppChannel(appChanID)

	log.Infoln("==================  Second cond pay and app settle =======================")
	appChanID, err = c1.NewAppChannelOnDeployedContract(
		testapp.ContractAddr,
		999,
		players,
		testapp.Timeout.Uint64())
	if err != nil {
		t.Error(err)
		return
	}

	sessionQuery = &app.SessionQuery{
		Session: ctype.Hex2Bytes(appChanID),
		Query:   []byte{2},
	}
	serializedSessionQuery, err = proto.Marshal(sessionQuery)
	cond = &entity.Condition{
		ConditionType:           entity.ConditionType_DEPLOYED_CONTRACT,
		DeployedContractAddress: ctype.Hex2Bytes(testapp.ContractAddr),
		ArgsQueryFinalization:   ctype.Hex2Bytes(appChanID),
		ArgsQueryOutcome:        serializedSessionQuery,
	}
	payID2, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{cond}, 100)
	if err != nil {
		t.Error(err)
		return
	}

	err = waitForPaymentPending(payID2, c1, c2)
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

	appState = testapp.GetAppState(2, 999)
	c1Sig, _ = c1.SignData(appState)
	c2Sig, _ = c2.SignData(appState)
	stateProof = &app.StateProof{AppState: appState}
	stateProof.Sigs = append(stateProof.Sigs, c1Sig)
	stateProof.Sigs = append(stateProof.Sigs, c2Sig)
	serializedStateProof, err = proto.Marshal(stateProof)
	if err != nil {
		t.Error(err)
		return
	}

	go tf.AdvanceBlocksUntilDone(done)

	err = c1.SettleAppChannel(appChanID, serializedStateProof)
	if err != nil {
		t.Error(err)
		return
	}
	done <- true

	finalized, result, err = c1.GetAppChannelBooleanOutcome(appChanID, []byte{2})
	if err != nil {
		t.Error(err)
		return
	}
	if !finalized {
		err = fmt.Errorf("condition not finalized")
		if err != nil {
			t.Error(err)
			return
		}
	}
	if !result {
		err = fmt.Errorf("result not satisfied")
		if err != nil {
			t.Error(err)
			return
		}
	}
	sleep(1)
	c1.DeleteAppChannel(appChanID)

	log.Infoln("==================  Resolve pay onchain =======================")
	amount, _, err := c1.SettleConditionalPayOnChain(payID1)
	if err != nil {
		t.Error(err)
		return
	}
	if amount != "1" {
		err = fmt.Errorf("pay result not match. expect 1 got %s", amount)
		if err != nil {
			t.Error(err)
			return
		}
	}

	amount, _, err = c1.SettleConditionalPayOnChain(payID2)
	if err != nil {
		t.Error(err)
		return
	}
	if amount != "1" {
		err = fmt.Errorf("pay result not match. expect 1 got %s", amount)
		if err != nil {
			t.Error(err)
			return
		}
	}

	log.Infoln("==================  Kill C1 =======================")
	c1.KillWithoutRemovingKeystore()

	log.Infoln("==================  Settle onchain resolved pays =======================")
	err = c2.SettleOnChainResolvedPay(payID1)
	if err != nil {
		t.Error(err)
		return
	}
	err = c2.SettleOnChainResolvedPay(payID2)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(7)

	log.Infoln("================== Restart C1 =======================")
	c1, err = tf.StartC1WithoutProxy(c1KeyStore)
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

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
		tf.AddAmtStr(initialBalance, "2"),
		"0",
		tf.AddAmtStr(initialBalance, "-2"))
	if err != nil {
		t.Error(err)
		return
	}

	err = c1.ConfirmOnChainResolvedPays(tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(1)
	err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-2"),
		"0",
		tf.AddAmtStr(initialBalance, "2"))
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

func disputePayWithDeployedGomoku(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for disputePayWithDeployedGomoku token", tokenAddr, addrs)
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
	if bytes.Compare(ctype.Hex2Bytes(addrs[0]), ctype.Hex2Bytes(addrs[1])) == 1 {
		c1KeyStore = ks[1]
		c2KeyStore = ks[0]
		c1EthAddr = addrs[1]
		c2EthAddr = addrs[0]
	}

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

	players := []string{c1EthAddr, c2EthAddr}
	appChanID, err := c1.NewAppChannelOnDeployedContract(
		testapp.GomokuAddr,
		testapp.Nonce.Uint64(),
		players,
		testapp.Timeout.Uint64())
	if err != nil {
		t.Error(err)
		return
	}

	appChanID2, err := c2.NewAppChannelOnDeployedContract(
		testapp.GomokuAddr,
		testapp.Nonce.Uint64(),
		players,
		testapp.Timeout.Uint64())
	if err != nil {
		t.Error(err)
		return
	}
	if appChanID != appChanID2 {
		err = fmt.Errorf("appChanID does not match")
		if err != nil {
			t.Error(err)
			return
		}
	}

	sessionQuery := &app.SessionQuery{
		Session: ctype.Hex2Bytes(appChanID),
		Query:   []byte{2},
	}
	serializedSessionQuery, err := proto.Marshal(sessionQuery)
	c1Cond := &entity.Condition{
		ConditionType:           entity.ConditionType_DEPLOYED_CONTRACT,
		DeployedContractAddress: ctype.Hex2Bytes(testapp.GomokuAddr),
		ArgsQueryFinalization:   ctype.Hex2Bytes(appChanID),
		ArgsQueryOutcome:        serializedSessionQuery,
	}
	currentTime, err := c1.GetCurrentBlockNumber()
	if err != nil {
		t.Error(err)
		return
	}
	payID, err := c1.SendPaymentWithBooleanConditions(
		c2EthAddr, sendAmt, tokenType, tokenAddr, []*entity.Condition{c1Cond}, currentTime+100)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(1)

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
	sleep(1)

	log.Info("============ Test generic channel onchain API =============")
	appState := testapp.GetGomokuState()
	c1Sig, _ := c1.SignData(appState)
	c2Sig, _ := c2.SignData(appState)
	stateProof := &app.StateProof{AppState: appState}
	stateProof.Sigs = append(stateProof.Sigs, c1Sig)
	stateProof.Sigs = append(stateProof.Sigs, c2Sig)
	serializedStateProof, err := proto.Marshal(stateProof)
	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)
	go tf.AdvanceBlocksUntilDone(done)

	err = c2.SettleAppChannel(appChanID, serializedStateProof)
	if err != nil {
		t.Error(err)
		return
	}

	state, err := c2.GetAppChannelState(appChanID, 2)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(state, testapp.GetGetGomokuBoardState()) {
		err = fmt.Errorf("incorrect state %x", state)
		if err != nil {
			t.Error(err)
			return
		}
	}

	finalizedTime, err := c2.GetAppChannelSettleFinalizedTime(appChanID)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.WaitUntilDeadline(finalizedTime)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.ApplyAppChannelAction(appChanID, []byte{5, 5})
	if err != nil {
		t.Error(err)
		return
	}

	actionDeadline, err := c2.GetAppChannelActionDeadline(appChanID)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.WaitUntilDeadline(actionDeadline)
	if err != nil {
		t.Error(err)
		return
	}

	err = c2.FinalizeAppChannelOnActionTimeout(appChanID)
	if err != nil {
		t.Error(err)
		return
	}

	finalized, result, err := c2.GetAppChannelBooleanOutcome(appChanID, []byte{2})
	if err != nil {
		t.Error(err)
		return
	}
	if !finalized {
		err = fmt.Errorf("condition not finalized")
		if err != nil {
			t.Error(err)
			return
		}
	}
	if !result {
		err = fmt.Errorf("result not satisfied")
		if err != nil {
			t.Error(err)
			return
		}
	}
	sleep(1)

	amount, _, err := c2.SettleConditionalPayOnChain(payID)
	if err != nil {
		t.Error(err)
		return
	}
	if amount != "1" {
		err = fmt.Errorf("pay result not match. expect 1 got %s", amount)
		if err != nil {
			t.Error(err)
			return
		}
	}

	err = c2.SettleOnChainResolvedPay(payID)
	if err != nil {
		t.Error(err)
		return
	}
	sleep(1)
	done <- true

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
