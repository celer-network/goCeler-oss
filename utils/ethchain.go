// Copyright 2020 Celer Network

package utils

import (
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
)

func SigIsValid(signer ctype.Addr, data []byte, sig []byte) bool {
	recoveredAddr := RecoverSigner(data, sig)
	return recoveredAddr == signer
}

func RecoverSigner(data []byte, sig []byte) ctype.Addr {
	if len(sig) == 65 { // we could return zeroAddr if len not 65
		if sig[64] == 27 || sig[64] == 28 {
			// SigToPub only expect v to be 0 or 1,
			// see https://github.com/ethereum/go-ethereum/blob/v1.8.23/internal/ethapi/api.go#L468.
			// we've been ok as our own code only has v 0 or 1, but using external signer may cause issue
			// we also fix v in celersdk.PublishSignedResult to be extra safe
			sig[64] -= 27
		}
	}
	pubKey, err := crypto.SigToPub(GeneratePrefixedHash(data), sig)
	if err != nil {
		log.Errorf("sig error: %v, sig: %x", err, sig)
		return ctype.Addr{}
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr
}

func GeneratePrefixedHash(data []byte) []byte {
	return crypto.Keccak256([]byte("\x19Ethereum Signed Message:\n32"), crypto.Keccak256(data))
}

func GetAddrAndPrivKey(keyStore string, passPhrase string) (ctype.Addr, string, error) {
	key, err := keystore.DecryptKey([]byte(keyStore), passPhrase)
	if err != nil {
		return ctype.ZeroAddr, "", err
	}
	privKey := ctype.Bytes2Hex(crypto.FromECDSA(key.PrivateKey))
	return key.Address, privKey, nil
}
