package cmd

import (
	"github.com/echovl/cardano-go/db"
	"github.com/echovl/cardano-go/node"
)

var DefaultCardanoNode = &node.CardanoCli{}
var DefaultDb = db.NewBadgerDB()
