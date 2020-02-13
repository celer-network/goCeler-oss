package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"syscall"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"golang.org/x/crypto/ssh/terminal"
)

// WeiToEthStr returns easy to read xxx.xxx string in ETH unit
func WeiToEthStr(wei *big.Int) string {
	if wei.Cmp(big.NewInt(0)) == 0 {
		return "0"
	}

	weiStr := wei.String()
	weiStrLen := len(weiStr)
	if weiStrLen > 18 {
		fractionalStr := strings.TrimRight(weiStr[weiStrLen-18:], "0")
		if fractionalStr == "" {
			return weiStr[:weiStrLen-18]
		}

		return weiStr[:weiStrLen-18] + "." + fractionalStr
	}

	return "0." + strings.Repeat("0", 18-weiStrLen) + strings.TrimRight(weiStr, "0")
}

// Wei2BigInt convert decimal wei string to big.Int
func Wei2BigInt(wei string) *big.Int {
	i := big.NewInt(0)
	_, ok := i.SetString(wei, 10)
	if !ok {
		return nil
	}
	return i
}

func ChkErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func ParseBackendProfileJSON(path string) *common.ProfileJSON {
	pj := new(common.ProfileJSON)
	raw, err := ioutil.ReadFile(path)
	ChkErr(err)
	err = json.Unmarshal(raw, pj)
	ChkErr(err)
	return pj
}

func GetKeyStore(keyStoreFile string) (string, string) {
	data, err := ioutil.ReadFile(keyStoreFile)
	if err != nil {
		log.Fatalln(err)
	}
	type ksStruct struct {
		Address string
	}
	var ks ksStruct
	json.Unmarshal(data, &ks)
	return string(data), ks.Address
}

func GetStringFromStdin(hintText string, confidential bool) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(hintText)

	var s string
	var err error
	if confidential && terminal.IsTerminal(syscall.Stdin) {
		byteS, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalln(err)
		}
		s = string(byteS)
	} else {
		s, err = reader.ReadString('\n')
		if err != nil {
			log.Fatalln(err)
		}
		s = strings.TrimSuffix(s, "\n")
	}
	fmt.Println()

	return s
}
