// Copyright 2020 Celer Network

package cli

import (
	"math/big"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/dispute"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/tools/toolsetup"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/watcher"
	"github.com/celer-network/goutils/log"
)

type Processor struct {
	myAddr     ctype.Addr
	profile    *common.CProfile
	nodeConfig common.GlobalNodeConfig
	dal        *storage.DAL
	transactor *transactor.Transactor
	disputer   *dispute.Processor
}

func (p *Processor) Setup(db, ospkey, disputer bool) {
	log.Debug("setup processor...")
	p.profile = common.ParseProfile(*pjson)
	overrideConfig(p.profile)
	p.myAddr = ctype.Hex2Addr(p.profile.SvrETHAddr)
	config.ChainID = big.NewInt(p.profile.ChainId)
	config.BlockDelay = p.profile.BlockDelayNum
	if db {
		p.dal = toolsetup.NewDAL(p.profile)
	}
	ethclient := toolsetup.NewEthClient(p.profile)
	p.nodeConfig = toolsetup.NewNodeConfig(p.profile, ethclient, p.dal)

	if *ksfile != "" {
		log.Debug("setup keystore...")
		var err error
		keyStore, passPhrase := toolsetup.ParseKeyStoreFile(*ksfile, *nopassword)
		if ospkey { // enforce using osp keystore
			myAddr, _, err2 := utils.GetAddrAndPrivKey(keyStore, passPhrase)
			if err2 != nil {
				log.Fatal(err2)
			}
			if p.myAddr != myAddr {
				log.Fatal("keystore and profile my address do not match")
			}
		}

		p.transactor, err = transactor.NewTransactor(keyStore, passPhrase, ethclient)
		if err != nil {
			log.Fatal(err)
		}

		if disputer {
			log.Debug("setup disputer...")
			transactorPool, err := transactor.NewPool([]*transactor.Transactor{p.transactor})
			if err != nil {
				log.Fatal(err)
			}
			var pollingInterval uint64 = 10
			if p.profile.PollingInterval != 0 {
				pollingInterval = p.profile.PollingInterval
			}
			watch := watcher.NewWatchService(ethclient, p.dal, pollingInterval)
			monitorService := monitor.NewService(
				watch, p.profile.BlockDelayNum, false, p.nodeConfig.GetRPCAddr())
			p.disputer = dispute.NewProcessor(
				p.nodeConfig, p.transactor, transactorPool, nil, monitorService, p.dal, false)
		}
	}
}

func overrideConfig(profile *common.CProfile) {
	profile.BlockDelayNum = uint64(*blkdelay)
	if *storesql != "" {
		profile.StoreSql = *storesql
	} else if *storedir != "" {
		profile.StoreDir = *storedir
	}
}
