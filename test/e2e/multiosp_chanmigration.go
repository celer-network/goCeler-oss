// Copyright 2020 Celer Network

package e2e

import (
	"path/filepath"
	"testing"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/tools/toolsetup"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

func migrateChannelBetweenOsps(args ...*tf.ServerController) func(*testing.T) {
	return func(t *testing.T) {
		log.Info("============== start test migrateChannelBetweenOsps ==============")
		defer log.Info("============== end test migrateChannelBetweenOsps ==============")
		o1 := args[0]
		o2 := args[1]
		// Let osp2 initiate openning channel with osp1.
		err := requestOpenChannel(o2AdminWeb, ospEthAddr, initOspToOspBalance, initOspToOspBalance, tokenAddrEth)
		if err != nil {
			log.Warn(err)
		}

		tokenType := entity.TokenType_ETH
		tokenAddr := tokenAddrEth
		ks, addrs, err := tf.CreateAccountsWithBalance(1, accountBalance)
		if err != nil {
			t.Error(err)
			return
		}
		log.Infoln("create accounts for migrateChannelBetweenOsps", addrs)

		c1KeyStore := ks[0]
		c1EthAddr := addrs[0]

		c1, err := tf.StartC1WithoutProxy(c1KeyStore)
		if err != nil {
			t.Error(err)
			return
		}
		defer c1.Kill()

		resp, err := c1.OpenChannel(c1EthAddr, tokenType, tokenAddr, initialBalance, initialBalance)
		if err != nil {
			t.Error(err)
			return
		}
		cid := ctype.Hex2Cid(resp.GetChannelId())

		err = c1.AssertBalance(tokenAddr, initialBalance, "0", initialBalance)
		if err != nil {
			t.Error(err)
			return
		}
		c1.KillWithoutRemovingKeystore()
		sleep(2)

		// stop s1 and replace the config file
		log.Infoln("================================ restart o1 with new profile ==================================")
		o1.Kill()
		o2.Kill()
		sleep(1)

		oldLedger := tf.E2eProfile.Ethereum.Contracts.Ledger
		newLedger := "6666666666666666666666666666666666666666"
		newProfile := *tf.E2eProfile
		newProfile.Ethereum.Contracts.Ledger = newLedger
		newProfile.Ethereum.Contracts.Ledgers = make(map[string]string)
		newProfile.Ethereum.Contracts.Ledgers[oldLedger] = "value doesn't matter now"
		newLedgerProfileFileName := "newledgerprofile.json"
		newLedgerProfilePath := filepath.Join(outRootDir, newLedgerProfileFileName)
		SaveProfile(&newProfile, newLedgerProfilePath)
		o1 = tf.StartServerController(outRootDir+toBuild["server"],
			"-profile", newLedgerProfilePath,
			"-port", o1Port,
			"-storedir", sStoreDir,
			"-ks", ospKeystore,
			"-nopassword",
			"-rtc", rtConfigMultiOSP,
			"-svrname", "o1",
			"-logcolor",
			"-logprefix", "o1_"+ospEthAddr[:4])
		defer o1.Kill()

		log.Infoln("================================ reconnect c1 to o1 ==================================")
		c1, err = tf.StartClientWithoutProxy(c1KeyStore, newLedgerProfileFileName, "c1")
		if err != nil {
			t.Error(err)
			return
		}
		defer c1.Kill()

		log.Infoln("================================ reconnect o2 to o1 ==================================")
		o2 = tf.StartServerController(outRootDir+toBuild["server"],
			"-profile", newLedgerProfilePath,
			"-port", o2Port,
			"-storedir", sStoreDir,
			"-adminrpc", o2AdminRPC,
			"-adminweb", o2AdminWeb,
			"-ks", osp2Keystore,
			"-nopassword",
			"-rtc", rtConfigMultiOSP,
			"-svrname", "o2",
			"-logcolor",
			"-logprefix", "o2_"+osp2EthAddr[:4])
		defer o2.Kill()
		if err = utils.RequestRegisterStream(o2AdminWeb, ctype.Hex2Addr(ospEthAddr), localhost+o1Port); err != nil {
			t.Error(err)
			return
		}
		sleep(2)

		// check database
		profile := &common.CProfile{
			StoreDir: sStoreDir + "/" + osp1EthAddr,
		}
		dal := toolsetup.NewDAL(profile)
		// check channel migration info for client
		_, state, _, found, err := dal.GetChanMigration(cid, ctype.Hex2Addr(newLedger))
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Errorf("Channel migration info not found for channel %x", cid)
			return
		}
		if state != 0 {
			t.Errorf("unexpected migration state: want(%d), get(%d)", 0, state)
			return
		}
		// check channel migration info for o2
		osp2Addr := ctype.Hex2Addr(osp2EthAddr)
		cids, found, err := dal.GetPeerCids(osp2Addr)
		if err != nil {
			t.Error(err)
			return
		}
		if !found {
			t.Errorf("No channels found for peer %x", osp2Addr)
			return
		}
		for _, cid := range cids {
			_, state, _, found, err = dal.GetChanMigration(cid, ctype.Hex2Addr(newLedger))
			if err != nil {
				t.Error(err)
				return
			}
			if !found {
				t.Errorf("Channel migration info not found for channel %x", cid)
				return
			}
			if state != 0 {
				t.Errorf("unexpected migration state: want(%d), get(%d)", 0, state)
				return
			}
		}

		log.Infoln("================================= disconnect c1 and o2 to o1 ================================")
		sleep(1)
		c1.KillWithoutRemovingKeystore()
		o2.Kill()
		sleep(2)

		log.Infoln("================================= reconnect c1 and o2 to o1 ================================")
		c1, err = tf.StartClientWithoutProxy(c1KeyStore, newLedgerProfileFileName, "c1")
		if err != nil {
			t.Error(err)
			return
		}
		defer c1.Kill()

		o2 = tf.StartServerController(outRootDir+toBuild["server"],
			"-profile", newLedgerProfilePath,
			"-port", o2Port,
			"-storedir", sStoreDir,
			"-adminrpc", o2AdminRPC,
			"-adminweb", o2AdminWeb,
			"-ks", osp2Keystore,
			"-nopassword",
			"-rtc", rtConfigMultiOSP,
			"-svrname", "o2",
			"-logcolor",
			"-logprefix", "o2_"+osp2EthAddr[:4])
		defer o2.Kill()
		if err = utils.RequestRegisterStream(o2AdminWeb, ctype.Hex2Addr(ospEthAddr), localhost+o1Port); err != nil {
			t.Error(err)
			return
		}
		sleep(1)
	}
}
