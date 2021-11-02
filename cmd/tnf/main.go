package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/test-network-function/test-network-function/cmd/tnf/addclaim"
	"github.com/test-network-function/test-network-function/cmd/tnf/generate/catalog"
	"github.com/test-network-function/test-network-function/cmd/tnf/generate/handler"
	"github.com/test-network-function/test-network-function/cmd/tnf/grade"
	"github.com/test-network-function/test-network-function/cmd/tnf/jsontest"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tnf",
		Short: "A CLI for creating, validating , and test-network-function tests.",
	}

	generate = &cobra.Command{
		Use:   "generate",
		Short: "generator tool for various tnf artifacts.",
	}
)

func main() {
	claimAddFile := addclaim.NewCommand()
	rootCmd.AddCommand(claimAddFile)
	rootCmd.AddCommand(generate)
	generateCatalog := catalog.NewCommand()
	generate.AddCommand(generateCatalog)
	handlercmd := handler.NewCommand()
	generate.AddCommand(handlercmd)
	jsontestCli := jsontest.NewCommand()
	rootCmd.AddCommand(jsontestCli)
	gradetool := grade.NewCommand()
	rootCmd.AddCommand(gradetool)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
