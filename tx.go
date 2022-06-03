package cardano

import (
	"encoding/hex"
	"math"

	"github.com/echovl/cardano-go/crypto"
	"golang.org/x/crypto/blake2b"
)

const utxoEntrySizeWithoutVal = 27

type UTxO struct {
	TxHash  Hash32
	Spender Address
	Amount  *Value
	Index   uint64
}

type Tx struct {
	_             struct{} `cbor:",toarray"`
	Body          TxBody
	WitnessSet    WitnessSet
	IsValid       bool
	AuxiliaryData *AuxiliaryData // or null
}

// Bytes returns the CBOR encoding of the transaction as bytes.
func (tx *Tx) Bytes() []byte {
	bytes, err := cborEnc.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return bytes
}

// Hex returns the CBOR encoding of the transaction as hex.
func (tx Tx) Hex() string {
	return hex.EncodeToString(tx.Bytes())
}

// Hash returns the transaction body hash using blake2b.
func (tx *Tx) Hash() (Hash32, error) {
	return tx.Body.Hash()
}

type WitnessSet struct {
	VKeyWitnessSet []VKeyWitness  `cbor:"0,keyasint,omitempty"`
	Scripts        []NativeScript `cbor:"1,keyasint,omitempty"`
}

type VKeyWitness struct {
	_         struct{}      `cbor:",toarray"`
	VKey      crypto.PubKey // ed25519 public key
	Signature []byte        // ed25519 signature
}

type TxInput struct {
	_      struct{} `cbor:",toarray"`
	TxHash Hash32
	Index  uint64
	Amount *Value `cbor:"-"`
}

// NewTxInput creates a new instance of TxInput
func NewTxInput(txHash Hash32, index uint, amount *Value) *TxInput {
	return &TxInput{TxHash: txHash, Index: uint64(index), Amount: amount}
}

type TxOutput struct {
	_       struct{} `cbor:",toarray"`
	Address Address
	Amount  *Value
}

// NewTxOutput creates a new instance of TxOutput
func NewTxOutput(addr Address, amount *Value) *TxOutput {
	return &TxOutput{Address: addr, Amount: amount}
}

type TxBody struct {
	Inputs  []*TxInput  `cbor:"0,keyasint"`
	Outputs []*TxOutput `cbor:"1,keyasint"`
	Fee     Coin        `cbor:"2,keyasint"`

	// Optionals
	TTL                   Uint64        `cbor:"3,keyasint,omitempty"`
	Certificates          []Certificate `cbor:"4,keyasint,omitempty"`
	Withdrawals           interface{}   `cbor:"5,keyasint,omitempty"` // unsupported
	Update                interface{}   `cbor:"6,keyasint,omitempty"` // unsupported
	AuxiliaryDataHash     *Hash32       `cbor:"7,keyasint,omitempty"`
	ValidityIntervalStart Uint64        `cbor:"8,keyasint,omitempty"`
	Mint                  *MultiAsset   `cbor:"9,keyasint,omitempty"`
	ScriptDataHash        *Hash32       `cbor:"10,keyasint,omitempty"`
	Collateral            []TxInput     `cbor:"11,keyasint,omitempty"`
	RequiredSigners       []AddrKeyHash `cbor:"12,keyasint,omitempty"`
	NetworkID             Uint64        `cbor:"13,keyasint,omitempty"`
}

// Hash returns the transaction body hash using blake2b256.
func (body *TxBody) Hash() (Hash32, error) {
	bytes, err := cborEnc.Marshal(body)
	if err != nil {
		return Hash32{}, err
	}
	hash := blake2b.Sum256(bytes)
	return hash[:], nil
}

func minUTXO(txOut *TxOutput, protocol *ProtocolParams) Coin {
	var size uint
	if txOut.Amount.OnlyCoin() {
		size = 1
	} else {
		numAssets := txOut.Amount.MultiAsset.NumAssets()
		assetsLength := txOut.Amount.MultiAsset.AssetsLength()
		numPIDs := txOut.Amount.MultiAsset.NumPIDs()

		size = 6 + uint(math.Floor(
			float64(numAssets*12+assetsLength+numPIDs*28+7)/8,
		))
	}
	return Coin(utxoEntrySizeWithoutVal+size) * protocol.CoinsPerUTXOWord
}
