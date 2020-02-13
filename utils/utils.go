// Copyright 2018 Celer Network

package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/protobuf/jsonpb"
	proto "github.com/golang/protobuf/proto"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Dec2HexStr decimal string to hex
func Dec2HexStr(dec string) string {
	i := new(big.Int)
	i.SetString(dec, 10)
	return i.Text(16)
}

// Hex2DecStr hex string to decimal
func Hex2DecStr(hex string) string {
	i := new(big.Int)
	i.SetString(hex, 16)
	return i.Text(10)
}

func BytesToBigInt(in []byte) *big.Int {
	ret := big.NewInt(0)
	ret.SetBytes(in)
	return ret
}

// convert decimal wei string to big.Int
func Wei2BigInt(wei string) *big.Int {
	i := big.NewInt(0)
	_, ok := i.SetString(wei, 10)
	if !ok {
		return nil
	}
	return i
}

// float in 10e18 wei to wei
func Float2Wei(f float64) *big.Int {
	if f < 0 {
		return nil
	}
	wei := decimal.NewFromFloat(f).Mul(decimal.NewFromFloat(10).Pow(decimal.NewFromFloat(18)))
	weiInt := new(big.Int)
	weiInt.SetString(wei.String(), 10)
	return weiInt
}

// left padding
func Pad(origin []byte, n int) []byte {
	m := len(origin)
	padded := make([]byte, n)
	pn := n - m
	for i := m - 1; i >= 0; i-- {
		padded[pn+i] = origin[i]
	}
	return padded
}

func TryLock(m *sync.Mutex) bool {
	const mutexLocked = 1 << iota
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(m)), 0, mutexLocked)
}

// use for celer client dialing celer server/proxy
// support os ca and celer ca
func GetClientTlsOption() grpc.DialOption {
	cpool, _ := x509.SystemCertPool()
	if cpool == nil {
		cpool = x509.NewCertPool()
	}
	cpool.AppendCertsFromPEM(CelerCA)
	if sdkCert != nil && sdkKey != nil {
		cert, _ := tls.X509KeyPair(sdkCert, sdkKey)
		creds := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      cpool,
		})
		return grpc.WithTransportCredentials(creds)
	}
	return grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(cpool, ""))
}

// GetClientTlsConfig returns tls.Config with system and celerCA, for https interaction
func GetClientTlsConfig() *tls.Config {
	cpool, _ := x509.SystemCertPool()
	if cpool == nil {
		cpool = x509.NewCertPool()
	}
	cpool.AppendCertsFromPEM(CelerCA)
	return &tls.Config{
		RootCAs: cpool,
	}
}

func ValidateAndFormatAddress(address string) (string, error) {
	if !ethcommon.IsHexAddress(address) {
		return "", errors.New("Invalid address")
	}
	return ctype.Bytes2Hex(ctype.Hex2Bytes(address)), nil
}

// GetTokenAddrStr returns string for tokenInfo
func GetTokenAddrStr(tokenInfo *entity.TokenInfo) string {
	switch tktype := tokenInfo.TokenType; tktype {
	case entity.TokenType_ETH:
		return common.EthContractAddr
	case entity.TokenType_ERC20:
		return ctype.Bytes2Hex(tokenInfo.TokenAddress)
	}
	return ""
}

// GetTokenInfoFromAddress returns TokenInfo from tkaddr
// only support ERC20 for now
func GetTokenInfoFromAddress(tkaddr ethcommon.Address) *entity.TokenInfo {
	tkInfo := new(entity.TokenInfo)
	if tkaddr == ctype.ZeroAddr {
		tkInfo.TokenType = entity.TokenType_ETH
	} else {
		tkInfo.TokenType = entity.TokenType_ERC20
		tkInfo.TokenAddress = tkaddr.Bytes()
	}
	return tkInfo
}

func WaitMined(ctx context.Context, ec *ethclient.Client,
	tx *ethtypes.Transaction, blockDelay uint64) (*ethtypes.Receipt, error) {
	return WaitMinedWithTxHash(ctx, ec, tx.Hash().Hex(), blockDelay)
}

// WaitMined waits for tx to be mined on the blockchain
// It returns tx receipt when the tx has been mined and enough block confirmations have passed
func WaitMinedWithTxHash(ctx context.Context, ec *ethclient.Client,
	txHash string, blockDelay uint64) (*ethtypes.Receipt, error) {
	// an error possibly returned when a transaction is pending
	const missingFieldErr = "missing required field 'transactionHash' for Log"

	if ec == nil {
		return nil, errors.New("nil ethclient")
	}
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()
	// wait tx to be mined
	txHashBytes := ethcommon.HexToHash(txHash)
	for {
		receipt, rerr := ec.TransactionReceipt(ctx, txHashBytes)
		if rerr == nil {
			log.Debugf("Transaction mined. Waiting for %d block confirmations", blockDelay)
			if blockDelay == 0 {
				return receipt, rerr
			}
			break
		} else if rerr == ethereum.NotFound || rerr.Error() == missingFieldErr {
			// Wait for the next round
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-queryTicker.C:
			}
		} else {
			return receipt, rerr
		}
	}
	// wait for enough block confirmations
	ddl := big.NewInt(0)
	latestBlockHeader, err := ec.HeaderByNumber(ctx, nil)
	if err == nil {
		ddl.Add(new(big.Int).SetUint64(blockDelay), latestBlockHeader.Number)
	}
	for {
		latestBlockHeader, err := ec.HeaderByNumber(ctx, nil)
		if err == nil && ddl.Cmp(latestBlockHeader.Number) < 0 {
			receipt, rerr := ec.TransactionReceipt(ctx, txHashBytes)
			if rerr == nil {
				log.Debugln("tx confirmed!")
				return receipt, rerr
			} else if rerr == ethereum.NotFound || rerr.Error() == missingFieldErr {
				return nil, errors.New("tx is dropped due to chain re-org")
			} else {
				return receipt, rerr
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
}

// Serialize a protobuf to json string
func PbToJSONString(pb proto.Message) string {
	m := jsonpb.Marshaler{}
	ret, err := m.MarshalToString(pb)
	if err != nil {
		log.Errorln("pb2json err: ", err, pb)
		return ""
	}
	return ret
}

func GetAddressFromKeystore(ksBytes []byte) (string, error) {
	type ksStruct struct {
		Address string
	}
	var ks ksStruct
	if err := json.Unmarshal(ksBytes, &ks); err != nil {
		return "", err
	}
	return ks.Address, nil
}
func RequestBuildRoutingTable(adminHTTPAddr string) error {
	url := fmt.Sprintf("http://%s/admin/route/build", adminHTTPAddr)
	log.Debugln("URL:>", url)
	reqJSONByte, err := json.Marshal(&rpc.BuildRoutingTableRequest{})
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSONByte))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return errors.New("Status is " + resp.Status)
	}
	return nil
}

// CelerCA root CA file, generated via certstrap
// new CA file w/ 10yr expiration
var CelerCA = []byte(`-----BEGIN CERTIFICATE-----
MIIE5DCCAsygAwIBAgIBATANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQDEwdDZWxl
ckNBMB4XDTE5MDkxNzIwMTEyNFoXDTI5MDkxNzIwMTEyNFowEjEQMA4GA1UEAxMH
Q2VsZXJDQTCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAJMk0gwPrlv/
/8aVaKAoURvPMHEgpDyj8c08A55kjNlStgjtpBpBEfNiLg3nIBeyrrmt/9QXLexD
RxewLgm/9uUB/U1Q4ZBQdAdsSHTyU+wap1HKJ7GX1jhZMjY95vIfTbrSIbo2clym
zKlbwvvBlNYgweHz0YiyWUCsqi2wH++ybUNzmgs0qI+lE/Fg4k8sReVcix5rUvNF
na9tpvGdV9u+iZNlwkeb3Hp9Ank5MR0830LzG2uf95p+d0fXmfl92wxdAFWnEhWi
uPK4Zfqt2orTIpY1uhiDl4d4kf1p0Niowf9FNOHMURYbTQqFMGFLOZI7+dOPW8Wy
AfkcZcgBfEQ2rGkd3+kb8A2pOTBaFqG9HkspKe9d/dXKKZMW63nLU1MIERWmj2S+
uEBAObnNCpmWPDDFnUpNaACt76tqRP+jYaOaEdPp8svOB6mD47lQemb2SEuWd8oa
afbv/tS6rdRvJEPJ2PgHSzIrYG3cTrLrNDhjFa3CoPwilWtP414+AfvwZTRfjNWy
kcVjK8kurHmrhNzhYNC/rtaQquU2NwS0UYf7+sRLytgEK6+6HahHXJD7/6aRyzrm
1bwYANmgrWf7LfvEf4ezGb1m+qLU3B2/Lzh3HpEBw/ySTFXNHKtv7ZbsN+ccjez6
au3fgns4jO8nCF7rAJLChopKHMkGeLkdAgMBAAGjRTBDMA4GA1UdDwEB/wQEAwIB
BjASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBSoQm3esw5att/hencMPoPd
qhyvXTANBgkqhkiG9w0BAQsFAAOCAgEAIJATh5yJd7XzfnfM8f2MNKRbUWzdDhE8
AHPjtKoOCsKOyua41xIpPvM+emdg8oTOZdNRlaoDiO1/8DB7PU1k1iXFaZ/MrgeM
Cz8pP9MvXLSXmg039hYREWV7pFvdbqhvfnOU+pj/uMwif1pl6+CRDxxSdwqUNeJr
gmqbDFBvdRa5DQJm7rbIYpSMc5P/GHZcVgOb+g3y6iODaPL/VR7Uo1xVvxzjgxpI
09QcYiDNK5vPondgaoh7W3c+KuaEKO18G8TEN0NGFOadk5ZjJ9uq+8aGfy51qny8
SMOI5/wW+7HODeQmSqtaxVlhZdmWa/iIzya/NGe+5JhRKgBKN9BIysEiVc4i4ver
utwMnSqqDSCZKUD6FeJn+CUimDf9nb9xbsZ8a+5pw2D6/iaZ+mJd1Pv0vHX5NxMJ
36Rj0MMB9I1xY9C2/ugiP1a/JG+Ve4n1r4GX1S2MfYH/k8wYcs4cVLQ21nphvTW5
osJePOuWfBuWD77selYHU/PhlzNVq2bSWHDQlQJoQr12dGk0NiYAf0FtTWRQoMkq
nwCu157ZSeK2bWffJUcLnTFV63ftZmsqEjYHHVQrbthc+LBpTT4ZYOBerEfQfoKi
c6Z63v4v3R9WB5VYZWH7nh+lMJBPhhL1043iN4Be3Z27GJ0jKIPQAL2gfNiukuz/
D/qaayaXjbo=
-----END CERTIFICATE-----`)

var sdkCert []byte
var sdkKey []byte
