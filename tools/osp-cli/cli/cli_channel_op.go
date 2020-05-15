// Copyright 2020 Celer Network

package cli

import (
	"io/ioutil"
	"strings"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

func (p *Processor) IntendSettle() {
	if *batchfile != "" {
		dat, err := ioutil.ReadFile(*batchfile)
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
		cid := ctype.Hex2Cid(*chanid)
		err := p.disputer.IntendSettlePaymentChannel(cid, true)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func (p *Processor) ConfirmSettle() {
	if *batchfile != "" {
		dat, err := ioutil.ReadFile(*batchfile)
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
		cid := ctype.Hex2Cid(*chanid)
		err := p.disputer.ConfirmSettlePaymentChannel(cid, true)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (p *Processor) IntendWithdraw() {
	fromCid := ctype.Hex2Cid(*chanid)
	amt := utils.Float2Wei(*amount)
	toCid := ctype.ZeroCid
	if *withdrawto != "" {
		toCid = ctype.Hex2Cid(*withdrawto)
	}
	err := p.disputer.IntendWithdraw(fromCid, amt, toCid)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) ConfirmWithdraw() {
	cid := ctype.Hex2Cid(*chanid)
	err := p.disputer.ConfirmWithdraw(cid)
	if err != nil {
		log.Fatal(err)
	}
}
