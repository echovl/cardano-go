package cmd

import (
	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/node"
)

var DefaultCardanoNode = &node.CardanoCli{}
var DefaultDb = db.NewBadgerDB()
