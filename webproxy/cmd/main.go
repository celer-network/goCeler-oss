// Copyright 2018-2019 Celer Network

package main

import (
	"flag"

	"github.com/celer-network/goCeler-oss/webproxy"
)

var (
	port                 = flag.Int("port", 29980, "proxy listening port")
	serverNetworkAddress = flag.String("server", "", "server network address")
)

func main() {
	flag.Parse()
	webproxy.NewProxy(*port, *serverNetworkAddress).Start()
}
