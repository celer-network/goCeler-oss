// Copyright 2018-2020 Celer Network

// Setup blockchain etc for e2e tests
package e2e

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/celer-network/goCeler/common"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

var (
	reuse = flag.String("reuse", "", "data dir of previously failed test run, skip geth/building binaries if set")
)

// TestMain handles common setup (start chain, deploy, build  etc)
// and teardown. Test specific setup should be done in TestXxx
func TestMain(m *testing.M) {
	flag.Parse()
	log.SetPrefix("testing")
	log.EnableColor()
	if *reuse != "" {
		outRootDir = *reuse
	} else {
		// mkdir out root
		outRootDir = fmt.Sprintf("%s%d/", outRootDirPrefix, time.Now().Unix())
		err := os.MkdirAll(outRootDir, os.ModePerm)
		chkErr(err, "creating root dir")
	}

	fmt.Println("Using folder:", outRootDir)
	// set testing pkg level path
	tf.SetOutRootDir(outRootDir)
	var gethpid int
	// no reuse, do geth stuff, build bins and setup onchain
	if *reuse == "" {
		// start geth, not waiting for it to be fully ready. also watch geth proc
		// if geth exits with non-zero, os.Exit(1)
		ethProc, err := StartChain()
		chkErr(err, "starting chain")
		gethpid = ethProc.Pid
		// build binaries should take long enough for geth to be fully started
		err = buildBins(outRootDir)
		chkErr(err, "build binaries")
		// deploy contracts and fund ethpool etc, also update appAddrMap
		tf.E2eProfile, tokenAddrErc20 = SetupOnChain(appAddrMap, true)
		// profile.json is the default OSP profile
		noProxyProfile = outRootDir + "profile.json"
		saveProfile(tf.E2eProfile, noProxyProfile)
		// client profiles for multiosp tests
		c2o2Profile := *tf.E2eProfile
		c2o2Profile.Osp.Address = osp2EthAddr
		c2o2Profile.Osp.Host = localhost + o2Port
		saveProfile(&c2o2Profile, outRootDir+"c2o2.json")
		c3o3Profile := *tf.E2eProfile
		c3o3Profile.Osp.Address = osp3EthAddr
		c3o3Profile.Osp.Host = localhost + o3Port
		saveProfile(&c3o3Profile, outRootDir+"c3o3.json")
		c4o4Profile := *tf.E2eProfile
		c4o4Profile.Osp.Address = osp4EthAddr
		c4o4Profile.Osp.Host = localhost + o4Port
		saveProfile(&c4o4Profile, outRootDir+"c4o4.json")
		c5o5Profile := *tf.E2eProfile
		c5o5Profile.Osp.Address = osp5EthAddr
		c5o5Profile.Osp.Host = localhost + o5Port
		saveProfile(&c5o5Profile, outRootDir+"c5o5.json")

		saveMisc(outRootDir+"misc.json", gethpid, tokenAddrErc20, appAddrMap)
	} else {
		noProxyProfile = outRootDir + "profile.json"
		tf.E2eProfile = common.ParseProfileJSON(noProxyProfile)
		// restore from file
		m := loadMisc(outRootDir + "misc.json")
		gethpid = m.GethPid
		tokenAddrErc20 = m.Erc20
		appAddrMap = m.AppMap
	}

	//TODO: update rt_config.json and tokens.json

	// run all e2e tests
	ret := m.Run()

	if ret == 0 {
		fmt.Println("All tests passed! 🎉🎉🎉")
		syscall.Kill(gethpid, syscall.SIGTERM)
		os.RemoveAll(outRootDir)
		os.RemoveAll(sStoreDir)
		os.Exit(0)
	} else {
		fmt.Println("Tests failed. 🚧🚧🚧 Geth still running for debug. 🚧🚧🚧", "Geth PID:", gethpid)
		fmt.Println("To skip setup and rerun tests, add flag -reuse", outRootDir)
		os.Exit(ret)
	}
}
