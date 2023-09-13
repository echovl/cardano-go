package byron

import (
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/internal/cbor"
	"golang.org/x/crypto/blake2b"
)

// Class container for Cardano Byron legacy BIP32 constants.

const (
	// HMAC message format
	HMAC_MESSAGE_FORMAT            = "Root Seed Chain %d"
	TWEAK_BIT1              uint8  = 0x07
	TWEAK_BIT2              uint8  = 0x80
	INDEF_LEN_ARRAY_START   byte   = 0x9F
	INDEF_LEN_ARRAY_END     byte   = 0xFF
	HD_PATH_KEY_PBKDF2_SALT string = "address-hashing"
	//PBKDF2 rounds used for deriving the HD path key
	HD_PATH_KEY_PBKDF2_ROUNDS int = 500
	//PBKDF2 output byte length used for deriving the HD path key
	HD_PATH_KEY_PBKDF2_OUT_BYTE_LEN int = 32
)

var (
	// ChaCha20-Poly1305 associated data for HD path decryption/encryption
	CHACHA20_POLY1305_ASSOC_DATA []byte = []byte("")
	// ChaCha20-Poly1305 nonce for HD path decryption/encryption
	CHACHA20_POLY1305_NONCE []byte = []byte("serokellfore")
)

func ByronLegacyFromSeed(entropy []byte) (crypto.XPrvKey, error) {
	entropy, err := cbor.Marshal(entropy)
	if err != nil {
		return nil, err
	}
	entropy2 := blake2b.Sum256(entropy)
	entropy, err = cbor.Marshal(entropy2[:])
	if err != nil {
		return nil, err
	}
	//fmt.Println(err, hex.EncodeToString(entropy))
	il, ir := HashRepeatedly(entropy[:], 1)
	return append(append([]byte{}, il...), ir...), nil
}

func HashRepeatedly(key []byte, itrNum int) ([]byte, []byte) {
	h := hmac.New(sha512.New, key)
	h.Write([]byte(fmt.Sprintf(HMAC_MESSAGE_FORMAT, itrNum)))
	data := h.Sum(nil)
	//fmt.Println("sum data:", hex.EncodeToString(data))

	il, ir := data[:len(data)/2], data[len(data)/2:]
	ilSha := sha512.Sum512(il)
	il = TweakMasterKeyBits(ilSha[:])
	if (il[31] & 0x20) != 0 {
		// fmt.Println("HashRepeatedly ", itrNum, il[31], il[31] & 0x20)
		return HashRepeatedly(key, itrNum+1)
	}
	return il, ir
}

func TweakMasterKeyBits(key []byte) []byte {
	// Clear the lowest 3 bits of the first byte of kL
	key[0] = key[0] & ^TWEAK_BIT1
	// Clear the highest bit of the last byte of kL
	key[31] = key[31] & ^TWEAK_BIT2
	// Set the second-highest bit of the last byte of kL
	key[31] = key[31] | 0x40
	return key
}
