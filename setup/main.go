package main

import (
	"context"
	"flag"
	"os"

	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/ethpool"
	rr "github.com/celer-network/goCeler-oss/chain/channel-eth-go/routerregistry"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/ctype"
	envutils "github.com/celer-network/goCeler-oss/setup/utils"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	cfgfile = flag.String("profile", "", "Backend profile json file for server")
	amt     = flag.Float64("amt", 0, "Token amount (in ETH) to deposit")
	ksfile  = flag.String("ks", "", "Server keystore file")
)

func main() {
	flag.Parse()
	if *cfgfile == "" {
		log.Fatalln("profile was not set")
	}
	if *amt <= 0 {
		log.Fatalln("incorrect amt")
	}
	if *ksfile == "" {
		log.Fatalln("ks was not set")
	}

	passphrase := envutils.GetStringFromStdin("Enter the passphrase of ksfile: ", true)

	ksf, err := os.Open(*ksfile)
	envutils.ChkErr(err)

	_, svrAddr := envutils.GetKeyStore(*ksfile)
	cfg := envutils.ParseBackendProfileJSON(*cfgfile)
	log.Infof("Server: %s", svrAddr)

	conn, err := ethclient.Dial(cfg.Ethereum.Gateway)
	envutils.ChkErr(err)
	auth, err := bind.NewTransactor(ksf, passphrase)
	envutils.ChkErr(err)

	svrAddress := ctype.Hex2Addr(svrAddr)
	amtWei := utils.Float2Wei(*amt)

	ethPool := cfg.Ethereum.Contracts.EthPool
	ep, err := ethpool.NewEthPool(ctype.Hex2Addr(ethPool), conn)
	envutils.ChkErr(err)
	log.Infof("Depositing %s ETH to ethpool %s", envutils.WeiToEthStr(amtWei), ethPool)
	auth.Value = amtWei
	tx, err := ep.Deposit(auth, svrAddress)
	envutils.ChkErr(err)
	receipt, err := utils.WaitMined(context.Background(), conn, tx, 0)
	envutils.ChkErr(err)
	if receipt.Status != 1 {
		log.Fatal("Deposit transaction failed")
	}
	log.Info("Deposit Done.")
	total, err := ep.BalanceOf(nil, svrAddress)
	envutils.ChkErr(err)
	log.Infoln("Balance:", envutils.WeiToEthStr(total), "ETH")

	// approve to ledger
	// reset auth args
	auth.Value = nil
	auth.GasPrice = nil
	auth.GasLimit = 0
	log.Infoln("Approving...")
	tx, err = ep.Approve(auth, ctype.Hex2Addr(cfg.Ethereum.Contracts.Ledger), total)
	envutils.ChkErr(err)
	receipt, err = utils.WaitMined(context.Background(), conn, tx, 0)
	envutils.ChkErr(err)
	if receipt.Status != 1 {
		log.Fatal("Approve transaction failed")
	}
	log.Info("Approve Done.")
	allowed, err := ep.Allowance(nil, auth.From, ctype.Hex2Addr(cfg.Ethereum.Contracts.Ledger))
	log.Infof("Allowance from %s to %s: %s", auth.From.Hex(), cfg.Ethereum.Contracts.Ledger, envutils.WeiToEthStr(allowed))

	// mark self as router
	r, err := rr.NewRouterRegistry(ctype.Hex2Addr(cfg.Ethereum.Contracts.RouterRegistry), conn)
	envutils.ChkErr(err)
	blk, err := r.RouterInfo(nil, svrAddress)
	envutils.ChkErr(err)
	if blk.Uint64() == 0 {
		log.Infof("Registering %s as router...", svrAddr)
		tx, err = r.RegisterRouter(auth)
		envutils.ChkErr(err)
		receipt, err = utils.WaitMined(context.Background(), conn, tx, 0)
		envutils.ChkErr(err)
		if receipt.Status != 1 {
			log.Fatal("Register transaction failed")
		}
		log.Info("Register Done.")
	} else {
		log.Infoln("Registered at block", blk.String())
	}
	log.Infoln("Welcome to Celer Network!")
}
