// Copyright 2019-2020 Celer Network

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/celer-network/goCeler/chain/channel-eth-go/wallet"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/tools/toolsetup"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestE2EChannelMigrationTool(t *testing.T) {
	log.Info("============== start test E2EChannelMigrationTool ==============")
	defer log.Info("============== end test E2EChannelMigrationTool ==============")
	buildPkgBin(outRootDir, "tools/channel-migration", "chanmigration")
	// purge database before start testing
	os.RemoveAll(sStoreDir)
	defer os.RemoveAll(sStoreDir)

	o := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", noProxyProfile,
		"-port", sPort,
		"-selfrpc", sSelfRPC,
		"-storedir", sStoreDir,
		"-ks", ospKeystore,
		"-nopassword",
		"-rtc", rtConfig,
		"-svrname", "joe",
		"-logprefix", "o",
		"-logcolor")
	defer o.Kill()

	tokenType := entity.TokenType_ETH
	tokenAddr := tokenAddrEth

	ks, addrs, err := tf.CreateAccountsWithBalance(2, accountBalance)
	if err != nil {
		t.Error(err)
		return
	}
	log.Infoln("create accounts for TestE2EChannelMigrationTool", addrs)

	profileName := "profile.json"

	c1KeyStore := ks[0]
	c1EthAddr := addrs[0]
	c1, err := tf.StartClientWithoutProxy(c1KeyStore, profileName, "c1")
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	c2KeyStore := ks[1]
	c2EthAddr := addrs[1]
	c2, err := tf.StartClientWithoutProxy(c2KeyStore, profileName, "c2")
	if err != nil {
		t.Error(err)
		return
	}
	defer c2.Kill()

	// c1 open channel with osp
	resp, err := c1.OpenChannel(c1EthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	cid1 := ctype.Hex2Cid(resp.GetChannelId())

	err = c1.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	// c2 open channel with osp
	resp, err = c2.OpenChannel(c2EthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
	if err != nil {
		t.Error(err)
		return
	}
	cid2 := ctype.Hex2Cid(resp.GetChannelId())

	err = c2.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
	if err != nil {
		t.Error(err)
		return
	}

	// kill c1, c2 to wait OSP restart with new ledger
	c1.KillWithoutRemovingKeystore()
	c2.KillWithoutRemovingKeystore()
	o.Kill()
	sleep(1)

	log.Infoln("======================== restart osp with new ledger ==========================")

	newProfile := *tf.E2eProfile
	for addr, version := range tf.E2eProfile.Ethereum.Contracts.Ledgers {
		if version == "ledger2" {
			newProfile.Ethereum.Contracts.Ledger = addr
			break
		}
	}
	newLedgerAddr := ctype.Hex2Addr(newProfile.Ethereum.Contracts.Ledger)
	newProfileFileName := "new_ledger_profile.json"
	newProfileFilePath := filepath.Join(outRootDir, newProfileFileName)
	SaveProfile(&newProfile, newProfileFilePath)
	o = tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", newProfileFilePath,
		"-port", sPort,
		"-selfrpc", sSelfRPC,
		"-storedir", sStoreDir,
		"-ks", ospKeystore,
		"-nopassword",
		"-rtc", rtConfig,
		"-svrname", "joe",
		"-logprefix", "o ",
		"-logcolor",
	)
	defer o.Kill()

	log.Infoln("===================== restart clients with new ledger =========================")
	c1, err = tf.StartClientWithoutProxy(c1KeyStore, newProfileFileName, "c1")
	if err != nil {
		t.Error(err)
		return
	}
	defer c1.Kill()

	c2, err = tf.StartClientWithoutProxy(c2KeyStore, newProfileFileName, "c2")
	if err != nil {
		t.Error(err)
		return
	}
	defer c2.Kill()

	// test pay between c1 and c2 works
	p1, err := c1.SendPayment(c2EthAddr, sendAmt, tokenType, tokenAddr)
	if err != nil {
		t.Error(err)
		return
	}

	if err = waitForPaymentCompletion(p1, c1, c2); err != nil {
		t.Error(err)
		return
	}

	if err = c1.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "-1"),
		"0",
		tf.AddAmtStr(initialBalance, "1"),
	); err != nil {
		t.Error(err)
		return
	}

	if err = c2.AssertBalance(
		tokenAddr,
		tf.AddAmtStr(initialBalance, "1"),
		"0",
		tf.AddAmtStr(initialBalance, "-1"),
	); err != nil {
		t.Error(err)
		return
	}

	// keep mining blocks
	done := make(chan bool)
	go tf.AdvanceBlocksUntilDone(done)
	defer func() {
		done <- true
	}()

	// use channel migration tool to migrate channels
	if _, err = tf.StartProcess(outRootDir+"chanmigration",
		"-profile", newProfileFilePath,
		"-ks", ospKeystore,
		"-storedir", sStoreDir+"/"+ospEthAddr,
	).Wait(); err != nil {
		t.Error(err)
		return
	}

	// check the wallet contract about channel operator
	conn, err := ethclient.Dial(tf.E2eProfile.Ethereum.Gateway)
	if err != nil {
		t.Error(err)
		return
	}
	walletContract, err := wallet.NewCelerWalletCaller(ctype.Hex2Addr(tf.E2eProfile.Ethereum.Contracts.Wallet), conn)
	if err != nil {
		t.Error(err)
		return
	}
	ledger, err := walletContract.GetOperator(&bind.CallOpts{}, cid1)
	if err != nil {
		t.Error(err)
		return
	}
	if ledger != newLedgerAddr {
		t.Errorf("unexpected ledger addr of channel(%x): want(%x), get(%x)", cid1, newLedgerAddr, ledger)
		return
	}

	ledger, err = walletContract.GetOperator(&bind.CallOpts{}, cid2)
	if err != nil {
		t.Error(err)
		return
	}
	if ledger != newLedgerAddr {
		t.Errorf("unexpected ledger addr of channel(%x): want(%x), get(%x)", cid2, newLedgerAddr, ledger)
		return
	}
	// check the database of channels' ledger
	profile := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp1EthAddr,
	}

	sleep(2)
	dal := toolsetup.NewDAL(profile)

	ledger, found, err := dal.GetChanLedger(cid1)
	if err != nil {
		t.Error(err)
		return
	}
	if !found {
		t.Errorf("channel %x not found", cid1)
		return
	}
	if ledger != newLedgerAddr {
		t.Errorf("unexpected ledger addr of channel(%x): want(%x), get(%x)", cid1, newLedgerAddr, ledger)
		return
	}

	ledger, found, err = dal.GetChanLedger(cid2)
	if err != nil {
		t.Error(err)
		return
	}
	if !found {
		t.Errorf("channel %x not found", cid2)
		return
	}
	if ledger != newLedgerAddr {
		t.Errorf("unexpected ledger addr of channel(%x): want(%x), get(%x)", cid2, newLedgerAddr, ledger)
		return
	}

	// migration info should be deleted
	_, _, _, found, err = dal.GetChanMigration(cid1, newLedgerAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if found {
		t.Errorf("channel %x migration should not exist", cid1)
		return
	}

	_, _, _, found, err = dal.GetChanMigration(cid2, newLedgerAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if found {
		t.Errorf("channel %x migration should not exist", cid2)
		return
	}
}
