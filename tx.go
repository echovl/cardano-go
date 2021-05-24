package cardano

import (
	"encoding/hex"

	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/blake2b"
)

type protocolParams struct {
	MinimumUtxoValue uint64
	PoolDeposit      uint64
	KeyDeposit       uint64
	MinFeeA          uint64
	MinFeeB          uint64
}

type transactionID string

func (id transactionID) Bytes() []byte {
	bytes, err := hex.DecodeString(string(id))
	if err != nil {
		panic(err)
	}

	return bytes
}

type transaction struct {
	_          struct{} `cbor:",toarray"`
	Body       transactionBody
	WitnessSet transactionWitnessSet
	Metadata   *transactionMetadata // or null
}

func (tx *transaction) Bytes() []byte {
	bytes, err := cbor.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (tx *transaction) CborHex() string {
	return hex.EncodeToString(tx.Bytes())
}

func (tx *transaction) ID() transactionID {
	txHash := blake2b.Sum256(tx.Body.Bytes())
	return transactionID(hex.EncodeToString(txHash[:]))
}

type transactionWitnessSet struct {
	VKeyWitnessSet []vkeyWitness `cbor:"0,keyasint,omitempty"`
	// TODO: add optional fields 1-4
}

type vkeyWitness struct {
	_         struct{} `cbor:",toarray"`
	VKey      []byte   // ed25519 public key
	Signature []byte   // ed25519 signature
}

// Cbor map
type transactionMetadata map[uint64]transactionMetadatum

// This could be cbor map, array, int, bytes or a text
type transactionMetadatum struct{}

type transactionBody struct {
	Inputs       []transactionInput  `cbor:"0,keyasint"`
	Outputs      []transactionOutput `cbor:"1,keyasint"`
	Fee          uint64              `cbor:"2,keyasint"`
	Ttl          uint64              `cbor:"3,keyasint"`
	Certificates []certificate       `cbor:"4,keyasint,omitempty"` // Omit for now
	Withdrawals  *uint               `cbor:"5,keyasint,omitempty"` // Omit for now
	Update       *uint               `cbor:"6,keyasint,omitempty"` // Omit for now
	MetadataHash *uint               `cbor:"7,keyasint,omitempty"` // Omit for now
}

func (body *transactionBody) Bytes() []byte {
	bytes, err := cbor.Marshal(body)
	if err != nil {
		panic(err)
	}
	return bytes
}

type transactionInput struct {
	_     struct{} `cbor:",toarray"`
	ID    []byte   // HashKey 32 bytes
	Index uint64
}

type transactionOutput struct {
	_       struct{} `cbor:",toarray"`
	Address []byte
	Amount  uint64
}

// TODO: This should a cbor array with one element:
//  stake_registration
//	stake_deregistration
//	stake_delegation
//	pool_registration
//	pool_retirement
//	genesis_key_delegation
//	move_instantaneous_rewards_cert
type certificate struct{}
