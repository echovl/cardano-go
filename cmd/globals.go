package cmd

import (
	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/provider"
)

var DefaultProvider = &provider.NodeCli{}
var DefaultDb = db.NewBadgerDB()
