package cardano

import (
	"encoding/hex"

	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/blake2b"
)

type ProtocolParams struct {
	MinimumUtxoValue uint64
	PoolDeposit      uint64
	KeyDeposit       uint64
	MinFeeA          uint64
	MinFeeB          uint64
}

type Transaction struct {
	_          struct{} `cbor:",toarray"`
	Body       TransactionBody
	WitnessSet TransactionWitnessSet
	Metadata   *TransactionMetadata // or null
}

func (tx *Transaction) Bytes() []byte {
	bytes, err := cbor.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (tx *Transaction) CborHex() string {
	return hex.EncodeToString(tx.Bytes())
}

func (tx *Transaction) ID() TxId {
	txHash := blake2b.Sum256(tx.Body.Bytes())
	return TxId(hex.EncodeToString(txHash[:]))
}

type TransactionWitnessSet struct {
	VKeyWitnessSet []VKeyWitness `cbor:"0,keyasint,omitempty"`
	// TODO: add optional fields 1-4
}

type VKeyWitness struct {
	_         struct{} `cbor:",toarray"`
	VKey      []byte   // ed25519 public key
	Signature []byte   // ed25519 signature
}

// Cbor map
type TransactionMetadata map[uint64]TransactionMetadatum

// This could be cbor map, array, int, bytes or a text
type TransactionMetadatum struct{}

type TransactionBody struct {
	Inputs       []TransactionInput  `cbor:"0,keyasint"`
	Outputs      []TransactionOutput `cbor:"1,keyasint"`
	Fee          uint64              `cbor:"2,keyasint"`
	Ttl          uint64              `cbor:"3,keyasint"`
	Certificates []Certificate       `cbor:"4,keyasint,omitempty"` // Omit for now
	Withdrawals  *uint               `cbor:"5,keyasint,omitempty"` // Omit for now
	Update       *uint               `cbor:"6,keyasint,omitempty"` // Omit for now
	MetadataHash *uint               `cbor:"7,keyasint,omitempty"` // Omit for now
}

func (body *TransactionBody) Bytes() []byte {
	bytes, err := cbor.Marshal(body)
	if err != nil {
		panic(err)
	}
	return bytes
}

type TransactionInput struct {
	_     struct{} `cbor:",toarray"`
	ID    []byte   // HashKey 32 bytes
	Index uint64
}

type TransactionOutput struct {
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
type Certificate struct{}
