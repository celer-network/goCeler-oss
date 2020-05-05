// Copyright 2018-2020 Celer Network

package toolsetup

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payregistry"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payresolver"
	"github.com/celer-network/goCeler/chain/channel-eth-go/routerregistry"
	"github.com/celer-network/goCeler/chain/channel-eth-go/virtresolver"
	"github.com/celer-network/goCeler/chain/channel-eth-go/wallet"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/cobj"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/crypto/ssh/terminal"
)

func NewEthClient(profile *common.CProfile) *ethclient.Client {
	ethCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var wsOrigin string
	if profile.WsOrigin != "" {
		wsOrigin = profile.WsOrigin
	} else {
		wsOrigin = "http://celer.network"
	}
	var rpcClient *ethrpc.Client
	var err error
	ethInstance := profile.ETHInstance
	if strings.HasPrefix(ethInstance, "ws") {
		rpcClient, err = ethrpc.DialWebsocket(ethCtx, ethInstance, wsOrigin)
		if err != nil {
			// Retry once for stability.
			time.Sleep(time.Second)
			rpcClient, err = ethrpc.DialWebsocket(ethCtx, ethInstance, wsOrigin)
			if err != nil {
				log.Fatalf("DialETH failed: %s", err)
			}
		}
	} else {
		rpcClient, err = ethrpc.Dial(ethInstance)
		if err != nil {
			// Retry once for stability.
			time.Sleep(time.Second)
			rpcClient, err = ethrpc.Dial(ethInstance)
			if err != nil {
				log.Fatalf("DialETH failed: %s", err)
			}
		}
	}
	return ethclient.NewClient(rpcClient)
}

func NewDAL(profile *common.CProfile) *storage.DAL {
	var kvstore storage.KVStore
	var err error
	if profile.StoreSql != "" {
		db := profile.StoreSql
		log.Infof("Setting up server store at %s", db)
		kvstore, err = storage.NewKVStoreSQL("postgres", db)
		if err != nil {
			log.Fatalf("Cannot setup server store: %s: %s", db, err)
		}
	} else if profile.StoreDir != "" {
		dir := profile.StoreDir
		log.Infof("Setting up local store at %s", dir)
		fpath := filepath.Join(dir, "sqlite", "celer.db")
		kvstore, err = storage.NewKVStoreSQL("sqlite3", fpath)
		if err != nil {
			log.Fatalf("Cannot setup local store: %s: %s", dir, err)
		}
	} else {
		log.Fatalln("no database path found")
	}

	return storage.NewDAL(kvstore)
}

func NewNodeConfig(profile *common.CProfile, ethclient *ethclient.Client, dal *storage.DAL) *cobj.CelerGlobalNodeConfig {
	addr := ctype.Hex2Addr(profile.SvrETHAddr)
	return cobj.NewCelerGlobalNodeConfig(
		addr,
		ethclient,
		profile,
		wallet.CelerWalletABI,
		ledger.CelerLedgerABI,
		virtresolver.VirtContractResolverABI,
		payresolver.PayResolverABI,
		payregistry.PayRegistryABI,
		routerregistry.RouterRegistryABI,
		dal,
	)
}

func ParseKeyStoreFile(ksfile string, noPassword bool) (string, string) {
	var ksBytes []byte
	var keyStore string
	var passPhrase string
	var err error
	if ksfile != "" {
		ksBytes, err = ioutil.ReadFile(ksfile)
		if err != nil {
			log.Fatal(err)
		}
		keyStore = string(ksBytes)
	}
	if noPassword {
		passPhrase = ""
	} else {
		passPhrase = readPassword(ksBytes)
	}

	return keyStore, passPhrase
}

func readPassword(ksBytes []byte) string {
	ksAddress, err := getAddressFromKeystore(ksBytes)
	if err != nil {
		log.Fatal(err)
	}

	ksPasswordStr := ""
	if terminal.IsTerminal(syscall.Stdin) {
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
	return ksPasswordStr
}

func getAddressFromKeystore(ksBytes []byte) (string, error) {
	type ksStruct struct {
		Address string
	}
	var ks ksStruct
	if err := json.Unmarshal(ksBytes, &ks); err != nil {
		return "", err
	}
	return ks.Address, nil
}
