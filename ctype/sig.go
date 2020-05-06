// Copyright 2020 Celer Network

package ctype

// this file defines Sig as [65]byte and funcs
const sigLength = 65

// Sig is r,s,v
type Sig [sigLength]byte

// ZeroSig is all 0, indicate invalid Sig
var ZeroSig Sig

// Bytes returns a new byte slice from s content
func (s Sig) Bytes() []byte { return s[:] }

// Hex returns hex string w/o 0x prefix
func (s Sig) Hex() string {
	return Bytes2Hex(s[:])
}

// Bytes2Sig create a new Sig based on b's content
// if len(b) isn't 65, return ZeroSig
// Returned Sig has s[64] fixed, ie. 27/28->0/1
func Bytes2Sig(b []byte) Sig {
	if len(b) != sigLength {
		return ZeroSig
	}
	var s Sig
	copy(s[:], b)
	if s[64] == 27 || s[64] == 28 {
		s[64] -= 27
	}
	return s
}
