package main

import (
	"context"
	"flag"
	"github.com/SyntropyNet/terraform-provider-syntropystack/syntropy"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"log"
	"math/rand"
	"time"
)

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "v0.0.1"

	// goreleaser can also pass the specific commit if you want
	//commit string = ""
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/SyntropyNet/syntropystack",
		Debug:   debugMode,
	}

	err := providerserver.Serve(context.Background(), syntropy.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
