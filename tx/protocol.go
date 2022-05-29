package tx

import "github.com/echovl/cardano-go/types"

type ProtocolParams struct {
	CoinsPerUTXOWord types.Coin
	PoolDeposit      types.Coin
	KeyDeposit       types.Coin
	MinFeeA          types.Coin
	MinFeeB          types.Coin
}
