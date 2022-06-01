package cardano

import (
	"encoding/hex"
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/ed25519"
	"golang.org/x/crypto/blake2b"
)

var utxoEntrySizeWithoutVal = 27

type UTxO struct {
	TxHash  Hash32
	Spender Address
	Amount  Coin
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
	VKeyWitnessSet []VKeyWitness `cbor:"0,keyasint,omitempty"`
	// TODO: add optional fields 1-4
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
	Amount Coin `cbor:"-"`
}

// NewTxInput creates a new instance of TxInput
func NewTxInput(txHash string, index uint, amount Coin) (*TxInput, error) {
	txHash32, err := NewHash32(txHash)
	if err != nil {
		return nil, err
	}

	return &TxInput{TxHash: txHash32, Index: uint64(index), Amount: amount}, nil
}

type TxOutput struct {
	_       struct{} `cbor:",toarray"`
	Address Address
	Amount  Coin
}

// NewTxOutput creates a new instance of TxOutput
func NewTxOutput(addr string, amount Coin) (*TxOutput, error) {
	addrOut, err := NewAddress(addr)
	if err != nil {
		return nil, err
	}

	return &TxOutput{Address: addrOut, Amount: amount}, nil
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
	Mint                  interface{}   `cbor:"9,keyasint,omitempty"` // unsupported
	ScriptDataHash        *Hash32       `cbor:"10,keyasint,omitempty"`
	Collateral            []TxInput     `cbor:"11,keyasint,omitempty"`
	RequiredSigners       []AddrKeyHash `cbor:"12,keyasint,omitempty"`
	NetworkID             Uint64        `cbor:"13,keyasint,omitempty"`
}

// Hash returns the transaction body hash using blake2b.
func (body *TxBody) Hash() (Hash32, error) {
	bytes, err := cborEnc.Marshal(body)
	if err != nil {
		return Hash32{}, err
	}
	hash := blake2b.Sum256(bytes)
	return hash[:], nil
}

// AddSignatures sets the transaction's witness set.
func (body *TxBody) AddSignatures(publicKeys [][]byte, signatures [][]byte) (*Tx, error) {
	if len(publicKeys) != len(signatures) {
		return nil, fmt.Errorf("missmatch length of publicKeys and signatures")
	}
	if len(signatures) != len(body.Inputs) {
		return nil, fmt.Errorf("missmatch length of signatures and inputs")
	}

	witnessSet := WitnessSet{}

	for i := 0; i < len(publicKeys); i++ {
		if len(signatures[i]) != ed25519.SignatureSize {
			return nil, fmt.Errorf("invalid signature length %v", len(signatures[i]))
		}
		witness := VKeyWitness{VKey: publicKeys[i], Signature: signatures[i]}
		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}

	return &Tx{
		Body:          *body,
		WitnessSet:    witnessSet,
		AuxiliaryData: nil,
		IsValid:       true,
	}, nil
}

func minUTXO(txOut *TxOutput, protocol *ProtocolParams) Coin {
	return Coin(utxoEntrySizeWithoutVal+1) * protocol.CoinsPerUTXOWord
}
