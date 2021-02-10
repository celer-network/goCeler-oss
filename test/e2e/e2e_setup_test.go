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
	"github.com/celer-network/goCeler/ctype"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

var (
	reuse    = flag.String("reuse", "", "data dir of previously failed test run, skip geth/building binaries if set")
	multinet = flag.Bool("multinet", false, "multilple networks")
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
		tf.E2eProfile, tokenAddrErc20 = SetupOnChain(appAddrMap, 0, true)
		// profile.json is the default OSP profile
		noProxyProfile = outRootDir + "profile.json"
		saveProfile(tf.E2eProfile, noProxyProfile)
		// client profiles for multiosp tests
		o2Profile := *tf.E2eProfile
		o2Profile.Osp.Address = osp2EthAddr
		o2Profile.Osp.Host = localhost + o2Port
		saveProfile(&o2Profile, outRootDir+"profile-o2.json")
		o3Profile := *tf.E2eProfile
		o3Profile.Osp.Address = osp3EthAddr
		o3Profile.Osp.Host = localhost + o3Port
		saveProfile(&o3Profile, outRootDir+"profile-o3.json")
		o4Profile := *tf.E2eProfile
		o4Profile.Osp.Address = osp4EthAddr
		o4Profile.Osp.Host = localhost + o4Port
		saveProfile(&o4Profile, outRootDir+"profile-o4.json")
		o5Profile := *tf.E2eProfile
		o5Profile.Osp.Address = osp5EthAddr
		o5Profile.Osp.Host = localhost + o5Port
		saveProfile(&o5Profile, outRootDir+"profile-o5.json")

		saveMisc(outRootDir+"misc.json", gethpid, tokenAddrErc20, appAddrMap)

		if *multinet {
			var n1profile, n2profile *common.ProfileJSON
			log.Infoln("Setup network 1 ...")
			n1profile, tokenAddrNet1 = SetupOnChain(make(map[string]ctype.Addr), 1, true)
			o6Profile := *n1profile
			o6Profile.Osp.Address = osp6EthAddr
			o6Profile.Osp.Host = localhost + o6Port
			saveProfile(&o6Profile, outRootDir+"profile-o6-n1.json")
			o7Profile := *n1profile
			o7Profile.Osp.Address = osp7EthAddr
			o7Profile.Osp.Host = localhost + o7Port
			saveProfile(&o7Profile, outRootDir+"profile-o7-n1.json")
			log.Infoln("Setup network 2 ...")
			n2profile, tokenAddrNet2 = SetupOnChain(make(map[string]ctype.Addr), 2, true)
			o8Profile := *n2profile
			o8Profile.Osp.Address = osp8EthAddr
			o8Profile.Osp.Host = localhost + o8Port
			saveProfile(&o8Profile, outRootDir+"profile-o8-n2.json")
			o9Profile := *n2profile
			o9Profile.Osp.Address = osp9EthAddr
			o9Profile.Osp.Host = localhost + o9Port
			saveProfile(&o9Profile, outRootDir+"profile-o9-n2.json")
		}

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
