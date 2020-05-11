// Copyright 2018-2020 Celer Network

package e2e

import "github.com/celer-network/goCeler/ctype"

const (
	// outPathPrefix is the path prefix for all output from e2e (incl. chain data, binaries etc)
	// the code will append epoch second to this and create the folder
	// the folder will be deleted after test ends successfully
	outRootDirPrefix = "/tmp/celer_e2e_"

	// etherbase and osp addr/priv key in hex
	etherBaseAddr    = "b5bb8b7f6f1883e0c01ffb8697024532e6f3238c"
	osp1EthAddr      = "0015f5863ddc59ab6610d7b6d73b2eacd43e6b7e"
	osp2EthAddr      = "00290a43e5b2b151d530845b2d5a818240bc7c70"
	osp3EthAddr      = "003ea363bccfd7d14285a34a6b1deb862df0bc84"
	osp4EthAddr      = "00495b55a68b5d5d1b0860b2c9eeb839e7d3a362"
	osp5EthAddr      = "005e9930a80df77fe686225a95be93548cdfa4b0"
	depositorEthAddr = "d0c5e4abfadbc0e71bfb6c4955e66b8a6bf4da51"
	ospEthAddr       = osp1EthAddr

	etherBasePriv = "69ef4da8204644e354d759ca93b94361474259f63caac6e12d7d0abcca0063f8"
	osp1Priv      = "06c5923cbaf9bc3617fba223e8ca1c9fd1e290c74124aa1359fd6119a1bb2704"
	osp2Priv      = "b6736f13344545561b1f279ffa935c9f614eceba097437823b95b5a615856306"
	osp3Priv      = "cd6edaa990081545e3e1c936e774a42db11431e0f6f7f8b4dfc851b2e408759c"
	osp4Priv      = "1d762733ae659befd6a2d70883de7739c1f0304bee207390d662424795d123dc"
	osp5Priv      = "2e727669a76f26ac61ee3216dae36d66f359620b9b27ab425248f8320581fe67"
	depositorPriv = "c76dc6d854247299174f9582566cd195a57d8236a7fb2d0c085377f274584cec"

	ethGateway = "http://127.0.0.1:8545"

	// try to do some allocation for port: 10xyz are osp,
	// x is osp 0-based index
	// yz are osp ports like grpc, adminweb, selfrpc etc
	localhost = "127.0.0.1:"

	sPort     = "10000"
	sSelfRPC  = "localhost:30000"
	sAdminRPC = "localhost:11000"
	sAdminWeb = "localhost:8090"

	// below are ports/addrs for multi-osp tests
	// o1 is the server for osp1, it is mapped to the default server above
	o1Port     = sPort
	o1SelfRPC  = sSelfRPC
	o1AdminRPC = sAdminRPC
	o1AdminWeb = sAdminWeb
	// o2 is the single server of osp2
	o2Port     = "10002"
	o2AdminRPC = "localhost:11002"
	o2AdminWeb = "localhost:8290"
	// o3 is the single server of osp3
	o3Port     = "10003"
	o3AdminRPC = "localhost:11003"
	o3AdminWeb = "localhost:8390"
	// o4 is the single server of osp4
	o4Port     = "10004"
	o4AdminRPC = "localhost:11004"
	o4AdminWeb = "localhost:8490"
	// o5 is the single server of osp5
	o5Port     = "10005"
	o5AdminRPC = "localhost:11005"
	o5AdminWeb = "localhost:8590"

	sStoreDir               = "/tmp/sStore"
	c1StoreSettleDisputeDir = "/tmp/c1StoreSettleDispute"
	oracleStoreDir          = "/tmp/oracleStore"

	sendAmt      = "1"
	tokenAddrEth = ctype.EthTokenAddrStr

	accountBalance      = "50000000000000000000" // 50 ETH
	initialBalance      = "5000000000000000000"  // 5 ETH
	initOspToOspBalance = "8000000000000000000"  // 8 ETH

	rtConfig         = "../../testing/profile/rt_config.json"
	rtConfigMultiOSP = "../../testing/profile/rt_config_multiosp.json"
	tokensConfig     = "../../testing/profile/tokens.json"

	osp1Keystore    = "../../testing/env/keystore/osp1.json"
	osp2Keystore    = "../../testing/env/keystore/osp2.json"
	osp3Keystore    = "../../testing/env/keystore/osp3.json"
	osp4Keystore    = "../../testing/env/keystore/osp4.json"
	osp5Keystore    = "../../testing/env/keystore/osp5.json"
	depositKeystore = "../../testing/env/keystore/osp1_depositor.json"
	ospKeystore     = osp1Keystore
)
