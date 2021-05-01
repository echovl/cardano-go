package crypto

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
)

func deriveChildPrivateKey(kp KeyPair, index uint32) ([]byte, error) {
	fmt.Println("kp", kp)
	xpriv := generateExtendedPrivateKey(kp.Priv)
	zmac := hmac.New(sha512.New, kp.ChainCode)
	ccmac := hmac.New(sha512.New, kp.ChainCode)

	sindex := serializeIndex(index)
	if isHardenedDerivation(index) {
		zmac.Write([]byte{0x0})
		zmac.Write(xpriv)
		zmac.Write(sindex)
		ccmac.Write([]byte{0x1})
		ccmac.Write(xpriv)
		ccmac.Write(sindex)
	} else {
	}

	z := zmac.Sum(nil)
	zl := z[:32]
	zr := z[32:]

	kl := add28Mul8(xpriv[:32], zl)
	kr := addMod256(xpriv[32:], zr)

	cc := ccmac.Sum(nil)
	cc = cc[32:]

	fmt.Println("z: ", zl, zr)

	return []byte{}, nil
}

func deriveChildPublicKey(kp KeyPair, index uint32) ([]byte, error) {
	return []byte{}, nil
}

func serializeIndex(index uint32) []byte {
	return []byte{byte(index), byte(index >> 8), byte(index >> 16), byte(index >> 24)}
}

func generateExtendedPrivateKey(priv []byte) []byte {
	xpriv := sha512.Sum512(priv)
	xpriv[0] &= 248
	xpriv[31] &= 127
	xpriv[31] |= 64

	return xpriv[:]
}

func isHardenedDerivation(index uint32) bool {
	return index >= 0x80000000
}

func add28Mul8(x, y []byte) []byte {
	out := make([]byte, 32)

	ix := binary.LittleEndian.Uint32(x)
	iy := binary.LittleEndian.Uint32(y[:28])
	binary.LittleEndian.PutUint32(out, ix+8*iy)

	return out
}

func addMod256(x, y []byte) []byte {
	out := make([]byte, 32)

	ix := binary.LittleEndian.Uint32(x)
	iy := binary.LittleEndian.Uint32(y)
	binary.LittleEndian.PutUint32(out, ix+iy)

	return out
}
