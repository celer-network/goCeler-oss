// Copyright 2018-2020 Celer Network

package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/celer-network/goCeler/ctype"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/utils"
)

func setMultiOSP() []*tf.ServerController {
	// Need register all osps on-chain before starting osps.
	tf.RegisterRouters([]string{osp1Keystore, osp2Keystore, osp3Keystore, osp4Keystore, osp5Keystore})
	os.RemoveAll(sStoreDir)
	// Be careful: because the limit of test set up, state between the two osps is kept during tests.
	// Each tests need to reset state between the two osps.
	o1 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", noProxyProfile,
		"-port", o1Port,
		"-storedir", sStoreDir,
		"-ks", ospKeystore,
		"-nopassword",
		"-rtc", rtConfigMultiOSP,
		"-defaultroute", osp3EthAddr,
		"-svrname", "o1",
		"-logcolor",
		"-logprefix", "o1_"+ospEthAddr[:4])

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

	o3 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", noProxyProfile,
		"-port", o3Port,
		"-storedir", sStoreDir,
		"-adminrpc", o3AdminRPC,
		"-adminweb", o3AdminWeb,
		"-ks", osp3Keystore,
		"-nopassword",
		"-rtc", rtConfigMultiOSP,
		"-defaultroute", osp4EthAddr,
		"-svrname", "o3",
		"-logcolor",
		"-logprefix", "o3_"+osp3EthAddr[:4])

	o4 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", noProxyProfile,
		"-port", o4Port,
		"-storedir", sStoreDir,
		"-adminrpc", o4AdminRPC,
		"-adminweb", o4AdminWeb,
		"-ks", osp4Keystore,
		"-nopassword",
		"-rtc", rtConfigMultiOSP,
		"-defaultroute", ospEthAddr,
		"-svrname", "o4",
		"-logcolor",
		"-logprefix", "o4_"+osp4EthAddr[:4])

	o5 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", noProxyProfile,
		"-port", o5Port,
		"-storedir", sStoreDir,
		"-adminrpc", o5AdminRPC,
		"-adminweb", o5AdminWeb,
		"-ks", osp5Keystore,
		"-nopassword",
		"-rtc", rtConfigMultiOSP,
		"-defaultroute", osp4EthAddr,
		"-svrname", "o5",
		"-logcolor",
		"-logprefix", "o5_"+osp5EthAddr[:4])

	time.Sleep(time.Second)
	/* OSPs connect with each other with topology:
	  o4---o5
	 /  \    \
	o3---o1---o2
	*/
	utils.RequestRegisterStream(o2AdminWeb, ctype.Hex2Addr(ospEthAddr), localhost+o1Port)
	utils.RequestRegisterStream(o3AdminWeb, ctype.Hex2Addr(ospEthAddr), localhost+o1Port)
	utils.RequestRegisterStream(o4AdminWeb, ctype.Hex2Addr(ospEthAddr), localhost+o1Port)
	utils.RequestRegisterStream(o4AdminWeb, ctype.Hex2Addr(osp3EthAddr), localhost+o3Port)
	utils.RequestRegisterStream(o4AdminWeb, ctype.Hex2Addr(osp5EthAddr), localhost+o5Port)
	utils.RequestRegisterStream(o5AdminWeb, ctype.Hex2Addr(osp2EthAddr), localhost+o2Port)
	time.Sleep(time.Second)

	return []*tf.ServerController{o1, o2, o3, o4, o5}
}

func TestE2EMultiOSP(t *testing.T) {
	svrs := setMultiOSP()
	defer tearDownMultiSvr([]Killable{svrs[0], svrs[1], svrs[2], svrs[3], svrs[4]})

	// Be careful: because the limit of test set up, state between the two osps is kept during tests.
	// Each tests need to reset state between the two osps.
	t.Run("e2e-multiosp-open-channel", func(t *testing.T) {
		t.Run("multiOspOpenChannelPolicyTest", multiOspOpenChannelPolicyTest)
		t.Run("multiOspOpenChannelPolicyFallbackTest", multiOspOpenChannelPolicyFallbackTest)
		t.Run("multiOspOpenChannelTest", multiOspOpenChannelTest)
	})
	t.Run("e2e-multiosp-routing", func(t *testing.T) {
		t.Run("multiOspRouting", multiOspRouting(svrs...))
	})
	t.Run("e2e-multiosp-channel-migration", func(t *testing.T) {
		t.Run("migrateChannelBetweenOsps", migrateChannelBetweenOsps(svrs...))
	})
}
