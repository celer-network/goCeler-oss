// Copyright 2020 Celer Network

package cli

import (
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/chain/channel-eth-go/ethpool"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/route/routerregistry"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

func (p *Processor) EthPoolDeposit() {
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
	p.queryEthPoolLedgerAllowance()
}

func (p *Processor) EthPoolWithdraw() {
	// withdraw ETH from EthPool contract
	err := p.withdrawEthPool()
	if err != nil {
		return
	}
}

func (p *Processor) RegisterRouter() {
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

func (p *Processor) DeregisterRouter() {
	// check router registration
	blk, err := p.queryRouterRegistry()
	if err != nil {
		return
	}
	// registry router
	if blk == 0 {
		log.Info("OSP not registered as a network router")
		return
	}
	p.deregisterRouter()
}

func (p *Processor) depositEthPool() error {
	log.Infof("deposit %f ETH to EthPool and wait transaction to be mined...", *amount)
	amtWei := utils.Float2Wei(*amount)
	ethPoolAddr := ctype.Hex2Addr(p.profile.EthPoolAddr)

	receipt, err := p.transactor.TransactWaitMined(
		"ethpool deposit",
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := ethpool.NewEthPoolTransactor(ethPoolAddr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.Deposit(opts, p.myAddr)
		},
		config.TransactOptions(eth.WithEthValue(amtWei))...)
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("ethpool deposit transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (p *Processor) approveEthPoolToLedger() error {
	log.Info("approve EthPool balance to CelerLedger and wait transaction to be mined...")
	balance, err := p.queryEthPoolBalance()
	if err != nil {
		return err
	}
	ethPoolAddr := ctype.Hex2Addr(p.profile.EthPoolAddr)
	ledgerAddr := ctype.Hex2Addr(p.profile.LedgerAddr)

	receipt, err := p.transactor.TransactWaitMined(
		"ethpool approve",
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := ethpool.NewEthPoolTransactor(ethPoolAddr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.Approve(opts, ledgerAddr, balance)
		},
		config.TransactOptions()...)
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("ethpool approve transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (p *Processor) queryEthPoolBalance() (*big.Int, error) {
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

func (p *Processor) queryEthPoolLedgerAllowance() (*big.Int, error) {
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

func (p *Processor) withdrawEthPool() error {
	log.Infof("withdraw %f ETH from EthPool and wait transaction to be mined...", *amount)
	amtWei := utils.Float2Wei(*amount)
	ethPoolAddr := ctype.Hex2Addr(p.profile.EthPoolAddr)

	receipt, err := p.transactor.TransactWaitMined(
		"ethpool withdraw",
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := ethpool.NewEthPoolTransactor(ethPoolAddr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.Withdraw(opts, amtWei)
		},
		config.TransactOptions()...)
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("ethpool withdraw transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (p *Processor) queryRouterRegistry() (uint64, error) {
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

func (p *Processor) registerRouter() error {
	log.Info("register OSP as state channel router and wait transaction to be mined...")
	routerRegistryAddr := ctype.Hex2Addr(p.profile.RouterRegistryAddr)

	receipt, err := p.transactor.TransactWaitMined(
		"register router",
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := routerregistry.NewRouterRegistryTransactor(routerRegistryAddr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.RegisterRouter(opts)
		},
		config.TransactOptions()...)
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("register router transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (p *Processor) deregisterRouter() error {
	log.Info("deregister OSP as state channel router and wait transaction to be mined...")
	routerRegistryAddr := ctype.Hex2Addr(p.profile.RouterRegistryAddr)

	receipt, err := p.transactor.TransactWaitMined(
		"deregister router",
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := routerregistry.NewRouterRegistryTransactor(routerRegistryAddr, transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.DeregisterRouter(opts)
		},
		config.TransactOptions()...)
	if err != nil {
		log.Error(err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("deregister router transaction %x failed", receipt.TxHash)
	}
	return nil
}
