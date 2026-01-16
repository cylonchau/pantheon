package main

import (
	"flag"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/cylonchau/pantheon/pkg/server"

	_ "github.com/cylonchau/pantheon/docs"
)

// @title Pantheon server
// @version v0.0.0-dev
// @description Prometheus hub, distrubed prometheus targent manager.

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @BasePath /
// @schemes http
func main() {
	command := server.NewProxyCommand()
	flagset := flag.CommandLine
	klog.InitFlags(flagset)
	pflag.CommandLine.AddGoFlagSet(flagset)
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
