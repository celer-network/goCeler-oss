// Copyright 2018-2020 Celer Network

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/intfs"
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

var (
	pjson           = flag.String("profile", "", "OSP profile")
	storesql        = flag.String("storesql", "", "sql database URL")
	storedir        = flag.String("storedir", "", "local database directory")
	ksfile          = flag.String("ks", "", "key store file")
	blkDelay        = flag.Int("blkdelay", 0, "block delay for wait mined")
	noPassword      = flag.Bool("nopassword", false, "Assume empty password for keystores")
	intendSettle    = flag.String("intendsettle", "", "channel ID or channel list file (for batch) to intend settle")
	confirmSettle   = flag.String("confirmsettle", "", "channel ID channel list file (for batch) to confirm settle")
	batch           = flag.Bool("batch", false, "batch operation")
	intendWithdraw  = flag.String("intendwithdraw", "", "channel ID to withdraw fund from")
	withdrawAmt     = flag.Float64("withdrawamt", 0, "amount to withdraw in 10^18wei")
	withdrawTo      = flag.String("withdrawto", "", "channel ID to withdraw fund to, only needed if withdraw to another channel")
	confirmWithdraw = flag.String("confirmwithdraw", "", "channel ID of confirm withdraw from")
)

func main() {
	flag.Parse()
	var p processor
	p.setup()
	fmt.Println()

	if *intendSettle != "" {
		if *batch {
			dat, err := ioutil.ReadFile(*intendSettle)
			if err != nil {
				log.Fatal(err)
			}
			cids := strings.Fields(string(dat))
			for _, cid := range cids {
				err = p.disputer.IntendSettlePaymentChannel(ctype.Hex2Cid(cid), false)
				if err != nil {
					log.Fatalln(cid, err)
				}
			}
		} else {
			cid := ctype.Hex2Cid(*intendSettle)
			err := p.disputer.IntendSettlePaymentChannel(cid, true)
			if err != nil {
				log.Fatalln(*intendSettle, err)
			}
		}
	}
	if *confirmSettle != "" {
		if *batch {
			dat, err := ioutil.ReadFile(*confirmSettle)
			if err != nil {
				log.Fatal(err)
			}
			cids := strings.Fields(string(dat))
			for _, cid := range cids {
				err = p.disputer.ConfirmSettlePaymentChannel(ctype.Hex2Cid(cid), false)
				if err != nil {
					log.Fatalln(cid, err)
				}
			}
		} else {
			cid := ctype.Hex2Cid(*confirmSettle)
			err := p.disputer.ConfirmSettlePaymentChannel(cid, true)
			if err != nil {
				log.Fatalln(*confirmSettle, err)
			}
		}
	}
	if *intendWithdraw != "" && *withdrawAmt > 0 {
		fromCid := ctype.Hex2Cid(*intendWithdraw)
		amt := utils.Float2Wei(*withdrawAmt)
		toCid := ctype.ZeroCid
		if *withdrawTo != "" {
			toCid = ctype.Hex2Cid(*withdrawTo)
		}
		err := p.disputer.IntendWithdraw(fromCid, amt, toCid)
		if err != nil {
			log.Fatalln(err)
		}
	}
	if *confirmWithdraw != "" {
		cid := ctype.Hex2Cid(*confirmWithdraw)
		err := p.disputer.ConfirmWithdraw(cid)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

type processor struct {
	nodeConfig     common.GlobalNodeConfig
	transactor     *transactor.Transactor
	transactorPool *transactor.Pool
	monitorService intfs.MonitorService
	disputer       *dispute.Processor
	dal            *storage.DAL
}

func (p *processor) setup() {
	profile := common.ParseProfile(*pjson)
	overrideConfig(profile)
	config.ChainID = big.NewInt(profile.ChainId)
	config.BlockDelay = profile.BlockDelayNum

	ethclient := toolsetup.NewEthClient(profile)

	p.dal = toolsetup.NewDAL(profile)

	p.nodeConfig = toolsetup.NewNodeConfig(profile, ethclient, p.dal)

	keyStore, passPhrase := toolsetup.ParseKeyStoreFile(*ksfile, *noPassword)

	var err error
	p.transactor, err = transactor.NewTransactor(keyStore, passPhrase, ethclient)
	if err != nil {
		log.Fatal(err)
	}
	p.transactorPool, err = transactor.NewPool([]*transactor.Transactor{p.transactor})
	if err != nil {
		log.Fatal(err)
	}

	var pollingInterval uint64 = 10
	if profile.PollingInterval != 0 {
		pollingInterval = profile.PollingInterval
	}
	watch := watcher.NewWatchService(ethclient, p.dal, pollingInterval)
	p.monitorService = monitor.NewService(
		watch, profile.BlockDelayNum, false, p.nodeConfig.GetRPCAddr())

	p.disputer = dispute.NewProcessor(
		p.nodeConfig, p.transactor, p.transactorPool, nil, p.monitorService, p.dal, false)
}

func overrideConfig(profile *common.CProfile) {
	profile.BlockDelayNum = uint64(*blkDelay)
	if *storesql != "" {
		profile.StoreSql = *storesql
	} else if *storedir != "" {
		profile.StoreDir = *storedir
	}
}
