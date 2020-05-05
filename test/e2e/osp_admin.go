// Copyright 2020 Celer Network

package e2e

import (
	"testing"

	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goutils/log"
)

func ospAdminTest(t *testing.T) {
	log.Info("============== start test ospAdminTest ==============")
	defer log.Info("============== end test ospAdminTest ==============")
	buildPkgBin(outRootDir, "tools/osp-admin", "ospadmin")

	o2 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", noProxyProfile,
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
	sleep(2)

	// register stream
	tf.StartProcess(outRootDir+"ospadmin",
		"-adminhostport", o1AdminWeb,
		"-registerstream",
		"-peeraddr", osp2EthAddr,
		"-peerhostport", localhost+o2Port,
		"-logcolor",
		"-logprefix", "oa").Wait()
	sleep(3)

	// open channel
	tf.StartProcess(outRootDir+"ospadmin",
		"-adminhostport", o1AdminWeb,
		"-openchannel",
		"-peerdeposit", initialBalance,
		"-selfdeposit", initialBalance,
		"-peeraddr", osp2EthAddr,
		"-logcolor",
		"-logprefix", "oa").Wait()
	sleep(3)

	// check o1 balance
	free, err := getEthBalance(localhost+o1Port, osp2EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if free != initialBalance {
		t.Errorf("expect %s sending capacity %s, got %s", osp1EthAddr, initialBalance, free)
		return
	}

	// check o2 balance
	free, err = getEthBalance(localhost+o2Port, osp1EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if free != initialBalance {
		t.Errorf("expect %s sending capacity %s, got %s", osp2EthAddr, initialBalance, free)
		return
	}

	// send token
	tf.StartProcess(outRootDir+"ospadmin",
		"-adminhostport", o1AdminWeb,
		"-sendtoken",
		"-receiver", osp2EthAddr,
		"-amount", "99",
		"-logcolor",
		"-logprefix", "oa").Wait()
	sleep(3)

	// check o1 balance
	free, err = getEthBalance(localhost+o1Port, osp2EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if free != tf.AddAmtStr(initialBalance, "-99") {
		t.Errorf("expect %s sending capacity %s, got %s", osp1EthAddr, tf.AddAmtStr(initialBalance, "-99"), free)
		return
	}

	// check o2 balance
	free, err = getEthBalance(localhost+o2Port, osp1EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if free != tf.AddAmtStr(initialBalance, "99") {
		t.Errorf("expect %s sending capacity %s, got %s", osp2EthAddr, tf.AddAmtStr(initialBalance, "99"), free)
		return
	}

	// make deposit
	tf.StartProcess(outRootDir+"ospadmin",
		"-adminhostport", o1AdminWeb,
		"-deposit",
		"-peeraddr", osp2EthAddr,
		"-amount", "100",
		"-logcolor",
		"-logprefix", "oa").Wait()
	sleep(5)

	// check o1 balance
	free, err = getEthBalance(localhost+o1Port, osp2EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if free != tf.AddAmtStr(initialBalance, "1") {
		t.Errorf("expect %s sending capacity %s, got %s", osp1EthAddr, tf.AddAmtStr(initialBalance, "1"), free)
		return
	}
}
