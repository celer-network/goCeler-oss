// Copyright 2018-2020 Celer Network

package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils/bar"
	"github.com/celer-network/goutils/jsonpbhex"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/jsonpb"
	proto "github.com/golang/protobuf/proto"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// set EmitDefaults so json has complete schema
// use our own AnyResolver so unknown Any can be properly
// marshaled to base64 string
var jsonpbMarshaler = jsonpb.Marshaler{
	EmitDefaults: true,
	AnyResolver:  bar.BetterAnyResolver,
}

var jsonpbHex = jsonpbhex.Marshaler{
	HexBytes:    true,
	AnyResolver: bar.BetterAnyResolver,
}

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

func ValidateAndFormatAddress(address string) (ctype.Addr, error) {
	if !ethcommon.IsHexAddress(address) {
		return ctype.ZeroAddr, errors.New("INVALID_ADDRESS")
	}
	return ctype.Hex2Addr(address), nil
}

// GetTokenAddr returns token address
func GetTokenAddr(tokenInfo *entity.TokenInfo) ctype.Addr {
	switch tktype := tokenInfo.TokenType; tktype {
	case entity.TokenType_ETH:
		return ctype.EthTokenAddr
	case entity.TokenType_ERC20:
		return ctype.Bytes2Addr(tokenInfo.TokenAddress)
	}
	return ctype.InvalidTokenAddr
}

// GetTokenAddrStr returns string for tokenInfo
func GetTokenAddrStr(tokenInfo *entity.TokenInfo) string {
	return ctype.Addr2Hex(GetTokenAddr(tokenInfo))
}

func PrintToken(tokenInfo *entity.TokenInfo) string {
	if tokenInfo.GetTokenType() == entity.TokenType_ETH {
		return "ETH"
	}
	return GetTokenAddrStr(tokenInfo)
}

func PrintTokenAddr(tkaddr ctype.Addr) string {
	if tkaddr == ctype.EthTokenAddr {
		return "ETH"
	}
	return ctype.Addr2Hex(tkaddr)
}

// GetTokenInfoFromAddress returns TokenInfo from tkaddr
// only support ERC20 for now
func GetTokenInfoFromAddress(tkaddr ctype.Addr) *entity.TokenInfo {
	tkInfo := new(entity.TokenInfo)
	if tkaddr == ctype.EthTokenAddr {
		tkInfo.TokenType = entity.TokenType_ETH
	} else {
		tkInfo.TokenType = entity.TokenType_ERC20
		tkInfo.TokenAddress = tkaddr.Bytes()
	}
	return tkInfo
}

// Uint64ToBytes converts uint to bytes in big-endian order.
func Uint64ToBytes(i uint64) []byte {
	ret := make([]byte, 8) // 8 bytes for uint64
	binary.BigEndian.PutUint64(ret, i)
	return ret
}

// GetTsAndSig returns current time and signature of current time using sign param passed in.
func GetTsAndSig(sign func([]byte) []byte) (ts uint64, sig []byte) {
	ts = uint64(time.Now().Unix())
	sig = sign(Uint64ToBytes(ts))
	return ts, sig
}

// PbToJSONString marshals a protobuf msg to json string
// Note we set EmitDefaults so json is always complete.
// If you think you have a use case for omit default in json,
// check w/ junda first before adding another func.
// The marshaler also uses our own BetterAnyResolver which handles unknown Any
// msg instead of throw error
func PbToJSONString(pb proto.Message) (string, error) {
	return jsonpbMarshaler.MarshalToString(pb)
}

// PbToJSONHexBytes output hex for bytes field instead of default base64
// WARNING: result json not compatible for unmarshal
// only use this for logging purpose
func PbToJSONHexBytes(pb proto.Message) (string, error) {
	return jsonpbHex.MarshalToString(pb)
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

func UnmarshalDelegationDescription(proof *rpc.DelegationProof) (*rpc.DelegationDescription, error) {
	if proof == nil {
		return nil, errors.New("nil delegation proof")
	}

	var desc rpc.DelegationDescription
	err := proto.Unmarshal(proof.GetDelegationDescriptionBytes(), &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
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
