package encoding

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

var CBOR, _ = cbor.CanonicalEncOptions().EncMode()

func GetTypeFromCBORArray(data []byte) (uint64, error) {
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
