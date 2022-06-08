package cardano

// ProtocolParams is a Cardano Protocol Parameters.
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

// ProtocolVersion is the protocol version number.
type ProtocolVersion struct {
	_     struct{} `cbor:"_,toarray"`
	Major uint
	Minor uint
}
