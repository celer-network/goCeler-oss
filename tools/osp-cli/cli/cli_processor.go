// Copyright 2020 Celer Network

package cli

import (
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/dispute"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/tools/toolsetup"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/eth/monitor"
	"github.com/celer-network/goutils/eth/watcher"
	"github.com/celer-network/goutils/log"
)

type Processor struct {
	myAddr     ctype.Addr
	profile    *common.CProfile
	nodeConfig common.GlobalNodeConfig
	dal        *storage.DAL
	transactor *eth.Transactor
	disputer   *dispute.Processor
}

func (p *Processor) Setup(db, ospkey, disputer bool) {
	log.Debug("setup processor...")
	p.profile = common.ParseProfile(*pjson)
	overrideProfile(p.profile)
	config.SetGlobalConfigFromProfile(p.profile)
	p.myAddr = ctype.Hex2Addr(p.profile.SvrETHAddr)
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
			myAddr, _, err2 := eth.GetAddrPrivKeyFromKeystore(keyStore, passPhrase)
			if err2 != nil {
				log.Fatal(err2)
			}
			if p.myAddr != myAddr {
				log.Fatal("keystore and profile my address do not match")
			}
		}

		p.transactor, err = eth.NewTransactor(keyStore, passPhrase, ethclient, config.ChainId)
		if err != nil {
			log.Fatal(err)
		}

		if disputer {
			log.Debug("setup disputer...")
			transactorPool, err := eth.NewTransactorPool([]*eth.Transactor{p.transactor})
			if err != nil {
				log.Fatal(err)
			}
			watch := watcher.NewWatchService(ethclient, p.dal, config.BlockIntervalSec)
			monitorService := monitor.NewService(watch, p.profile.BlockDelayNum, false)
			p.disputer = dispute.NewProcessor(
				p.nodeConfig, p.transactor, transactorPool, nil, monitorService, p.dal, false)
		}
	}
}

func overrideProfile(profile *common.CProfile) {
	profile.BlockDelayNum = uint64(*blkdelay)
	if *storesql != "" {
		profile.StoreSql = *storesql
	} else if *storedir != "" {
		profile.StoreDir = *storedir
	}
}
