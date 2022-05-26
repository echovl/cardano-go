package wallet

import (
	"github.com/echovl/cardano-go/node"
	"github.com/echovl/cardano-go/types"
)

type Options struct {
	Node node.Node
	DB   DB
}

func (o *Options) init() {
	if o.Node == nil {
		o.Node = node.NewCli(types.Testnet)
	}
	if o.DB == nil {
		o.DB = newBadgerDB()
	}
}
