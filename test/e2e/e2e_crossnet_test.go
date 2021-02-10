// Copyright 2021 Celer Network

package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/storage"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/tools/toolsetup"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

const (
	n0NetId = 1000
	n1NetId = 1001
	n2NetId = 1002
)

func setMultiNet() []Killable {
	os.RemoveAll(sStoreDir)

	o1 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", noProxyProfile,
		"-port", o1Port,
		"-storedir", sStoreDir,
		"-ks", ospKeystore,
		"-nopassword",
		"-rtc", rtConfig,
		"-svrname", "o1",
		"-logcolor",
		"-logprefix", "o1_"+ospEthAddr[:4])

	o2 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", outRootDir+"profile-o2.json",
		"-port", o2Port,
		"-storedir", sStoreDir,
		"-adminrpc", o2AdminRPC,
		"-adminweb", o2AdminWeb,
		"-ks", osp2Keystore,
		"-nopassword",
		"-rtc", rtConfig,
		"-svrname", "o2",
		"-logcolor",
		"-logprefix", "o2_"+osp2EthAddr[:4])

	o6 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", outRootDir+"profile-o6-n1.json",
		"-port", o6Port,
		"-storedir", sStoreDir,
		"-adminrpc", o6AdminRPC,
		"-adminweb", o6AdminWeb,
		"-ks", osp6Keystore,
		"-nopassword",
		"-rtc", rtConfig,
		"-svrname", "o6",
		"-logcolor",
		"-logprefix", "o6_"+osp6EthAddr[:4])

	o7 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", outRootDir+"profile-o7-n1.json",
		"-port", o7Port,
		"-storedir", sStoreDir,
		"-adminrpc", o7AdminRPC,
		"-adminweb", o7AdminWeb,
		"-ks", osp7Keystore,
		"-nopassword",
		"-rtc", rtConfig,
		"-svrname", "o7",
		"-logcolor",
		"-logprefix", "o7_"+osp7EthAddr[:4])

	o8 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", outRootDir+"profile-o8-n2.json",
		"-port", o8Port,
		"-storedir", sStoreDir,
		"-adminrpc", o8AdminRPC,
		"-adminweb", o8AdminWeb,
		"-ks", osp8Keystore,
		"-nopassword",
		"-rtc", rtConfig,
		"-svrname", "o8",
		"-logcolor",
		"-logprefix", "o8_"+osp8EthAddr[:4])

	o9 := tf.StartServerController(outRootDir+toBuild["server"],
		"-profile", outRootDir+"profile-o9-n2.json",
		"-port", o9Port,
		"-storedir", sStoreDir,
		"-adminrpc", o9AdminRPC,
		"-adminweb", o9AdminWeb,
		"-ks", osp9Keystore,
		"-nopassword",
		"-rtc", rtConfig,
		"-svrname", "o9",
		"-logcolor",
		"-logprefix", "o9_"+osp9EthAddr[:4])

	time.Sleep(time.Second)
	/* OSPs connect with each other with topology:
	   n0          n1          n2
	o1===o2-----o6===o7-----o8===o9
	*/
	utils.RequestRegisterStream(o2AdminWeb, ctype.Hex2Addr(ospEthAddr), localhost+o1Port)
	utils.RequestRegisterStream(o6AdminWeb, ctype.Hex2Addr(osp2EthAddr), localhost+o2Port)
	utils.RequestRegisterStream(o7AdminWeb, ctype.Hex2Addr(osp6EthAddr), localhost+o6Port)
	utils.RequestRegisterStream(o8AdminWeb, ctype.Hex2Addr(osp7EthAddr), localhost+o7Port)
	utils.RequestRegisterStream(o9AdminWeb, ctype.Hex2Addr(osp8EthAddr), localhost+o8Port)
	time.Sleep(time.Second)
	return []Killable{o1, o2, o6, o7, o8, o9}
}

func TestE2ECrossNet(t *testing.T) {
	toKill := setMultiNet()
	defer tearDownMultiSvr(toKill)

	t.Run("e2e-crossnet", func(t *testing.T) {
		t.Run("crossNetSendEth", crossNetSendEth)
		t.Run("crossNetSendErc20", crossNetSendErc20)
	})
}

func crossNetSendEth(t *testing.T) {
	log.Info("============== start test crossNetSendEth ==============")
	defer log.Info("============== end test crossNetSendEth ==============")
	t.Parallel()
	crossNetSendToken(t, entity.TokenType_ETH, tokenAddrEth)
}

func crossNetSendErc20(t *testing.T) {
	log.Info("============== start test crossNetSendErc20 ==============")
	defer log.Info("============== end test crossNetSendErc20 ==============")
	t.Parallel()
	crossNetSendToken(t, entity.TokenType_ERC20, tokenAddrErc20)
}

func crossNetSendToken(t *testing.T, tokenType entity.TokenType, tokenAddr string) {
	tkaddr0 := tokenAddr
	tkaddr1 := tokenAddr
	tkaddr2 := tokenAddr
	if tokenType == entity.TokenType_ERC20 {
		tkaddr1 = tokenAddrNet1
		tkaddr2 = tokenAddrNet2
	}

	// Let osp2 initiate openning channel with osp1.
	err := requestOpenChannel(o2AdminWeb, osp1EthAddr, initialBalance, initialBalance, tkaddr0)
	if err != nil {
		log.Warn(err)
	}
	// Let osp7 initiate openning channel with osp6.
	err = requestOpenChannel(o7AdminWeb, osp6EthAddr, initialBalance, initialBalance, tkaddr1)
	if err != nil {
		t.Error(err)
		return
	}
	// Let osp9 initiate openning channel with osp8.
	err = requestOpenChannel(o9AdminWeb, osp8EthAddr, initialBalance, initialBalance, tkaddr2)
	if err != nil {
		t.Error(err)
		return
	}

	updateCrossNetTables()

	sleep(6)

	dal1, dal2, dal6, dal7, dal8, dal9 := getCrossNetOspDALs()

	log.Infoln("p1: o1 pay o9")
	p1, err := requestSendCrossNetToken(o1AdminWeb, osp9EthAddr, "1", tkaddr0, n2NetId)
	if err != nil {
		t.Error(err)
		return
	}
	originalPayId := ctype.Hex2PayID(p1)

	cid12, found, err := dal1.GetCidByPeerToken(
		ctype.Hex2Addr(osp2EthAddr), utils.GetTokenInfoFromAddress(ctype.Hex2Addr(tkaddr0)))
	if err != nil {
		t.Error(err)
		return
	}
	if !found {
		t.Error("channel cid12 not found")
		return
	}
	err = checkOspPayState(dal1, p1, ctype.ZeroCid, structs.PayState_NULL, cid12, structs.PayState_COSIGNED_PAID, 5)
	if err != nil {
		t.Errorf("p1 err at o1: %s", err)
		return
	}
	err = checkOspPayState(dal2, p1, cid12, structs.PayState_COSIGNED_PAID, ctype.ZeroCid, structs.PayState_NULL, 5)
	if err != nil {
		t.Errorf("p1 err at o2: %s", err)
		return
	}

	p2, state, _, found, err := dal6.GetCrossNetInfoByOrignalPayID(originalPayId)
	if err != nil {
		t.Error(err)
		return
	}
	if !found {
		t.Error("p2 on o6 not found")
		return
	}
	if state != structs.CrossNetPay_INGRESS {
		t.Errorf("invalid crossnet pay state %d", state)
	}
	cid67, found, err := dal6.GetCidByPeerToken(
		ctype.Hex2Addr(osp7EthAddr), utils.GetTokenInfoFromAddress(ctype.Hex2Addr(tkaddr1)))
	if err != nil {
		t.Error(err)
		return
	}
	if !found {
		t.Error("channel cid67 not found")
		return
	}
	err = checkOspPayState(dal6, ctype.PayID2Hex(p2), ctype.ZeroCid, structs.PayState_NULL, cid67, structs.PayState_COSIGNED_PAID, 5)
	if err != nil {
		t.Errorf("p2 err at o6: %s", err)
		return
	}
	err = checkOspPayState(dal7, ctype.PayID2Hex(p2), cid67, structs.PayState_COSIGNED_PAID, ctype.ZeroCid, structs.PayState_NULL, 5)
	if err != nil {
		t.Errorf("p2 err at o7: %s", err)
		return
	}

	p3, _, _, found, err := dal8.GetCrossNetInfoByOrignalPayID(originalPayId)
	if err != nil {
		t.Error(err)
		return
	}
	if !found {
		t.Error("p3 on o8 not found")
		return
	}
	cid89, found, err := dal8.GetCidByPeerToken(
		ctype.Hex2Addr(osp9EthAddr), utils.GetTokenInfoFromAddress(ctype.Hex2Addr(tkaddr2)))
	if err != nil {
		t.Error(err)
		return
	}
	if !found {
		t.Error("channel cid89 not found")
		return
	}
	err = checkOspPayState(dal8, ctype.PayID2Hex(p3), ctype.ZeroCid, structs.PayState_NULL, cid89, structs.PayState_COSIGNED_PAID, 5)
	if err != nil {
		t.Errorf("p3 err at o8: %s", err)
		return
	}
	err = checkOspPayState(dal9, ctype.PayID2Hex(p3), cid89, structs.PayState_COSIGNED_PAID, ctype.ZeroCid, structs.PayState_NULL, 5)
	if err != nil {
		t.Errorf("p3 err at o9: %s", err)
		return
	}

}

func getCrossNetOspDALs() (dal1, dal2, dal6, dal7, dal8, dal9 *storage.DAL) {
	profile1 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp1EthAddr,
	}
	dal1 = toolsetup.NewDAL(profile1)

	profile2 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp2EthAddr,
	}
	dal2 = toolsetup.NewDAL(profile2)

	profile6 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp6EthAddr,
	}
	dal6 = toolsetup.NewDAL(profile6)

	profile7 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp7EthAddr,
	}
	dal7 = toolsetup.NewDAL(profile7)

	profile8 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp8EthAddr,
	}
	dal8 = toolsetup.NewDAL(profile8)

	profile9 := &common.CProfile{
		StoreDir: sStoreDir + "/" + osp9EthAddr,
	}
	dal9 = toolsetup.NewDAL(profile9)

	return
}

func updateCrossNetTables() {
	//   n0          n1          n2
	// o1===o2-----o6===o7-----o8===o9
	profiles := []string{
		noProxyProfile,
		outRootDir + "profile-o2.json",
		outRootDir + "profile-o6-n1.json",
		outRootDir + "profile-o7-n1.json",
		outRootDir + "profile-o8-n2.json",
		outRootDir + "profile-o9-n2.json"}
	addrs := []string{osp1EthAddr, osp2EthAddr, osp6EthAddr, osp7EthAddr, osp8EthAddr, osp9EthAddr}
	names := []string{"o1", "o2", "o6", "o7", "o8", "o9"}

	for i := 0; i < 6; i++ {
		tf.StartProcess(outRootDir+"ospcli",
			"-profile", profiles[i],
			"-storedir", sStoreDir+"/"+addrs[i],
			"-dbupdate", "xnet",
			"-file", xnetConfigDir+names[i]+".json",
			"-logcolor",
			"-logprefix", "cli-"+names[i]).Wait()
	}
}
