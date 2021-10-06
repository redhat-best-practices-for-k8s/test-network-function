package main

import (
	"bufio"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

type myHandler struct {
	UpperHandlername string
	LowerHandlername string
}

var (
	pathrelativetoroot string
	handlerDirectory   string
	handlername        string

	rootCmd = &cobra.Command{
		Use:   "tnf",
		Short: "A CLI for creating, validating, and test-network-function tests.",
	}

	generate = &cobra.Command{
		Use:   "generate",
		Short: "generator tool for various tnf artifacts.",
	}

	handler = &cobra.Command{
		Use:   "handler",
		Short: "adding new handler.",
		RunE:  generateHandlerFiles,
	}
)

const (
	newHandlersDirectoryPermissions = 0755
)

func generateHandlerFiles(cmd *cobra.Command, args []string) error {
	handlername = args[0]
	pathrelativetoroot = path.Join("..", "..")
	handlerDirectory = path.Join(pathrelativetoroot, "pkg", "tnf", "handlers")
	newHandlerDirectory := path.Join(handlerDirectory, handlername)

	err := os.Mkdir(newHandlerDirectory, newHandlersDirectoryPermissions)
	if err != nil {
		return err
	}

	myhandler := myHandler{LowerHandlername: handlername, UpperHandlername: strings.Title(handlername)}

	// pathfile this is the path of the file from template file that will creat

	pathfile := path.Join(handlerDirectory, "handler_template", "doc.tmpl")
	namefile := "" + "doc.go"
	err = createfile(pathfile, namefile, myhandler, newHandlerDirectory) // here creating file by doc.tmpl
	if err != nil {
		return err
	}

	pathfile = path.Join(handlerDirectory, "handler_template", "handler_test.tmpl")
	namefile = "" + myhandler.LowerHandlername + "_test.go"
	err = createfile(pathfile, namefile, myhandler, newHandlerDirectory) // here creating file by template_test.tmpl
	if err != nil {
		return err
	}

	pathfile = path.Join(handlerDirectory, "handler_template", "handler.tmpl")
	namefile = "" + myhandler.LowerHandlername + ".go"
	err = createfile(pathfile, namefile, myhandler, newHandlerDirectory) // here creating file by template.tmpl
	if err != nil {
		return err
	}
	return err
}

func createfile(pathfile, namefile string, myhandler myHandler, newHandlerDirectory string) error {
	ftpl, err := template.ParseFiles(pathfile)
	if err != nil {
		return err
	}

	temp := path.Join(newHandlerDirectory, namefile)
	f, err := os.Create(temp)
	if err != nil {
		return err
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	err = ftpl.Execute(w, myhandler)
	if err != nil {
		return err
	}
	w.Flush()

	return nil
}

func main() {
	rootCmd.AddCommand(generate)
	generate.AddCommand(handler)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
