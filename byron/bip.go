package byron

import (
	"crypto/sha512"
	"github.com/echovl/cardano-go/internal/cbor"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
)

func ChaCha20Poly1305EncodePath(hdPath []byte, path []uint32) ([]byte, error) {
	cpath, err := CborIndefiniteLenArrayEncoder(path)
	if err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.New(hdPath)
	if err != nil {
		return nil, err
	}
	return aead.Seal(nil, CHACHA20_POLY1305_NONCE, cpath, nil), nil
}

func CborIndefiniteLenArrayEncoder(path []uint32) ([]byte, error) {
	ret := make([]byte, 0)
	ret = append(ret, INDEF_LEN_ARRAY_START)
	for _, p := range path {
		cp, err := cbor.Marshal(p)
		if err != nil {
			return nil, err
		}
		ret = append(ret, cp...)
	}
	ret = append(ret, INDEF_LEN_ARRAY_END)
	return ret, nil
}

// root xpub contians public & chaincode
func HdPathKey(rootXpub []byte) []byte {
	return pbkdf2.Key(
		rootXpub,
		[]byte(HD_PATH_KEY_PBKDF2_SALT),
		HD_PATH_KEY_PBKDF2_ROUNDS,
		HD_PATH_KEY_PBKDF2_OUT_BYTE_LEN,
		sha512.New)
}
