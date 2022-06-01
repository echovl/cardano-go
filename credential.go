package cardano

import (
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"github.com/fxamacker/cbor/v2"
)

type StakeCredentialType uint64

const (
	KeyCredential    StakeCredentialType = 0
	ScriptCredential                     = 1
)

type keyStakeCredential struct {
	_       struct{} `cbor:",toarray"`
	Type    StakeCredentialType
	KeyHash AddrKeyHash
}

type scriptStakeCredential struct {
	_          struct{} `cbor:",toarray"`
	Type       StakeCredentialType
	ScriptHash Hash28
}

type StakeCredential struct {
	Type       StakeCredentialType
	KeyHash    AddrKeyHash
	ScriptHash Hash28
}

func (s *StakeCredential) Hash() Hash28 {
	if s.Type == KeyCredential {
		return s.KeyHash
	} else {
		return s.ScriptHash
	}
}

func NewKeyCredential(publicKey crypto.PubKey) (StakeCredential, error) {
	keyHash, err := blake224Hash(publicKey)
	if err != nil {
		return StakeCredential{}, err
	}
	return StakeCredential{Type: KeyCredential, KeyHash: keyHash}, nil
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
	case KeyCredential:
		cred = append(cred, s.Type, s.KeyHash)
	case ScriptCredential:
		cred = append(cred, s.Type, s.ScriptHash)
	}

	return cborEnc.Marshal(cred)

}

// UnmarshalCBOR implements cbor.Unmarshaler.
func (s *StakeCredential) UnmarshalCBOR(data []byte) error {
	credType, err := getTypeFromCBORArray(data)
	if err != nil {
		return fmt.Errorf("cbor: cannot unmarshal CBOR array into StakeCredential (%v)", err)
	}

	switch StakeCredentialType(credType) {
	case KeyCredential:
		cred := &keyStakeCredential{}
		if err := cbor.Unmarshal(data, cred); err != nil {
			return err
		}
		s.Type = KeyCredential
		s.KeyHash = cred.KeyHash
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
