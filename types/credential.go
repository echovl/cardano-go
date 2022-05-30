package types

import (
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/internal/encoding"
	"github.com/fxamacker/cbor/v2"
)

type StakeCredentialType uint64

const (
	AddrKeyCredential StakeCredentialType = 0
	ScriptCredential                      = 1
)

type addrKeyStakeCredential struct {
	_           struct{} `cbor:",toarray"`
	Type        StakeCredentialType
	AddrKeyHash AddrKeyHash
}

type scriptStakeCredential struct {
	_          struct{} `cbor:",toarray"`
	Type       StakeCredentialType
	ScriptHash Hash28
}

type StakeCredential struct {
	Type        StakeCredentialType
	AddrKeyHash AddrKeyHash
	ScriptHash  Hash28
}

func (s *StakeCredential) Hash() Hash28 {
	if s.Type == AddrKeyCredential {
		return s.AddrKeyHash
	} else {
		return s.ScriptHash
	}
}

func NewAddrKeyCredential(publicKey crypto.PubKey) (StakeCredential, error) {
	keyHash, err := blake224Hash(publicKey)
	if err != nil {
		return StakeCredential{}, err
	}
	return StakeCredential{Type: AddrKeyCredential, AddrKeyHash: keyHash}, nil
}

func NewScriptCredential(script []byte) (StakeCredential, error) {
	scriptHash, err := blake224Hash(script)
	if err != nil {
		return StakeCredential{}, err
	}
	return StakeCredential{Type: ScriptCredential, ScriptHash: scriptHash}, nil
}

// MarshalCBOR implements cbor.Marshaler.
func (s *StakeCredential) MarshalCBOR() ([]byte, error) {
	var cred []interface{}
	switch s.Type {
	case AddrKeyCredential:
		cred = append(cred, s.Type, s.AddrKeyHash)
	case ScriptCredential:
		cred = append(cred, s.Type, s.ScriptHash)
	}

	return encoding.CBOR.Marshal(cred)

}

// UnmarshalCBOR implements cbor.Unmarshaler.
func (s *StakeCredential) UnmarshalCBOR(data []byte) error {
	credType, err := encoding.GetTypeFromCBORArray(data)
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
