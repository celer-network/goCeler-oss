// Copyright 2018-2020 Celer Network

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/test/e2e"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

var autofund = flag.Bool("autofund", false, "automatically add and approve fund when setup")

const (
	outRootDir = "/tmp/celer_manual_test/"
	profileDir = outRootDir + "profile/"
	storeDir   = outRootDir + "store/" // OSP SQLite store path is `storeDir/ospAddr`

	osp1Addr = "6a6d2a97da1c453a4e099e8054865a0a59728863"
	osp2Addr = "ba756d65a1a03f07d205749f35e2406e4a8522ad"
	osp3Addr = "90506865876ac95f642c8214cc3b570121ba4e4e"
	osp4Addr = "59fb4bc14618ee4e9e83ff563c72b7fd1222731c"
	osp5Addr = "cc09df4fe5e09ff9dec5781170f146990dd77ba3"

	// config of test OSP 1
	o1Port      = "10001"
	o1AdminHttp = "localhost:8190"
	// config of test OSP 2
	o2Port      = "10002"
	o2AdminHttp = "localhost:8290"
	// config of test OSP 3
	o3Port      = "10003"
	o3AdminHttp = "localhost:8390"
	// config of test OSP 4
	o4Port      = "10004"
	o4AdminHttp = "localhost:8490"
	// config of test OSP 5
	o5Port      = "10005"
	o5AdminHttp = "localhost:8590"

	TestTokenAddr = "f3ccc0a86f8451ab193011fbb408db2e38eaf10a" // test ERC20 token
	SimpleAppAddr = "58712219a4bdbb0e581dcaf6f5c4c2b2d2f42158" // multi-session simple app
	GomokuAppAddr = "4e4a0101cd72258183586a51f8254e871b9c544a" // multi-session gomoku app

	osp1Keystore = "../../testing/env/keystore/osp1.json"
	osp2Keystore = "../../testing/env/keystore/osp2.json"
	osp3Keystore = "../../testing/env/keystore/osp3.json"
	osp4Keystore = "../../testing/env/keystore/osp4.json"
	osp5Keystore = "../../testing/env/keystore/osp5.json"
)

func main() {
	flag.Parse()
	// mkdir out root
	err := os.MkdirAll(outRootDir, os.ModePerm)
	e2e.CheckError(err, "creating root dir")
	fmt.Println("Using folder:", outRootDir)
	os.MkdirAll(profileDir, os.ModePerm)
	os.MkdirAll(storeDir, os.ModePerm)
	goCelerDir := os.Getenv("GOCELER") + "/"
	tf.SetEnvDir(goCelerDir + "testing/env/")
	tf.SetOutRootDir(outRootDir)
	e2e.SetEnvDir(goCelerDir + "testing/env/")
	e2e.SetOutRootDir(outRootDir)
	ethProc, err := e2e.StartChain()
	defer ethProc.Kill()
	time.Sleep(3 * time.Second)
	e2e.CheckError(err, "starting chain")

	tf.E2eProfile, _ = e2e.SetupOnChain(make(map[string]ctype.Addr), *autofund)
	if *autofund {
		// if auto fund, also register all osps on-chain as routers
		tf.RegisterRouters([]string{osp1Keystore, osp2Keystore, osp3Keystore, osp4Keystore, osp5Keystore})
	}

	profile := *tf.E2eProfile
	profile.Ethereum.Contracts.Ledgers = nil
	// osp1 profile
	profile.Osp.Address = osp1Addr
	profile.Osp.Host = o1Port
	e2e.SaveProfile(&profile, profileDir+"o1_profile.json")
	// osp2 profile
	profile.Osp.Address = osp2Addr
	profile.Osp.Host = o2Port
	e2e.SaveProfile(&profile, profileDir+"o2_profile.json")
	// osp3 profile
	profile.Osp.Address = osp3Addr
	profile.Osp.Host = o3Port
	e2e.SaveProfile(&profile, profileDir+"o3_profile.json")
	// osp4 profile
	profile.Osp.Address = osp4Addr
	profile.Osp.Host = o4Port
	e2e.SaveProfile(&profile, profileDir+"o4_profile.json")
	// osp5 profile
	profile.Osp.Address = osp5Addr
	profile.Osp.Host = o5Port
	e2e.SaveProfile(&profile, profileDir+"o5_profile.json")

	// Fund two clients
	// NOTE: Make sure this is done AFTER SetupOnChain to maintain stable contract addresses
	// tf.CreateAccountsWithBalance(2, "100000000000000000000")

	log.Infoln("Local testnet setup finished.")
	log.Infoln("Kill this process to shut down the testnet.")

	// adance block every 1 second
	go advanceBlocks()
	<-make(chan bool)
}

func advanceBlocks() {
	// set log to warn level to avoid excessive log while advancing blocks
	log.SetLevel(log.WarnLevel)
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			tf.AdvanceBlock()
		}
	}
}
