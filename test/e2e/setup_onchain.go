// Copyright 2018-2020 Celer Network

package e2e

import (
	"context"
	"flag"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/deploy"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ethpool"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	tf "github.com/celer-network/goCeler/testing"
	"github.com/celer-network/goCeler/testing/testapp"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var conclient *ethclient.Client
var etherBaseAuth *bind.TransactOpts
var channelAddrBundle deploy.CelerChannelAddrBundle
var ethPoolContract *ethpool.EthPool
var erc20Contract *chain.ERC20
var autoFund bool

var grpAddrs = [][]string{
	[]string{ospEthAddr, depositorEthAddr, osp2EthAddr, osp3EthAddr, osp4EthAddr, osp5EthAddr},
	[]string{osp6EthAddr, osp7EthAddr},
	[]string{osp8EthAddr, osp9EthAddr},
}
var grpPrivs = [][]string{
	[]string{osp1Priv, depositorPriv, osp2Priv, osp3Priv, osp4Priv, osp5Priv},
	[]string{osp6Priv, osp7Priv},
	[]string{osp8Priv, osp9Priv},
}

// SetupOnChain deploy contracts, and set limit etc
// return profile, tokenAddrErc20 and set testapp related addr
func SetupOnChain(appMap map[string]ctype.Addr, groupId uint64, autofund bool) (*common.ProfileJSON, string) {
	flag.Parse()
	autoFund = autofund
	var err error
	conclient, err = ethclient.Dial(outRootDir + "chaindata/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to conclientect to the Ethereum: %v", err)
	}
	ethbasePrivKey, _ := crypto.HexToECDSA(etherBasePriv)
	etherBaseAuth = bind.NewKeyedTransactor(ethbasePrivKey)
	price := big.NewInt(2e9) // 2Gwei
	etherBaseAuth.GasPrice = price
	etherBaseAuth.GasLimit = 7000000

	ctx := context.Background()
	// deploy celer channel contracts
	channelAddrBundle = deploy.DeployAll(etherBaseAuth, conclient, ctx, 0)
	// deploy router registry
	routerRegistryAddr := deploy.DeployRouterRegistry(ctx, etherBaseAuth, conclient, 0)

	// EthPool is used later when adding fund to addr
	ethPoolContract, err = ethpool.NewEthPool(channelAddrBundle.EthPoolAddr, conclient)
	if err != nil {
		log.Fatal(err)
	}

	// Disable channel deposit limit
	ledgerContract, err := ledger.NewCelerLedger(channelAddrBundle.CelerLedgerAddr, conclient)
	if err != nil {
		log.Fatal(err)
	}
	tx1, err := ledgerContract.DisableBalanceLimits(etherBaseAuth)
	if err != nil {
		log.Fatalf("Failed disable channel deposit limits: %v", err)
	}

	// Deploy sample ERC20 contract (MOON)
	initAmt := new(big.Int)
	initAmt.SetString("500000000000000000000000000000000000000000000", 10)
	var erc20Addr ctype.Addr
	var tx2 *ethtypes.Transaction
	erc20Addr, tx2, erc20Contract, err = chain.DeployERC20(etherBaseAuth, conclient, initAmt, "Moon", 18, "MOON")
	if err != nil {
		log.Fatalf("Failed to deploy ERC20: %v", err)
	}

	// Deploy MultiSessionApp contract
	appAddr1, tx3, _, err := testapp.DeploySimpleMultiSessionApp(etherBaseAuth, conclient, testapp.PlayerNum)
	if err != nil {
		log.Fatalf("Failed to deploy SimpleMultiSessionApp contract: %v", err)
	}
	appMap["SimpleMultiSessionApp"] = appAddr1

	// Deploy MultiSessionAppWithOracle contract
	timeout := new(big.Int).SetUint64(2)
	appAddr2, tx4, _, err := testapp.DeploySimpleMultiSessionAppWithOracle(
		etherBaseAuth, conclient, timeout, timeout, testapp.PlayerNum, ctype.Hex2Addr(etherBaseAddr))
	if err != nil {
		log.Fatalf("Failed to deploy SimpleMultiSessionAppWithOracle contract: %v", err)
	}
	appMap["SimpleMultiSessionAppWithOracle"] = appAddr2

	// Deploy MultiGomoku contract
	appAddr3, tx5, _, err := testapp.DeployMultiGomoku(etherBaseAuth, conclient, testapp.GomokuMinOffChain, testapp.GomokuMaxOnChain)
	if err != nil {
		log.Fatalf("Failed to deploy MultiGomoku contract: %v", err)
	}
	appMap["MultiGomoku"] = appAddr3

	// Deploy a new Celer Ledger for channel migration test
	log.Infoln("Deploying new CelerLedger contract...")
	newLedgerAddr, tx6, _, err := deploy.DeployContractWithLinks(
		etherBaseAuth,
		conclient,
		ledger.CelerLedgerABI,
		ledger.CelerLedgerBin,
		map[string]ctype.Addr{
			"LedgerStruct":       channelAddrBundle.LedgerStructAddr,
			"LedgerOperation":    channelAddrBundle.OperationAddr,
			"LedgerChannel":      channelAddrBundle.LedgerChannelAddr,
			"LedgerBalanceLimit": channelAddrBundle.BalanceLimitAddr,
			"LedgerMigrate":      channelAddrBundle.MigrateAddr,
		},
		channelAddrBundle.EthPoolAddr,
		channelAddrBundle.PayRegistryAddr,
		channelAddrBundle.CelerWalletAddr,
	)
	if err != nil {
		log.Fatalf("Failed to deploy new CelerLedger contract: %w", err)
	}

	// wait mined and check status for tx1, tx2, tx3, tx4, tx5, tx6
	receipt, err := eth.WaitMined(ctx, conclient, tx1, eth.WithPollingInterval(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	chkTxStatus(receipt.Status, "Disable balance limit")
	receipt, err = eth.WaitMined(ctx, conclient, tx2, eth.WithPollingInterval(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	chkTxStatus(receipt.Status, "Deploy ERC20 "+ctype.Addr2Hex(erc20Addr))
	receipt, err = eth.WaitMined(ctx, conclient, tx3, eth.WithPollingInterval(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	chkTxStatus(receipt.Status, "Deploy SimpleMultiSessionApp "+ctype.Addr2Hex(appAddr1))
	receipt, err = eth.WaitMined(ctx, conclient, tx4, eth.WithPollingInterval(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	chkTxStatus(receipt.Status, "Deploy SimpleMultiSessionAppWithOracle "+ctype.Addr2Hex(appAddr2))
	receipt, err = eth.WaitMined(ctx, conclient, tx5, eth.WithPollingInterval(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	chkTxStatus(receipt.Status, "Deploy MultiGomoku "+ctype.Addr2Hex(appAddr3))
	receipt, err = eth.WaitMined(ctx, conclient, tx6, eth.WithPollingInterval(time.Second))
	if err != nil {
		log.Fatalf("Failed to WaitMined v2 CelerLedger: %w", err)
	}
	chkTxStatus(receipt.Status, "Deploy v2 Ledger contract at "+ctype.Addr2Hex(newLedgerAddr))

	log.Infoln("Add fund to OSP accounts ...")
	if groupId < 3 {
		fundEthAddrs(grpAddrs[groupId], grpPrivs[groupId])
	}

	// contruct ledger map
	ledgers := map[string]string{
		ctype.Addr2Hex(channelAddrBundle.CelerLedgerAddr): "ledger1",
		ctype.Addr2Hex(newLedgerAddr):                     "ledger2",
	}

	profileContracts := common.ProfileContracts{
		Wallet:         ctype.Addr2Hex(channelAddrBundle.CelerWalletAddr),
		Ledger:         ctype.Addr2Hex(channelAddrBundle.CelerLedgerAddr),
		VirtResolver:   ctype.Addr2Hex(channelAddrBundle.VirtResolverAddr),
		EthPool:        ctype.Addr2Hex(channelAddrBundle.EthPoolAddr),
		PayResolver:    ctype.Addr2Hex(channelAddrBundle.PayResolverAddr),
		PayRegistry:    ctype.Addr2Hex(channelAddrBundle.PayRegistryAddr),
		RouterRegistry: ctype.Addr2Hex(routerRegistryAddr),
		Ledgers:        ledgers,
	}

	profileEth := common.ProfileEthereum{
		Gateway:          ethGateway,
		ChainId:          883,
		BlockIntervalSec: 1,
		BlockDelayNum:    0,
		DisputeTimeout:   10,
		Contracts:        profileContracts,
		CheckInterval: map[string]uint64{
			"CooperativeWithdraw": 2,
			"Deploy":              2,
			"Deposit":             2,
			"IntendSettle":        2,
			"OpenChannel":         2,
			"ConfirmSettle":       2,
			"IntendWithdraw":      2,
			"ConfirmWithdraw":     2,
			"RouterUpdated":       2,
			"MigrateChannelTo":    2,
		},
	}

	profileOsp := common.ProfileOsp{
		Host:    "localhost:10000",
		Address: ospEthAddr,
	}

	// output json file
	p := &common.ProfileJSON{
		Version:  "0.1",
		Ethereum: profileEth,
		Osp:      profileOsp,
	}
	return p, ctype.Addr2Hex(erc20Addr)
}

func fundEthAddr(addrStr, privKeyStr string) {
	addr := ctype.Hex2Addr(addrStr)
	err := tf.FundAddr("100000000000000000000", []*ctype.Addr{&addr})
	if err != nil {
		log.Fatalln("failed to fund addr", addrStr, err)
	}
	tx1, tx2 := fundEthAddrStep1(addrStr)
	fundEthAddrStep1Check(addrStr, tx1, tx2)
	tx3, tx4 := fundEthAddrStep2(addrStr, privKeyStr)
	fundEthAddrStep2Check(addrStr, tx3, tx4)
}

func fundEthAddrs(addrStrs, privKeyStr []string) {
	var addrs []*ctype.Addr
	for _, addrStr := range addrStrs {
		addr := ctype.Hex2Addr(addrStr)
		addrs = append(addrs, &addr)
	}
	err := tf.FundAddr("1000000000000000000000000", addrs) // 1 million ETH
	if err != nil {
		log.Fatalln("failed to fund", err)
	}

	var tx1s, tx2s, tx3s, tx4s []*ethtypes.Transaction
	for i, _ := range addrStrs {
		tx1, tx2 := fundEthAddrStep1(addrStrs[i])
		tx1s = append(tx1s, tx1)
		tx2s = append(tx2s, tx2)
	}
	for i, _ := range addrStrs {
		fundEthAddrStep1Check(addrStrs[i], tx1s[i], tx2s[i])
	}
	if autoFund {
		for i, _ := range addrStrs {
			tx3, tx4 := fundEthAddrStep2(addrStrs[i], privKeyStr[i])
			tx3s = append(tx3s, tx3)
			tx4s = append(tx4s, tx4)
		}
		for i, _ := range addrStrs {
			fundEthAddrStep2Check(addrStrs[i], tx3s[i], tx4s[i])
		}
	}
}

func fundEthAddrStep1(addrStr string) (*ethtypes.Transaction, *ethtypes.Transaction) {
	var tx1, tx2 *ethtypes.Transaction
	var err error
	addr := ctype.Hex2Addr(addrStr)
	if autoFund {
		ethAmt := new(big.Int)
		ethAmt.SetString("1000000000000000000000000", 10) // 1 million ETH
		etherBaseAuth.Value = ethAmt
		tx1, err = ethPoolContract.Deposit(etherBaseAuth, addr)
		if err != nil {
			log.Fatalln("failed to deposit into ethpool", addrStr, err)
		}
		etherBaseAuth.Value = big.NewInt(0)
	}
	moonAmt := new(big.Int)
	moonAmt.SetString("1000000000000000000000000000", 10) // 1 billion Moon
	tx2, err = erc20Contract.Transfer(etherBaseAuth, addr, moonAmt)
	if err != nil {
		log.Fatalln("failed to send MOON token for", addrStr, err)
	}

	return tx1, tx2
}

func fundEthAddrStep2(addrStr, privKeyStr string) (*ethtypes.Transaction, *ethtypes.Transaction) {
	var tx3, tx4 *ethtypes.Transaction
	var err error
	privKey, err := crypto.HexToECDSA(privKeyStr)
	if err != nil {
		log.Fatalln("failed to get private key", addrStr, err)
	}
	auth := bind.NewKeyedTransactor(privKey)
	auth.GasPrice = etherBaseAuth.GasPrice
	ethAmt := new(big.Int)
	ethAmt.SetString("1000000000000000000000000", 10) // 1 million ETH
	// Approve transferFrom of eth from ethpool for celerLedger
	tx3, err = ethPoolContract.Approve(auth, channelAddrBundle.CelerLedgerAddr, ethAmt)
	if err != nil {
		log.Fatalln("failed to approve ETH to celerLedger for", addrStr, err)
	}

	moonAmt := new(big.Int)
	moonAmt.SetString("1000000000000000000000000000", 10) // 1 billion Moon
	tx4, err = erc20Contract.Approve(auth, channelAddrBundle.CelerLedgerAddr, moonAmt)
	if err != nil {
		log.Fatalln("failed to approve MOON to celerLedger for", addrStr, err)
	}

	return tx3, tx4
}

func fundEthAddrStep1Check(addrStr string, tx1, tx2 *ethtypes.Transaction) {
	ctx := context.Background()
	// wait mined and check status for tx1 and tx2
	if autoFund {
		receipt, err := eth.WaitMined(ctx, conclient, tx1, eth.WithPollingInterval(time.Second))
		if err != nil {
			log.Fatalln("wait mined failed", addrStr, err)
		}
		chkTxStatus(receipt.Status, "deposit to ethpool for "+addrStr)
	}
	receipt, err := eth.WaitMined(ctx, conclient, tx2, eth.WithPollingInterval(time.Second))
	if err != nil {
		log.Fatalln("wait mined failed", addrStr, err)
	}
	chkTxStatus(receipt.Status, "transfer moon token to "+addrStr)
}

func fundEthAddrStep2Check(addrStr string, tx3, tx4 *ethtypes.Transaction) {
	ctx := context.Background()
	// wait mined and check status for tx3 and tx4
	receipt, err := eth.WaitMined(ctx, conclient, tx3, eth.WithPollingInterval(time.Second))
	if err != nil {
		log.Fatalln("wait mined failed", addrStr, err)
	}
	chkTxStatus(receipt.Status, addrStr+" approve ethpool to ledger")

	receipt, err = eth.WaitMined(ctx, conclient, tx4, eth.WithPollingInterval(time.Second))
	if err != nil {
		log.Fatalln("wait mined failed", addrStr, err)
	}
	chkTxStatus(receipt.Status, addrStr+" approve moon token")
	log.Infoln("finish funding for", addrStr)
}

// if status isn't 1 (sucess), log.Fatal
func chkTxStatus(s uint64, txname string) {
	if s != 1 {
		log.Fatal(txname + " tx failed")
	}
	log.Info(txname + " tx success")
}
