// Copyright 2020 Celer Network

package main

import (
	"flag"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/chain/channel-eth-go/ethpool"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/route/routerregistry"
	"github.com/celer-network/goCeler/tools/toolsetup"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	pjson      = flag.String("profile", "", "OSP profile")
	ethpoolamt = flag.Float64("ethpoolamt", 0, "amount of ETH to deposit into EthPool")
	ksfile     = flag.String("ks", "", "key store file")
	blkDelay   = flag.Int("blkDelay", 2, "block delay for wait mined")
	noPassword = flag.Bool("nopassword", false, "assume empty password for keystores")
)

type processor struct {
	profile    *common.CProfile
	transactor *transactor.Transactor
	myAddr     ctype.Addr
}

func main() {
	flag.Parse()
	if *pjson == "" {
		log.Fatalln("profile was not set")
	}
	if *ksfile == "" {
		log.Fatalln("keystore was not set")
	}

	var p processor
	p.setup()

	// deposit ETH to EthPool contract
	err := p.depositEthPool()
	if err != nil {
		return
	}
	// approve EthPool balance to Ledger contract
	err = p.approveEthPoolToLedger()
	if err != nil {
		return
	}

	// check router registration
	blk, err := p.queryRouterRegistry()
	if err != nil {
		return
	}
	// registry router
	if blk == 0 {
		err = p.registerRouter()
		if err != nil {
			return
		}
		p.queryRouterRegistry()
	}

	log.Infoln("Welcome to Celer Network!")
}

func (p *processor) depositEthPool() error {
	if *ethpoolamt <= 0 {
		err := fmt.Errorf("incorrect ethpool amt")
		log.Error(err)
		return err
	}
	amtWei := utils.Float2Wei(*ethpoolamt)
	ethPoolAddr := ctype.Hex2Addr(p.profile.EthPoolAddr)

	receiptChan := make(chan *types.Receipt, 1)
	_, err := p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				receiptChan <- receipt
			},
		},
		amtWei,
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := ethpool.NewEthPoolTransactor(ethPoolAddr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.Deposit(opts, p.myAddr)
		})
	if err != nil {
		log.Error(err)
		return err
	}
	receipt := <-receiptChan
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("ethpool deposit transaction %x succeeded", receipt.TxHash)
	} else {
		log.Errorf("ethpool deposit transaction %x failed", receipt.TxHash)
		return fmt.Errorf("tx failed")
	}
	return nil
}

func (p *processor) approveEthPoolToLedger() error {
	balance, err := p.queryEthPoolBalance()
	if err != nil {
		return err
	}
	ethPoolAddr := ctype.Hex2Addr(p.profile.EthPoolAddr)
	ledgerAddr := ctype.Hex2Addr(p.profile.LedgerAddr)

	receiptChan := make(chan *types.Receipt, 1)
	_, err = p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				receiptChan <- receipt
			},
		},
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := ethpool.NewEthPoolTransactor(ethPoolAddr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.Approve(opts, ledgerAddr, balance)
		})
	if err != nil {
		log.Error(err)
		return err
	}
	receipt := <-receiptChan
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("approve ethpool to ledger transaction %x succeeded", receipt.TxHash)
	} else {
		log.Errorf("approve ethpool to ledger transaction %x failed", receipt.TxHash)
		return fmt.Errorf("tx failed")
	}
	return nil
}

func (p *processor) queryEthPoolBalance() (*big.Int, error) {
	ethPoolAddr := ctype.Hex2Addr(p.profile.EthPoolAddr)
	contract, err := ethpool.NewEthPoolCaller(ethPoolAddr, p.transactor.ContractCaller())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	balance, err := contract.BalanceOf(&bind.CallOpts{}, p.myAddr)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Infoln("my balance at EthPool:", balance)
	return balance, nil
}

func (p *processor) queryEthPoolLedgerAllowance() (*big.Int, error) {
	ethPoolAddr := ctype.Hex2Addr(p.profile.EthPoolAddr)
	ledgerAddr := ctype.Hex2Addr(p.profile.LedgerAddr)
	contract, err := ethpool.NewEthPoolCaller(ethPoolAddr, p.transactor.ContractCaller())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	allowance, err := contract.Allowance(&bind.CallOpts{}, p.myAddr, ledgerAddr)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Infoln("allowance from EthPool to Ledger is:", allowance)
	return allowance, nil
}

func (p *processor) queryRouterRegistry() (uint64, error) {
	routerRegistryAddr := ctype.Hex2Addr(p.profile.RouterRegistryAddr)
	contract, err := routerregistry.NewRouterRegistryCaller(routerRegistryAddr, p.transactor.ContractCaller())
	if err != nil {
		log.Error(err)
		return 0, err
	}
	blk, err := contract.RouterInfo(&bind.CallOpts{}, p.myAddr)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	blknum := blk.Uint64()
	if blknum != 0 {
		log.Infoln("router registered / refreshed at block", blknum)
	}
	return blknum, nil
}

func (p *processor) registerRouter() error {
	routerRegistryAddr := ctype.Hex2Addr(p.profile.RouterRegistryAddr)
	receiptChan := make(chan *types.Receipt, 1)
	_, err := p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				receiptChan <- receipt
			},
		},
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := routerregistry.NewRouterRegistryTransactor(routerRegistryAddr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.RegisterRouter(opts)
		})
	if err != nil {
		log.Error(err)
		return err
	}
	receipt := <-receiptChan
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("register router transaction %x succeeded", receipt.TxHash)
	} else {
		log.Errorf("register router transaction %x failed", receipt.TxHash)
		return fmt.Errorf("tx failed")
	}
	return nil
}

func (p *processor) deregisterRouter() error {
	routerRegistryAddr := ctype.Hex2Addr(p.profile.RouterRegistryAddr)

	receiptChan := make(chan *types.Receipt, 1)
	_, err := p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				receiptChan <- receipt
			},
		},
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := routerregistry.NewRouterRegistryTransactor(routerRegistryAddr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.DeregisterRouter(opts)
		})
	if err != nil {
		log.Error(err)
		return err
	}
	receipt := <-receiptChan
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("deregister router transaction %x succeeded", receipt.TxHash)
	} else {
		log.Errorf("deregister router transaction %x failed", receipt.TxHash)
		return fmt.Errorf("tx failed")
	}
	return nil
}

func (p *processor) setup() {
	p.profile = common.ParseProfile(*pjson)
	overrideConfig(p.profile)
	config.ChainID = big.NewInt(p.profile.ChainId)
	config.BlockDelay = p.profile.BlockDelayNum

	ethclient := toolsetup.NewEthClient(p.profile)
	keyStore, passPhrase := toolsetup.ParseKeyStoreFile(*ksfile, *noPassword)

	var err error
	p.myAddr, _, err = utils.GetAddrAndPrivKey(keyStore, passPhrase)
	if err != nil {
		log.Fatal(err)
	}
	if p.myAddr != ctype.Hex2Addr(p.profile.SvrETHAddr) {
		log.Fatal("incorrect profile")
	}

	p.transactor, err = transactor.NewTransactor(keyStore, passPhrase, ethclient)
	if err != nil {
		log.Fatal(err)
	}
}

func overrideConfig(profile *common.CProfile) {
	profile.BlockDelayNum = uint64(*blkDelay)
}
