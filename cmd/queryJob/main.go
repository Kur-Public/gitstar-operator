package main

import (
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"gitstar-operator/pkg/gitOperation"
)

var (
	log = logf.NewDelegatingLogger(zap.Logger())
)

func main() {
	log.Info("start")

	gitOperation.Run("", "")
}
