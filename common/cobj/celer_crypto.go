// Copyright 2018-2019 Celer Network

package cobj

import (
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/ctype"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type CelerCrypto struct {
	privateKey string
	publicKey  string
}

func NewCelerCrypto(privateKey string, publicKey string) *CelerCrypto {
	c := &CelerCrypto{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
	return c
}

func (c *CelerCrypto) Sign(data []byte) ([]byte, error) {
	key, err := crypto.HexToECDSA(c.privateKey)
	if err != nil {
		return nil, err
	}
	sig, err := crypto.Sign(generatePrefixedHash(data), key)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (c *CelerCrypto) SigIsValid(signer string, data []byte, sig []byte) bool {
	recoveredAddr := RecoverSigner(data, sig)
	expectedAddr := ctype.Hex2Addr(signer)
	return recoveredAddr == expectedAddr
}

func RecoverSigner(data []byte, sig []byte) ethcommon.Address {
	pubKey, err := crypto.SigToPub(generatePrefixedHash(data), sig)
	if err != nil {
		log.Error(err)
		return ethcommon.Address{}
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr
}

func generatePrefixedHash(data []byte) []byte {
	return crypto.Keccak256([]byte("\x19Ethereum Signed Message:\n32"), crypto.Keccak256(data))
}
