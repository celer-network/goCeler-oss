// Try run with different flags. Example:
// go run clog/test/clogtest.go -logcolor -loglevel=debug
// go run clog/test/clogtest.go -loglevel=warn -loglocaltime -loglongfile

package main

import (
	"flag"

	log "github.com/celer-network/goCeler-oss/clog"
)

func main() {
	flag.Parse()
	log.Trace("trace every step")
	log.Debug("looking into what's really happening")
	log.Infof("x is set to %d", 2)
	log.Warnln("watch out!", "enemy is coming!")
	log.Error("something is wrong")
	log.Fatal("get me out of here!")
}
