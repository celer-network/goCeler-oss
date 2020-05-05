// Copyright 2019-2020 Celer Network

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/celer-network/goCeler/celersdk"
	"github.com/celer-network/goCeler/celersdkintf"
)

var (
	ks      = flag.String("ks", "", "path to keystore json file")
	profile = flag.String("profile", "", "path to profile json file")
	db      = flag.String("db", "", "path to save client db")
)

func main() {
	flag.Parse()
	k, _ := ioutil.ReadFile(*ks)
	p, _ := ioutil.ReadFile(*profile)
	cb := &appcb{
		Client: make(chan *celersdk.Client),
		Err:    make(chan error),
	}
	cc := getNewClient(string(k), string(p), *db, cb)
	if cc != nil {
		sleep(1)
		cc.Destroy()
	}
	sleep(5)
	cc = getNewClient(string(k), string(p), *db, cb)
	if cc != nil {
		sleep(1)
		cc.Destroy()
	}
	sleep(1)
}

func getNewClient(k, p, dir string, cb *appcb) *celersdk.Client {
	celersdk.InitClient(
		&celersdk.Account{
			Keystore: string(k),
		}, string(p), dir, cb)
	select {
	case <-cb.Err:
		// init client err, no usable client
		return nil
	case c := <-cb.Client:
		return c
	}
}

type appcb struct {
	Client chan *celersdk.Client
	Err    chan error
}

func (cb *appcb) HandleClientReady(c *celersdk.Client) {
	cb.Client <- c
}
func (cb *appcb) HandleClientInitErr(e *celersdkintf.E) {
	fmt.Println("init client err:", e)
	cb.Err <- fmt.Errorf("%s", e.Reason)
}

func (cb *appcb) HandleChannelOpened(token, cid string)                      {}
func (cb *appcb) HandleOpenChannelError(token, reason string)                {}
func (cb *appcb) HandleRecvStart(pay *celersdkintf.Payment)                  {}
func (cb *appcb) HandleRecvDone(pay *celersdkintf.Payment)                   {}
func (cb *appcb) HandleSendComplete(pay *celersdkintf.Payment)               {}
func (cb *appcb) HandleSendErr(pay *celersdkintf.Payment, e *celersdkintf.E) {}

func sleep(numSec int) {
	time.Sleep(time.Duration(numSec) * time.Second)
}
