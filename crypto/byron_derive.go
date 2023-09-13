package crypto

import (
	"crypto/hmac"
	"crypto/sha512"
	"github.com/echovl/ed25519"
	"math/big"
)

var (
	curveN = new(big.Int)
)

func init() {
	// The prime order for the base point.
	// N = 2^252 + 27742317777372353535851937790883648493
	qs, _ := new(big.Int).SetString("27742317777372353535851937790883648493", 10)
	curveN.SetBit(new(big.Int).SetInt64(0), 252, 1).Add(curveN, qs) // AKA Q
	//7237005577332262213973186563042994240857116359379907606001950938285454250989
	//fmt.Println("curveN:", curveN)
}

// Derive derives a children XPrv using BIP32-Ed25519
func (xsk XPrvKey) ByronDerive(index uint32) XPrvKey {
	xpriv := xsk[:64]
	chainCode := xsk[64:]
	zmac := hmac.New(sha512.New, chainCode)
	ccmac := hmac.New(sha512.New, chainCode)

	sindex := serializeBigIndex(index)
	//fmt.Println("bytes: ", hex.EncodeToString(xpriv), hex.EncodeToString(chainCode))
	if isHardenedDerivation(index) {
		zmac.Write([]byte{0x0})
		zmac.Write(xpriv)
		zmac.Write(sindex)
		ccmac.Write([]byte{0x1})
		ccmac.Write(xpriv)
		ccmac.Write(sindex)
	} else {
		pub := ed25519.PublicKeyFrom(ed25519.ExtendedPrivateKey(xpriv))
		zmac.Write([]byte{0x2})
		zmac.Write(pub)
		zmac.Write(sindex)
		ccmac.Write([]byte{0x3})
		ccmac.Write(pub)
		ccmac.Write(sindex)
	}
	z := zmac.Sum(nil)

	kl := NewByronPrivateKeyLeftPart(xsk[:32], z[:32])
	kr := NewPrivateKeyRightPart(xsk[32:64], z[32:64])
	cc := ccmac.Sum(nil)
	cc = cc[32:]

	cxsk := make([]byte, 96)
	copy(cxsk[:32], kl)
	copy(cxsk[32:64], kr)
	copy(cxsk[64:], cc)
	return cxsk
}

func serializeBigIndex(index uint32) []byte {
	return []byte{byte(index >> 24), byte(index >> 16), byte(index >> 8), byte(index)}
}

func NewByronPrivateKeyLeftPart(x, y []byte) []byte {
	// x 私钥 左半部分， y ByronDerive z的左半部分
	zl8 := make([]byte, len(y))
	for i := 0; i < len(y); i++ {
		zl8[i] = (y[i] * 8) & 0xFF
	}

	priv := new(big.Int).Add(new(big.Int).SetBytes(reverse(zl8)), new(big.Int).SetBytes(reverse(x)))
	priv.Mod(priv, curveN)
	return reverse(leftPadding(priv.Bytes(), 32))
}

func NewPrivateKeyRightPart(x, y []byte) []byte {
	// x 私钥 右半部分， y ByronDerive z的右半部分
	out := make([]byte, len(x))
	for i := 0; i < len(x); i++ {
		out[i] = (x[i] + y[i]) & 0xFF
	}
	return out
}

func leftPadding(in []byte, num int) []byte {
	if len(in) >= num {
		return in[:]
	}
	out := make([]byte, num)
	copy(out[num-len(in):], in[:])
	return out
}

func reverse(in []byte) []byte {
	l := len(in)
	out := make([]byte, l)
	for i := 0; i < l; i++ {
		out[l-1-i] = in[i]
	}
	return out
}
