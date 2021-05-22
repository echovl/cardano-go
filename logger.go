package cardano

import (
	"os"

	"go.uber.org/zap"
)

var log *zap.SugaredLogger
var homePath = os.Getenv("HOME")

const fallbackLogFilePath = "./.cardano_wallet_log"

func init() {
	config := zap.NewDevelopmentConfig()
	if homePath != "" {
		config.OutputPaths = []string{homePath + "/.cardano_wallet_log"}
	} else {
		config.OutputPaths = []string{fallbackLogFilePath}
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	log = logger.Sugar()
}
