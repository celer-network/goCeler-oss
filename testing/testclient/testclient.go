// Copyright 2018-2020 Celer Network

package main

import (
	"flag"
	"io/ioutil"

	"github.com/celer-network/goCeler/webapi"
	"github.com/celer-network/goutils/log"
)

var (
	grpcPort  = flag.Int("port", -1, "gRPC server listening port")
	ksPath    = flag.String("keystore", "", "Path to keystore json file")
	cfgPath   = flag.String("config", "profile.json", "Path to config json file")
	dataPath  = flag.String("datadir", "", "Path to the local database")
	extSigner = flag.Bool("extsign", false, "if set, exercise the external signer interface")
)

func main() {
	flag.Parse()
	ksBytes, err := ioutil.ReadFile(*ksPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := ioutil.ReadFile(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Infoln("start testclient on port", *grpcPort, "using ks", *ksPath)
	webapi.NewInternalApiServer(
		-1,
		*grpcPort,
		"http://localhost:*",
		string(ksBytes[:]),
		"",
		*dataPath,
		string(cfg[:]),
		*extSigner).Start()
}
