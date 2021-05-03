package crypto

import (
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
)

func DerivePrivateKey(xpriv XPriv, index uint32) (XPriv, error) {
	zmac := hmac.New(sha512.New, xpriv[64:])
	ccmac := hmac.New(sha512.New, xpriv[64:])

	sindex := serializeIndex(index)
	if isHardenedDerivation(index) {
		fmt.Printf("hardened!")
		zmac.Write([]byte{0x0})
		zmac.Write(xpriv[:64])
		zmac.Write(sindex)
		ccmac.Write([]byte{0x1})
		ccmac.Write(xpriv[:64])
		ccmac.Write(sindex)
	} else {
	}

	z := zmac.Sum(nil)
	zl := z[:32]
	zr := z[32:]

	kl := add28Mul8(xpriv[:32], zl)
	kr := addMod256(xpriv[32:64], zr)

	cc := ccmac.Sum(nil)
	cc = cc[32:]

	childXpriv := make([]byte, 96)

	copy(childXpriv[:32], kl)
	copy(childXpriv[32:64], kr)
	copy(childXpriv[64:], cc)

	return childXpriv, nil
}

func deriveChildPublicKey(kp KeyPair, index uint32) ([]byte, error) {
	return []byte{}, nil
}

func serializeIndex(index uint32) []byte {
	return []byte{byte(index), byte(index >> 8), byte(index >> 16), byte(index >> 24)}
}

func isHardenedDerivation(index uint32) bool {
	return index >= 0x80000000
}

func add28Mul8(x, y []byte) []byte {
	out := make([]byte, 32)
	var carry uint16

	for i, xi := range x[:28] {
		r := uint16(xi) + ((uint16(y[i])) << 3) + carry
		out[i] = byte(r & 0xff)
		carry = r >> 8
	}
	for i, xi := range x[28:32] {
		r := uint16(xi) + carry
		out[i] = byte(r & 0xff)
		carry = r >> 8
	}

	return out
}

func addMod256(x, y []byte) []byte {
	out := make([]byte, 32)
	var carry uint16

	for i, xi := range x[:32] {
		r := uint16(xi) + uint16(y[i]) + carry
		out[i] = byte(r & 0xff)
		carry = r >> 8

	}

	return out
}
