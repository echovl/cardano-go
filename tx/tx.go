package tx

import (
	"encoding/hex"
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/types"
	"github.com/echovl/ed25519"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/blake2b"
)

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

func (tx *Transaction) Hash() types.Hash32 {
	return tx.Body.Hash()
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

func CalculateFee(tx *Transaction, protocol types.ProtocolParams) types.Coin {
	txBytes := tx.Bytes()
	txLength := uint64(len(txBytes))
	return protocol.MinFeeA*types.Coin(txLength) + protocol.MinFeeB
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

type TransactionInput struct {
	_      struct{} `cbor:",toarray"`
	TxHash types.Hash32
	Index  uint64
}

type TransactionOutput struct {
	_       struct{} `cbor:",toarray"`
	Address []byte
	Amount  types.Coin
}

type TransactionBody struct {
	Inputs  []TransactionInput  `cbor:"0,keyasint"`
	Outputs []TransactionOutput `cbor:"1,keyasint"`
	Fee     types.Coin          `cbor:"2,keyasint"`

	// Optionals
	TTL                   types.Uint64        `cbor:"3,keyasint,omitempty"`
	Certificates          []Certificate       `cbor:"4,keyasint,omitempty"`
	Withdrawals           interface{}         `cbor:"5,keyasint,omitempty"` // Omit for now
	Update                interface{}         `cbor:"6,keyasint,omitempty"` // Omit for now
	AuxiliaryDataHash     types.Hash32        `cbor:"7,keyasint,omitempty"`
	ValidityIntervalStart types.Uint64        `cbor:"8,keyasint,omitempty"`
	Mint                  interface{}         `cbor:"9,keyasint,omitempty"` // Omit for now
	ScriptDataHash        types.Hash32        `cbor:"10,keyasint,omitempty"`
	Collateral            []TransactionInput  `cbor:"11,keyasint,omitempty"`
	RequiredSigners       []types.AddrKeyHash `cbor:"12,keyasint,omitempty"`
	NetworkID             types.Uint64        `cbor:"13,keyasint,omitempty"`
}

func (body *TransactionBody) Bytes() []byte {
	bytes, err := cbor.Marshal(body)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (body *TransactionBody) Hash() types.Hash32 {
	return blake2b.Sum256(body.Bytes())
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

func (body *TransactionBody) calculateMinFee(protocol types.ProtocolParams) types.Coin {
	fakeXSigningKey := crypto.NewExtendedSigningKey([]byte{
		0x0c, 0xcb, 0x74, 0xf3, 0x6b, 0x7d, 0xa1, 0x64, 0x9a, 0x81,
		0x44, 0x67, 0x55, 0x22, 0xd4, 0xd8, 0x09, 0x7c, 0x64, 0x12,
	}, "")

	witnessSet := WitnessSet{}
	for range body.Inputs {
		witness := VKeyWitness{
			VKey:      fakeXSigningKey.ExtendedVerificationKey()[:32],
			Signature: fakeXSigningKey.Sign(fakeXSigningKey.ExtendedVerificationKey()),
		}
		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}

	return CalculateFee(&Transaction{
		Body:          *body,
		WitnessSet:    witnessSet,
		AuxiliaryData: nil,
	}, protocol)
}

func (body *TransactionBody) addFee(inputAmount types.Coin, changeAddress types.Address, protocol types.ProtocolParams) error {
	// Set a temporary realistic fee in order to serialize a valid transaction
	body.Fee = 200000

	minFee := body.calculateMinFee(protocol)

	outputAmount := types.Coin(0)
	for _, txOut := range body.Outputs {
		outputAmount += txOut.Amount
	}
	outputWithFeeAmount := outputAmount + minFee

	if inputAmount < outputWithFeeAmount {
		return fmt.Errorf(
			"insuficient input in transaction, got %v want atleast %v",
			inputAmount,
			outputWithFeeAmount,
		)
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

type CertificateType uint

const (
	StakeRegistration        CertificateType = 0
	StakeDeregistration                      = 1
	StakeDelegation                          = 2
	PoolRegistration                         = 3
	PoolRetirement                           = 4
	GenesisKeyDelegation                     = 5
	MoveInstantaneousRewards                 = 6
)

type stakeRegistration struct {
	_               struct{} `cbor:",toarray"`
	Type            CertificateType
	StakeCredential StakeCredential
}

type stakeDeregistration struct {
	_               struct{} `cbor:",toarray"`
	Type            CertificateType
	StakeCredential StakeCredential
}

type stakeDelegation struct {
	_               struct{} `cbor:",toarray"`
	Type            CertificateType
	StakeCredential StakeCredential
	PoolKeyHash     types.PoolKeyHash
}

type poolRegistration struct {
	_             struct{} `cbor:",toarray"`
	Type          CertificateType
	Operator      types.PoolKeyHash
	VrfKeyHash    types.Hash32
	Pledge        types.Coin
	Margin        types.UnitInterval
	RewardAccount types.AddressBytes
	Owners        []types.AddrKeyHash
	Relays        []Relay
	PoolMetadata  *PoolMetadata // or null
}

type poolRetirement struct {
	_           struct{} `cbor:",toarray"`
	Type        CertificateType
	PoolKeyHash types.PoolKeyHash
	Epoch       uint64
}

type genesisKeyDelegation struct {
	_                   struct{} `cbor:",toarray"`
	Type                CertificateType
	GenesisHash         types.Hash28
	GenesisDelegateHash types.Hash28
	VrfKeyHash          types.Hash32
}

//	TODO: MoveInstantaneousRewardsCert requires a map with StakeCredential as a key
type Certificate struct {
	Type CertificateType

	// Common fields
	StakeCredential StakeCredential
	PoolKeyHash     types.PoolKeyHash
	VrfKeyHash      types.Hash32

	// Pool related fields
	Operator      types.PoolKeyHash
	Pledge        types.Coin
	Margin        types.UnitInterval
	RewardAccount types.AddressBytes
	Owners        []types.AddrKeyHash
	Relays        []Relay
	PoolMetadata  *PoolMetadata // or null
	Epoch         uint64

	// Genesis fields
	GenesisHash         types.Hash28
	GenesisDelegateHash types.Hash28
}

func (c *Certificate) MarshalCBOR() ([]byte, error) {
	var cert interface{}
	switch c.Type {
	case StakeRegistration:
		cert = stakeRegistration{
			Type:            c.Type,
			StakeCredential: c.StakeCredential,
		}
	case StakeDeregistration:
		cert = stakeDeregistration{
			Type:            c.Type,
			StakeCredential: c.StakeCredential,
		}
	case StakeDelegation:
		cert = stakeDelegation{
			Type:            c.Type,
			StakeCredential: c.StakeCredential,
			PoolKeyHash:     c.PoolKeyHash,
		}
	case PoolRegistration:
		cert = poolRegistration{
			Type:          c.Type,
			Operator:      c.Operator,
			VrfKeyHash:    c.VrfKeyHash,
			Pledge:        c.Pledge,
			Margin:        c.Margin,
			RewardAccount: c.RewardAccount,
			Owners:        c.Owners,
			Relays:        c.Relays,
			PoolMetadata:  c.PoolMetadata,
		}
	case PoolRetirement:
		cert = poolRetirement{
			Type:        c.Type,
			PoolKeyHash: c.PoolKeyHash,
			Epoch:       c.Epoch,
		}
	case GenesisKeyDelegation:
		cert = genesisKeyDelegation{
			Type:                c.Type,
			GenesisHash:         c.GenesisHash,
			GenesisDelegateHash: c.GenesisDelegateHash,
			VrfKeyHash:          c.VrfKeyHash,
		}
	}

	return cbor.Marshal(cert)
}

func (c *Certificate) UnmarshalCBOR(data []byte) error {
	certType, err := getTypeFromCBORArray(data)
	if err != nil {
		return fmt.Errorf("cbor: cannot unmarshal CBOR array into StakeCredential (%v)", err)
	}

	switch CertificateType(certType) {
	case StakeRegistration:
		cert := &stakeRegistration{}
		if err := cbor.Unmarshal(data, cert); err != nil {
			return err
		}
		c.Type = StakeRegistration
		c.StakeCredential = cert.StakeCredential
	case StakeDeregistration:
		cert := &stakeDeregistration{}
		if err := cbor.Unmarshal(data, cert); err != nil {
			return err
		}
		c.Type = StakeDeregistration
		c.StakeCredential = cert.StakeCredential
	case StakeDelegation:
		cert := &stakeDelegation{}
		if err := cbor.Unmarshal(data, cert); err != nil {
			return err
		}
		c.Type = StakeDelegation
		c.StakeCredential = cert.StakeCredential
		c.PoolKeyHash = cert.PoolKeyHash
	case PoolRegistration:
		cert := &poolRegistration{}
		if err := cbor.Unmarshal(data, cert); err != nil {
			return err
		}
		c.Type = PoolRegistration
		c.Operator = cert.Operator
		c.VrfKeyHash = cert.VrfKeyHash
		c.Pledge = cert.Pledge
		c.Margin = cert.Margin
		c.RewardAccount = cert.RewardAccount
		c.Owners = cert.Owners
		c.Relays = cert.Relays
		c.PoolMetadata = cert.PoolMetadata
	case PoolRetirement:
		cert := &poolRetirement{}
		if err := cbor.Unmarshal(data, cert); err != nil {
			return err
		}
		c.Type = PoolRetirement
		c.PoolKeyHash = cert.PoolKeyHash
		c.Epoch = cert.Epoch
	case GenesisKeyDelegation:
		cert := &genesisKeyDelegation{}
		if err := cbor.Unmarshal(data, cert); err != nil {
			return err
		}
		c.Type = GenesisKeyDelegation
		c.GenesisHash = cert.GenesisHash
		c.GenesisDelegateHash = cert.GenesisDelegateHash
		c.VrfKeyHash = cert.VrfKeyHash
	}

	return nil
}

type StakeCredentialType uint64

const (
	AddrKeyCredential StakeCredentialType = 0
	ScriptCredential                      = 1
)

type addrKeyStakeCredential struct {
	_           struct{} `cbor:",toarray"`
	Type        StakeCredentialType
	AddrKeyHash types.AddrKeyHash
}

type scriptStakeCredential struct {
	_          struct{} `cbor:",toarray"`
	Type       StakeCredentialType
	ScriptHash types.Hash28
}

type StakeCredential struct {
	Type        StakeCredentialType
	AddrKeyHash types.AddrKeyHash
	ScriptHash  types.Hash28
}

func (s *StakeCredential) MarshalCBOR() ([]byte, error) {
	var cred []interface{}
	switch s.Type {
	case AddrKeyCredential:
		cred = append(cred, s.Type, s.AddrKeyHash)
	case ScriptCredential:
		cred = append(cred, s.Type, s.ScriptHash)
	}

	return cbor.Marshal(cred)

}

func (s *StakeCredential) UnmarshalCBOR(data []byte) error {
	credType, err := getTypeFromCBORArray(data)
	if err != nil {
		return fmt.Errorf("cbor: cannot unmarshal CBOR array into StakeCredential (%v)", err)
	}

	switch StakeCredentialType(credType) {
	case AddrKeyCredential:
		cred := &addrKeyStakeCredential{}
		if err := cbor.Unmarshal(data, cred); err != nil {
			return err
		}
		s.Type = AddrKeyCredential
		s.AddrKeyHash = cred.AddrKeyHash
	case ScriptCredential:
		cred := &scriptStakeCredential{}
		if err := cbor.Unmarshal(data, cred); err != nil {
			return err
		}
		s.Type = ScriptCredential
		s.ScriptHash = cred.ScriptHash
	}

	return nil
}

type PoolMetadata struct {
	_    struct{} `cbor:",toarray"`
	URL  string
	Hash types.Hash32
}

type RelayType uint64

const (
	SingleHostAddr RelayType = 0
	SingleHostName           = 1
	MultiHostName            = 2
)

type singleHostAddr struct {
	_    struct{} `cbor:",toarray"`
	Type RelayType
	Port types.Uint64
	Ipv4 []byte
	Ipv6 []byte
}

type singleHostName struct {
	_       struct{} `cbor:",toarray"`
	Type    RelayType
	Port    types.Uint64
	DNSName string
}

type multiHostName struct {
	_       struct{} `cbor:",toarray"`
	Type    RelayType
	DNSName string
}

type Relay struct {
	Type    RelayType
	Port    types.Uint64
	Ipv4    []byte
	Ipv6    []byte
	DNSName string
}

func (r *Relay) MarshalCBOR() ([]byte, error) {
	var relay interface{}
	switch r.Type {
	case SingleHostAddr:
		relay = singleHostAddr{
			Type: r.Type,
			Port: r.Port,
			Ipv4: r.Ipv4,
			Ipv6: r.Ipv6,
		}
	case SingleHostName:
		relay = singleHostName{
			Type:    r.Type,
			Port:    r.Port,
			DNSName: r.DNSName,
		}
	case MultiHostName:
		relay = multiHostName{
			Type:    r.Type,
			DNSName: r.DNSName,
		}
	}

	return cbor.Marshal(relay)
}

func (r *Relay) UnmarshalCBOR(data []byte) error {
	relayType, err := getTypeFromCBORArray(data)
	if err != nil {
		return fmt.Errorf("cbor: cannot unmarshal CBOR array into Relay (%v)", err)
	}

	switch RelayType(relayType) {
	case SingleHostAddr:
		rl := &singleHostAddr{}
		if err := cbor.Unmarshal(data, rl); err != nil {
			return err
		}
		r.Type = SingleHostAddr
		r.Port = rl.Port
		r.Ipv4 = rl.Ipv4
		r.Ipv6 = rl.Ipv6
	case SingleHostName:
		rl := &singleHostName{}
		if err := cbor.Unmarshal(data, rl); err != nil {
			return err
		}
		r.Type = SingleHostName
		r.Port = rl.Port
		r.DNSName = rl.DNSName
	case MultiHostName:
		rl := &multiHostName{}
		if err := cbor.Unmarshal(data, rl); err != nil {
			return err
		}
		r.Type = MultiHostName
		r.DNSName = rl.DNSName
	}

	return nil
}

func getTypeFromCBORArray(data []byte) (uint64, error) {
	raw := []interface{}{}
	if err := cbor.Unmarshal(data, &raw); err != nil {
		return 0, err
	}

	if len(raw) == 0 {
		return 0, fmt.Errorf("empty CBOR array")
	}

	t, ok := raw[0].(uint64)
	if !ok {
		return 0, fmt.Errorf("invalid Type")
	}

	return t, nil
}
