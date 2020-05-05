// Copyright 2018-2020 Celer Network

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/webapi"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	webPort        = flag.Int("port", 29979, "websocket server listening port")
	grpcPort       = flag.Int("grpcport", -1, "gRPC server listening port")
	ksPath         = flag.String("keystore", "", "Path to keystore json file")
	cfgPath        = flag.String("config", "profile.json", "Path to config json file")
	dataPath       = flag.String("datadir", "", "Path to the local database")
	allowedOrigins = flag.String(
		"allowedorigins", "file://*,http://localhost:*", "Comma-separated list of allowed origins")
	password = flag.String(
		"password", "", "Password for the keystore. Prefer terminal input for better security")
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
	ksAddress, err := utils.GetAddressFromKeystore(ksBytes)
	if err != nil {
		log.Fatal(err)
	}
	ksPasswordStr := ""
	if flag.Lookup("password") != nil {
		ksPasswordStr = *password
	} else if terminal.IsTerminal(syscall.Stdin) {
		fmt.Printf("Enter password for %s: ", ksAddress)
		ksPassword, err2 := terminal.ReadPassword(syscall.Stdin)
		if err2 != nil {
			log.Fatalln("Cannot read password from terminal:", err2)
		}
		ksPasswordStr = string(ksPassword)
	} else {
		reader := bufio.NewReader(os.Stdin)
		ksPwd, err2 := reader.ReadString('\n')
		if err2 != nil {
			log.Fatalln("Cannot read password from stdin:", err2)
		}
		ksPasswordStr = strings.TrimSuffix(ksPwd, "\n")
	}
	_, err = keystore.DecryptKey(ksBytes, ksPasswordStr)
	if err != nil {
		log.Fatal(err)
	}
	webapi.NewApiServer(
		*webPort,
		*grpcPort,
		*allowedOrigins,
		string(ksBytes[:]),
		ksPasswordStr,
		*dataPath,
		string(cfg[:]), false).Start()
}
