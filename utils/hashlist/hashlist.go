// Copyright 2018-2020 Celer Network

package hashlist

import (
	"bytes"
	"errors"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
)

// DeleteHash deletes hash from list
func DeleteHash(list [][]byte, hash []byte) ([][]byte, error) {
	for i, e := range list {
		if bytes.Equal(e, hash) {
			return append(list[:i], list[i+1:]...), nil
		}
	}
	return list, common.ErrPayNotFound
}

func Exist(list [][]byte, hash []byte) bool {
	for _, e := range list {
		if bytes.Equal(e, hash) {
			return true
		}
	}
	return false
}

// SymmetricDifference returns list difference A - B
func Difference(a, b [][]byte) ([][]byte, error) {
	setA := make(map[string]bool)
	setB := make(map[string]bool)
	var diff [][]byte
	for _, hash := range b {
		key := ctype.Bytes2Hex(hash)
		if _, ok := setB[key]; ok {
			return nil, errors.New("list B has non-unique element")
		}
		setB[key] = true
	}
	for _, hash := range a {
		key := ctype.Bytes2Hex(hash)
		if _, ok := setA[key]; ok {
			return nil, errors.New("list A has non-unique element")
		}
		setA[key] = true
		if _, ok := setB[key]; !ok {
			diff = append(diff, hash)
		}
	}
	return diff, nil
}

// SymmetricDifference returns list difference A - B, B - A, error if A or B has non-unique element
func SymmetricDifference(a, b [][]byte) ([][]byte, [][]byte, error) {
	setA := make(map[string]bool)
	setB := make(map[string]bool)
	var diffAB [][]byte
	var diffBA [][]byte
	for _, hash := range b {
		key := ctype.Bytes2Hex(hash)
		if _, ok := setB[key]; ok {
			return nil, nil, errors.New("list B has non-unique element")
		}
		setB[key] = true
	}
	for _, hash := range a {
		key := ctype.Bytes2Hex(hash)
		if _, ok := setA[key]; ok {
			return nil, nil, errors.New("list A has non-unique element")
		}
		setA[key] = true
		if _, ok := setB[key]; !ok {
			diffAB = append(diffAB, hash)
		}
	}
	for _, hash := range b {
		if _, ok := setA[ctype.Bytes2Hex(hash)]; !ok {
			diffBA = append(diffBA, hash)
		}
	}
	return diffAB, diffBA, nil
}
