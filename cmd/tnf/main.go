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
	rootCmd.AddCommand(addclaim.NewCommand())
	rootCmd.AddCommand(generate)
	generate.AddCommand(catalog.NewCommand())
	generate.AddCommand(handler.NewCommand())
	rootCmd.AddCommand(jsontest.NewCommand())
	rootCmd.AddCommand(grade.NewCommand())
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
