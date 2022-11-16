package koios

import (
	"context"
	"math/big"

	"github.com/cardano-community/koios-go-client/v2"
	"github.com/echovl/cardano-go"
	"github.com/echovl/cardano-go/internal/cbor"
)

const koiosTestnetHost = "testnet.koios.rest"

type KoiosCli struct {
	*koios.Client
	ctx     context.Context
	network cardano.Network
}

var _ cardano.Node = &KoiosCli{}

func (kc *KoiosCli) UTxOs(a cardano.Address) ([]cardano.UTxO, error) {
	utxos := []cardano.UTxO{}
	opts := kc.NewRequestOptions()
	opts.QuerySet("select", "utxo_set")
	res, err := kc.GetAddressInfo(kc.ctx, koios.Address(a.Bech32()), opts)
	if err != nil || res.Data == nil {
		return nil, err
	}

	for _, utxo := range res.Data.UTxOs {
		utxo_ := cardano.UTxO{
			TxHash:  cardano.Hash32(utxo.TxHash),
			Index:   uint64(utxo.TxIndex),
			Spender: a,
			Amount: cardano.NewValue(
				cardano.Coin(utxo.Value.IntPart())),
		}
		if len(utxo.AssetList) > 0 {
			ma := utxo_.Amount.MultiAsset
			for _, as := range utxo.AssetList {
				asset := cardano.NewAssets()
				policyId := cardano.NewPolicyIDFromHash(
					cardano.Hash28(as.PolicyID.String()))
				asset.Set(cardano.NewAssetName(string(as.AssetName)),
					cardano.BigNum(as.Quantity.IntPart()))
				ma.Set(policyId, asset)
			}
		}
		utxos = append(utxos, utxo_)
	}
	return utxos, nil
}

// Tip returns the node's current tip
func (kc *KoiosCli) Tip() (*cardano.NodeTip, error) {
	opts := kc.NewRequestOptions()
	opts.QuerySet("select", "block_no,epoch_no,abs_slot")
	res, err := kc.GetTip(kc.ctx, opts)
	if err != nil {
		return nil, err
	}
	return &cardano.NodeTip{
		Epoch: uint64(res.Data.EpochNo),
		Block: uint64(res.Data.BlockNo),
		Slot:  uint64(res.Data.AbsSlot),
	}, nil
}

// SubmitTx submits a transaction to the node using cbor encoding
func (kc *KoiosCli) SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error) {
	cborHex, err := cbor.Marshal(tx)
	if err != nil {
		return nil, err
	}
	stx := koios.TxBodyJSON{
		// TODO: the Tx should embed the type/era
		Type:        "TxBabbage",
		Description: "",
		CborHex:     string(cborHex),
	}
	res, err := kc.SubmitSignedTx(kc.ctx, stx, nil)
	if err != nil {
		return nil, err
	}
	h, err := cardano.NewHash32(string(res.Data))
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// ProtocolParams returns the Node's Protocol Parameters
func (kc *KoiosCli) ProtocolParams() (*cardano.ProtocolParams, error) {
	opts := kc.NewRequestOptions()
	opts.QuerySet("limit", "1")
	res, err := kc.GetEpochParams(kc.ctx, nil, opts)
	if err != nil {
		return nil, err
	}
	ep := res.Data[0]

	var d, infl, tresuryExpRate, expRate big.Rat
	big.NewFloat(ep.Decentralisation).Rat(&d)
	big.NewFloat(ep.Influence).Rat(&infl)
	big.NewFloat(ep.TreasuryGrowthRate).Rat(&tresuryExpRate)
	big.NewFloat(ep.MonetaryExpandRate).Rat(&expRate)

	return &cardano.ProtocolParams{
		MinFeeA:            cardano.Coin(ep.MinFeeA.IntPart()),
		MinFeeB:            cardano.Coin(ep.MinFeeB.IntPart()),
		MaxBlockBodySize:   uint(ep.MaxBlockSize),
		MaxTxSize:          uint(ep.MaxTxSize),
		MaxBlockHeaderSize: uint(ep.MaxBhSize),
		KeyDeposit:         cardano.Coin(ep.KeyDeposit.IntPart()),
		PoolDeposit:        cardano.Coin(ep.PoolDeposit.IntPart()),
		MaxEpoch:           uint(ep.MaxEpoch),
		//NOpt                 uint
		PoolPledgeInfluence: cardano.Rational{
			P: infl.Num().Uint64(),
			Q: infl.Denom().Uint64(),
		},
		ExpansionRate: cardano.UnitInterval{
			P: expRate.Num().Uint64(),
			Q: expRate.Denom().Uint64(),
		},
		TreasuryGrowthRate: cardano.UnitInterval{
			P: tresuryExpRate.Num().Uint64(),
			Q: tresuryExpRate.Denom().Uint64(),
		},
		D: cardano.UnitInterval{
			P: d.Num().Uint64(),
			Q: d.Denom().Uint64(),
		},
		ExtraEntropy: []byte(ep.ExtraEntropy),
		ProtocolVersion: cardano.ProtocolVersion{
			Major: uint(ep.ProtocolMajor),
			Minor: uint(ep.ProtocolMinor),
		},

		MinPoolCost:      cardano.Coin(ep.MinPoolCost.IntPart()),
		CoinsPerUTXOWord: cardano.Coin(ep.CoinsPerUtxoSize.IntPart()),
		CostModels:       ep.CostModels,
		// ExecutionCosts:      interface{}
		// MaxTxExUnits         interface{}
		// MaxBlockTxExUnits    interface{}
		MaxValueSize:         uint(ep.MaxValSize),
		CollateralPercentage: uint(ep.CollateralPercent),
		MaxCollateralInputs:  uint(ep.MaxCollateralInputs),
	}, nil
}

// Network returns the node's current network type
func (kc *KoiosCli) Network() cardano.Network {
	return kc.network
}

func NewNode(network cardano.Network, ctx context.Context, opts ...koios.Option) cardano.Node {
	if ctx == nil {
		ctx = context.Background()
	}
	if network == cardano.Testnet {
		opts = append(opts, koios.Host(koiosTestnetHost))
	}

	c, err := koios.New(opts...)
	if err != nil {
		panic(err)
	}
	return &KoiosCli{Client: c, ctx: ctx, network: network}
}
