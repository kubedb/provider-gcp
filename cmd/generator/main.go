/*
Copyright 2021 Upbound Inc.
*/

package main

import (
	"fmt"
	dynamic_controller "kubedb.dev/provider-gcp/cmd/dynamic-controller"
	"os"
	"path/filepath"

	"github.com/crossplane/upjet/pkg/pipeline"

	pconfig "kubedb.dev/provider-gcp/config"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "" {
		panic("root directory is required to be given as argument")
	}
	rootDir := os.Args[1]
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		panic(fmt.Sprintf("cannot calculate the absolute path with %s", rootDir))
	}
	pc := pconfig.GetProvider()
	pipeline.Run(pc, absRootDir)
	dynamic_controller.GenerateController(pc, absRootDir)
}
