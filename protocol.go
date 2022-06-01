package cardano

type ProtocolParams struct {
	CoinsPerUTXOWord Coin
	PoolDeposit      Coin
	KeyDeposit       Coin
	MinFeeA          Coin
	MinFeeB          Coin
}
