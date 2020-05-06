// Copyright 2018-2020 Celer Network

package e2e

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/celer-network/goCeler/app"
	"github.com/celer-network/goCeler/common/cobj"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/testing/testapp"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
)

const oracleKeyStore = "../../testing/env/keystore/etherbase.json"

func disputeEthPayBySigTimeoutWithDeployedContract(t *testing.T) {
	log.Info("============== start test disputeEthPayBySigTimeoutWithDeployedContract ==============")
	defer log.Info("============== end test disputeEthPayBySigTimeoutWithDeployedContract ==============")
	t.Parallel()
	disputePayBySigTimeoutWithDeployedContract(t, entity.TokenType_ETH, tokenAddrEth)
}

func disputePayBySigTimeoutWithDeployedContract(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for disputePayBySigTimeoutWithDeployedContract token", tokenAddr, addrs)

	if tokenAddr != tokenAddrEth {
		err = tf.FundAccountsWithErc20(tokenAddr, addrs, accountBalance)
		if err != nil {
			t.Error(err)
			return
		}
	}

	c1, err := tf.StartC1WithoutProxy(ks[0])
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	c2, err := tf.StartC2WithoutProxy(ks[1])
	if err != nil {
		t.Error(err)
		return
	}
	defer c2.Kill()

	players := []string{addrs[0], addrs[1]}
	appChanID, err := openChannel(c1, players[0], tokenType, tokenAddr, players)
	if err != nil {
		t.Error(err)
		return
	}
	appChanID1, err := openChannel(c2, players[1], tokenType, tokenAddr, players)
	if err != nil {
		t.Error(err)
		return
	}
	if appChanID != appChanID1 {
		err = fmt.Errorf("appChanID does not match")
		if err != nil {
			t.Error(err)
			return
		}
	}

	payID, err := sendPayment(c1, appChanID, tokenType, tokenAddr, players[1])
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
	sleep(1)

	log.Info("============ Test generic channel onchain API =============")
	state := 2
	appState := testapp.GetAppStateWithOracle(1, uint8(state), testapp.Nonce.Uint64())
	c1Sig, _ := c1.SignData(appState)
	stateProof := &app.StateProof{AppState: appState}
	stateProof.Sigs = append(stateProof.Sigs, c1Sig)
	serializedStateProof, err := proto.Marshal(stateProof)
	if err != nil {
		t.Error(err)
		return
	}
	serializedOracleProof, err := getOracleProofBytes(serializedStateProof, players, players[0])
	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)
	go tf.AdvanceBlocksUntilDone(done)

	err = c1.SettleAppChannelBySigTimeout(appChanID, serializedOracleProof)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkSessionState(c1, appChanID, state)
	if err != nil {
		t.Error(err)
		return
	}

	err = settlePay(c2, payID)
	if err != nil {
		t.Error(err)
		return
	}
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

func openChannel(client *tf.ClientController, address string, tokenType entity.TokenType, tokenAddr string, players []string) (string, error) {
	_, err := client.OpenChannel(address, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		return "", err
	}

	err = client.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		return "", err
	}

	appChanID, err := client.NewAppChannelOnDeployedContract(
		testapp.ContractWithOracleAddr,
		testapp.Nonce.Uint64(),
		players,
		testapp.Timeout.Uint64())

	if err != nil {
		return "", err
	}

	return appChanID, nil
}

func sendPayment(client *tf.ClientController, appChanID string, tokenType entity.TokenType, tokenAddr string, peerAddress string) (string, error) {
	sessionQuery := &app.SessionQuery{
		Session: ctype.Hex2Bytes(appChanID),
		Query:   []byte{2},
	}
	serializedSessionQuery, err := proto.Marshal(sessionQuery)
	c1Cond := &entity.Condition{
		ConditionType:           entity.ConditionType_DEPLOYED_CONTRACT,
		DeployedContractAddress: ctype.Hex2Bytes(testapp.ContractWithOracleAddr),
		ArgsQueryFinalization:   ctype.Hex2Bytes(appChanID),
		ArgsQueryOutcome:        serializedSessionQuery,
	}

	currentTime, err := client.GetCurrentBlockNumber()
	if err != nil {
		return "", err
	}

	payID, err := client.SendPaymentWithBooleanConditions(
		peerAddress, sendAmt, tokenType, tokenAddr, []*entity.Condition{c1Cond}, currentTime+100)
	if err != nil {
		return "", err
	}
	sleep(1)

	return payID, nil
}

func getOracleProofBytes(stateProofBytes []byte, players []string, updater string) ([]byte, error) {
	oracleState := &app.OracleState{
		StateProof:  stateProofBytes,
		Updater:     ctype.Hex2Addr(updater).Bytes(),
		UpdateTime:  1,
		CurrentTime: 4,
		Players:     getSortedPlayers(players),
	}
	serializedOracleState, err := proto.Marshal(oracleState)
	if err != nil {
		return nil, err
	}

	oracleKsBytes, _ := ioutil.ReadFile(oracleKeyStore)
	key, _ := keystore.DecryptKey(oracleKsBytes, "")
	privKey := hex.EncodeToString(crypto.FromECDSA(key.PrivateKey))
	oracle, err := cobj.NewCelerSigner(privKey)
	if err != nil {
		return nil, err
	}
	oracleSig, _ := oracle.SignEthMessage(serializedOracleState)
	oracleProof := &app.OracleProof{
		OracleState: serializedOracleState,
		Sig:         oracleSig,
	}
	serializedOracleProof, err := proto.Marshal(oracleProof)
	if err != nil {
		return nil, err
	}

	return serializedOracleProof, err
}

func checkSessionState(client *tf.ClientController, appChanID string, expectedState int) error {
	state, err := client.GetAppChannelState(appChanID, 0)
	if err != nil {
		return err
	}
	if state == nil || big.NewInt(0).SetBytes(state).Cmp(big.NewInt(int64(expectedState))) != 0 {
		err = fmt.Errorf("incorrect state %x", state)
		if err != nil {
			return err
		}
	}
	finalized, result, err := client.GetAppChannelBooleanOutcome(appChanID, []byte{byte(expectedState)})
	if err != nil {
		return err
	}
	if !finalized {
		err = fmt.Errorf("condition not finalized")
		if err != nil {
			return err
		}
	}
	if !result {
		err = fmt.Errorf("result not satisfied")
		if err != nil {
			return err
		}
	}
	sleep(1)

	return nil
}

func settlePay(client *tf.ClientController, payID string) error {
	amount, _, err := client.SettleConditionalPayOnChain(payID)
	if err != nil {
		return err
	}
	if amount != "1" {
		err = fmt.Errorf("pay result not match. expect 1 got %s", amount)
		if err != nil {
			return err
		}
	}
	err = client.SettleOnChainResolvedPay(payID)
	if err != nil {
		return err
	}
	sleep(1)

	return nil
}

func getSortedPlayers(players []string) [][]byte {
	sortedPlayers := app.SortPlayers([]ctype.Addr{
		ctype.Hex2Addr(players[0]),
		ctype.Hex2Addr(players[1]),
	})

	return [][]byte{
		sortedPlayers[0].Bytes(),
		sortedPlayers[1].Bytes(),
	}
}
