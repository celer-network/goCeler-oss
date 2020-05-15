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
	tf.StartProcess(outRootDir+"ospcli",
		"-adminhostport", o1AdminWeb,
		"-registerstream",
		"-peer", osp2EthAddr,
		"-peerhostport", localhost+o2Port,
		"-logcolor",
		"-logprefix", "cli").Wait()
	sleep(3)

	// open channel
	tf.StartProcess(outRootDir+"ospcli",
		"-adminhostport", o1AdminWeb,
		"-openchannel",
		"-peerdeposit", "5",
		"-selfdeposit", "5",
		"-peer", osp2EthAddr,
		"-logcolor",
		"-logprefix", "cli").Wait()
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
	tf.StartProcess(outRootDir+"ospcli",
		"-adminhostport", o1AdminWeb,
		"-sendtoken",
		"-receiver", osp2EthAddr,
		"-amount", "1",
		"-logcolor",
		"-logprefix", "cli").Wait()
	sleep(3)

	// check o1 balance
	free, err = getEthBalance(localhost+o1Port, osp2EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if free != "4000000000000000000" {
		t.Errorf("expect %s sending capacity %s, got %s", osp1EthAddr, "4000000000000000000", free)
		return
	}

	// check o2 balance
	free, err = getEthBalance(localhost+o2Port, osp1EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if free != "6000000000000000000" {
		t.Errorf("expect %s sending capacity %s, got %s", osp2EthAddr, "6000000000000000000", free)
		return
	}

	// make deposit to o1
	tf.StartProcess(outRootDir+"ospcli",
		"-adminhostport", o1AdminWeb,
		"-deposit",
		"-peer", osp2EthAddr,
		"-amount", "0.1",
		"-logcolor",
		"-logprefix", "cli").Wait()
	sleep(5)

	// check o1 balance
	free, err = getEthBalance(localhost+o1Port, osp2EthAddr)
	if err != nil {
		t.Error(err)
		return
	}
	if free != "4100000000000000000" {
		t.Errorf("expect %s sending capacity %s, got %s", osp1EthAddr, "4100000000000000000", free)
		return
	}
}
