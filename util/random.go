package util

import (
	"crypto/rand"
	"log"
	"math/big"
)

var (
	charBytes          = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsLen           = big.NewInt(int64(len(charBytes)))
	lowercaseCharBytes = "0123456789abcdefghijklmnopqrstuvwxyz"
	lowercaseCharsLen  = big.NewInt(int64(len(lowercaseCharBytes)))
)

// RandString returns a random string of length n
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		ix, err := rand.Int(rand.Reader, charsLen)
		if err != nil {
			log.Panic(err) // should not happen
		}
		b[i] = charBytes[int(ix.Int64())]
	}
	return string(b)
}

// RandLowercaseString returns a random string of length n
func RandLowercaseString(n int) string {
	b := make([]byte, n)
	for i := range b {
		ix, err := rand.Int(rand.Reader, lowercaseCharsLen)
		if err != nil {
			log.Panic(err) // should not happen
		}
		b[i] = lowercaseCharBytes[int(ix.Int64())]
	}
	return string(b)
}
