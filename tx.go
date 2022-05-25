package cardano

import (
	"encoding/hex"
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/ed25519"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/blake2b"
)

type ProtocolParams struct {
	MinimumUtxoValue Coin
	PoolDeposit      uint64
	KeyDeposit       uint64
	MinFeeA          Coin
	MinFeeB          Coin
}

type TransactionID string

func (id TransactionID) Bytes() []byte {
	bytes, err := hex.DecodeString(string(id))
	if err != nil {
		panic(err)
	}

	return bytes
}

type Transaction struct {
	_             struct{} `cbor:",toarray"`
	Body          TransactionBody
	WitnessSet    WitnessSet
	IsValid       bool
	AuxiliaryData interface{} // or null
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

func (tx *Transaction) ID() TransactionID {
	return tx.Body.ID()
}

func DecodeTransaction(cborHex string) (*Transaction, error) {
	bytes, err := hex.DecodeString(cborHex)
	if err != nil {
		return nil, err
	}
	tx := Transaction{}
	if err := cbor.Unmarshal(bytes, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

func CalculateFee(tx *Transaction, protocol ProtocolParams) Coin {
	txBytes := tx.Bytes()
	txLength := uint64(len(txBytes))
	return protocol.MinFeeA*Coin(txLength) + protocol.MinFeeB
}

type WitnessSet struct {
	VKeyWitnessSet []VKeyWitness `cbor:"0,keyasint,omitempty"`
	// TODO: add optional fields 1-4
}

type VKeyWitness struct {
	_         struct{} `cbor:",toarray"`
	VKey      []byte   // ed25519 public key
	Signature []byte   // ed25519 signature
}

type TransactionBody struct {
	Inputs  []TransactionInput  `cbor:"0,keyasint"`
	Outputs []TransactionOutput `cbor:"1,keyasint"`
	Fee     Coin                `cbor:"2,keyasint"`

	// Optionals
	TTL                   Uint64             `cbor:"3,keyasint,omitempty"`
	Certificates          []Certificate      `cbor:"4,keyasint,omitempty"`
	Withdrawals           interface{}        `cbor:"5,keyasint,omitempty"` // Omit for now
	Update                interface{}        `cbor:"6,keyasint,omitempty"` // Omit for now
	AuxiliaryDataHash     *Hash32            `cbor:"7,keyasint,omitempty"`
	ValidityIntervalStart Uint64             `cbor:"8,keyasint,omitempty"`
	Mint                  interface{}        `cbor:"9,keyasint,omitempty"` // Omit for now
	ScriptDataHash        *Hash32            `cbor:"10,keyasint,omitempty"`
	Collateral            []TransactionInput `cbor:"11,keyasint,omitempty"`
	RequiredSigners       []AddrKeyHash      `cbor:"12,keyasint,omitempty"`
	NetworkID             Uint64             `cbor:"13,keyasint,omitempty"`
}

func (body *TransactionBody) Bytes() []byte {
	bytes, err := cbor.Marshal(body)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (body *TransactionBody) ID() TransactionID {
	hash := blake2b.Sum256(body.Bytes())
	return TransactionID(hex.EncodeToString(hash[:]))
}

func (body *TransactionBody) AddSignatures(publicKeys [][]byte, signatures [][]byte) (*Transaction, error) {
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

	return &Transaction{
		Body:          *body,
		WitnessSet:    witnessSet,
		AuxiliaryData: nil,
	}, nil
}

func (body *TransactionBody) calculateMinFee(protocol ProtocolParams) Coin {
	fakeXSigningKey := crypto.NewExtendedSigningKey([]byte{
		0x0c, 0xcb, 0x74, 0xf3, 0x6b, 0x7d, 0xa1, 0x64, 0x9a, 0x81, 0x44, 0x67, 0x55, 0x22, 0xd4, 0xd8, 0x09, 0x7c, 0x64, 0x12,
	}, "")

	witnessSet := WitnessSet{}
	for range body.Inputs {
		witness := VKeyWitness{VKey: fakeXSigningKey.ExtendedVerificationKey()[:32], Signature: fakeXSigningKey.Sign(fakeXSigningKey.ExtendedVerificationKey())}
		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}

	return CalculateFee(&Transaction{
		Body:          *body,
		WitnessSet:    witnessSet,
		AuxiliaryData: nil,
	}, protocol)
}

func (body *TransactionBody) addFee(inputAmount Coin, changeAddress Address, protocol ProtocolParams) error {
	// Set a temporary realistic fee in order to serialize a valid transaction
	body.Fee = 200000

	minFee := body.calculateMinFee(protocol)

	outputAmount := Coin(0)
	for _, txOut := range body.Outputs {
		outputAmount += txOut.Amount
	}
	outputWithFeeAmount := outputAmount + minFee

	if inputAmount < outputWithFeeAmount {
		return fmt.Errorf("insuficient input in transaction, got %v want atleast %v", inputAmount, outputWithFeeAmount)
	}

	if inputAmount == outputWithFeeAmount {
		body.Fee = minFee
		return nil
	}

	change := inputAmount - outputWithFeeAmount
	if change < protocol.MinimumUtxoValue {
		body.Fee = minFee + change // burn change
		return nil
	}

	newBody := TransactionBody{
		Inputs: body.Inputs,
		Outputs: append([]TransactionOutput{{
			Address: changeAddress.Bytes(),
			Amount:  change, // set a temporary value
		}}, body.Outputs...), // change will always be outputs[0] if present
		TTL: body.TTL,
	}
	newMinFee := newBody.calculateMinFee(protocol)
	if change+minFee-newMinFee < protocol.MinimumUtxoValue {
		body.Fee = minFee + change // burn change
		return nil
	}
	body.Outputs = newBody.Outputs
	body.Outputs[0].Amount = change + minFee - newMinFee
	body.Fee = newMinFee
	return nil
}

type TransactionInput struct {
	_     struct{} `cbor:",toarray"`
	ID    []byte   // HashKey 32 bytes
	Index uint64
}

type TransactionOutput struct {
	_       struct{} `cbor:",toarray"`
	Address []byte
	Amount  Coin
}

// One of StakeRegistrationCert, StakeDeregistrationCert, StakeDelegationCert, PoolRegistrationCert
// PoolRetirementCert, GenesisKeyDelegationCert or MoveInstantaneousRewardsCert
//	TODO: MoveInstantaneousRewardsCert requires a map with StakeCredential as a key
type Certificate interface{}

type StakeRegistrationCert struct {
	_               struct{} `cbor:",toarray"`
	ID              uint64   // 0
	StakeCredential StakeCredential
}

type StakeDeregistrationCert struct {
	_               struct{} `cbor:",toarray"`
	ID              uint64   // 1
	StakeCredential StakeCredential
}

type StakeDelegationCert struct {
	_               struct{} `cbor:",toarray"`
	ID              uint64   // 2
	StakeCredential StakeCredential
	PoolKeyHash     PoolKeyHash
}

type PoolRegistrationCert struct {
	_             struct{} `cbor:",toarray"`
	ID            uint64   // 3
	Operator      PoolKeyHash
	VrfKeyHash    Hash32
	Pledge        Coin
	Margin        UnitInterval
	RewardAccount AddressBytes
	Owners        []AddrKeyHash
	Relays        []Relay
	Metadata      *PoolMetadata // or null
}

type PoolRetirementCert struct {
	_           struct{} `cbor:",toarray"`
	ID          uint64   // 4
	PoolKeyHash PoolKeyHash
	Epoch       uint64
}

type GenesisKeyDelegationCert struct {
	_                   struct{} `cbor:",toarray"`
	ID                  uint64   // 5
	GenesisHash         Hash28
	GenesisDelegateHash Hash28
	VrfKeyHash          Hash32
}

type StakeCredential struct {
	_ struct{} `cbor:",toarray"`
	// 0 for AddrKeyHash, 1 for ScriptHash
	ID          uint64
	AddrKeyHash AddrKeyHash `cbor:",omitempty"`
	ScriptHash  Hash28      `cbor:",omitempty"`
}

type PoolMetadata struct {
	_    struct{} `cbor:",toarray"`
	URL  string
	Hash Hash32
}

// One of SingleHostAddr, SingleHostName or MultiHostName
type Relay interface{}

type SingleHostAddr struct {
	_    struct{} `cbor:",toarray"`
	ID   uint64   // 0
	Port *uint64  // or null
	Ipv4 []byte   // or null
	Ipv6 []byte   // or null
}

type SingleHostName struct {
	_       struct{} `cbor:",toarray"`
	ID      uint64   // 1
	Port    *uint64  // or null
	DNSName string   // A or AAA DNS record
}

type MultiHostName struct {
	_       struct{} `cbor:",toarray"`
	ID      uint64   // 2
	DNSName string   // SRV DNS record
}
