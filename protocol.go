package cardano

type ProtocolParams struct {
	MinFeeA              Coin
	MinFeeB              Coin
	MaxBlockBodySize     uint
	MaxTxSize            uint
	MaxBlockHeaderSize   uint
	KeyDeposit           Coin
	PoolDeposit          Coin
	MaxEpoch             uint
	NOpt                 uint
	PoolPledgeInfluence  Rational
	ExpansionRate        UnitInterval
	TreasuryGrowthRate   UnitInterval
	D                    UnitInterval
	ExtraEntropy         []byte
	ProtocolVersion      ProtocolVersion
	MinPoolCost          Coin
	CoinsPerUTXOWord     Coin
	CostModels           interface{}
	ExecutionCosts       interface{}
	MaxTxExUnits         interface{}
	MaxBlockTxExUnits    interface{}
	MaxValueSize         uint
	CollateralPercentage uint
	MaxCollateralInputs  uint
}

type ProtocolVersion struct {
	_     struct{} `cbor:"_,toarray"`
	Major uint
	Minor uint
}
