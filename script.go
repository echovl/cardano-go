package cardano

import "github.com/echovl/cardano-go/crypto"

type ScriptHashNamespace uint8

const (
	NativeScriptNamespace ScriptHashNamespace = iota
	PlutusScriptNamespace
)

type NativeScriptType uint64

const (
	ScriptPubKey NativeScriptType = iota
	ScriptAll
	ScriptAny
	ScriptNofK
	ScriptInvalidBefore
	ScriptInvalidAfter
)

type NativeScript struct {
	Type          NativeScriptType
	KeyHash       AddrKeyHash
	N             uint64
	Scripts       []NativeScript
	IntervalValue uint64
}

func NewScriptPubKey(publicKey crypto.PubKey) (NativeScript, error) {
	keyHash, err := publicKey.Hash()
	if err != nil {
		return NativeScript{}, err
	}
	return NativeScript{Type: ScriptPubKey, KeyHash: keyHash}, nil
}

// Hash returns the script hash using blake2b224.
func (ns *NativeScript) Hash() (Hash28, error) {
	bytes, err := ns.Bytes()
	if err != nil {
		return nil, err
	}
	bytes = append([]byte{byte(NativeScriptNamespace)}, bytes...)
	return Blake224Hash(append(bytes))
}

// Bytes returns the CBOR encoding of the script as bytes.
func (ns *NativeScript) Bytes() ([]byte, error) {
	return cborEnc.Marshal(ns)
}

// MarshalCBOR implements cbor.Marshaler.
func (ns *NativeScript) MarshalCBOR() ([]byte, error) {
	var script []interface{}
	switch ns.Type {
	case ScriptPubKey:
		script = append(script, ns.Type, ns.KeyHash)
	case ScriptAll, ScriptAny:
		script = append(script, ns.Type, ns.Scripts)
	case ScriptNofK:
		script = append(script, ns.Type, ns.N, ns.Scripts)
	case ScriptInvalidBefore, ScriptInvalidAfter:
		script = append(script, ns.Type, ns.IntervalValue)
	}
	return cborEnc.Marshal(script)
}
