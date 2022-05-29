package tx

import "github.com/echovl/cardano-go/types"

type ProtocolParams struct {
	MinUTXO     types.Coin
	PoolDeposit types.Coin
	KeyDeposit  types.Coin
	MinFeeA     types.Coin
	MinFeeB     types.Coin
}
