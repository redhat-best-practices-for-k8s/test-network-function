package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/spf13/cobra"
)

const success = "success"

type myHandler struct {
	Handlername string
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

func generateHandlerFiles(cmd *cobra.Command, args []string) error {
	handlername = args[0]
	pathrelativetoroot = path.Join("..", "..")
	handlerDirectory = path.Join(pathrelativetoroot, "pkg", "tnf", "handlers")
	newHandlerDirectory := path.Join(handlerDirectory, handlername)

	err := os.Mkdir(newHandlerDirectory, 0755)
	if err != nil {
		return err
	}

	myhandler := myHandler{Handlername: handlername}

	// pathfile this is the path of the file from template file that will creat

	pathfile := path.Join(handlerDirectory, "handler_template", "doc.tmpl")
	namefile := "" + "doc.go"
	err = createfile(pathfile, namefile, myhandler, newHandlerDirectory) // here creating file by doc.tmpl
	if err.Error() != success {
		return err
	}

	pathfile = path.Join(handlerDirectory, "handler_template", "handler_test.tmpl")
	namefile = "" + handlername + "_test.go"
	err = createfile(pathfile, namefile, myhandler, newHandlerDirectory) // here creating file by template_test.tmpl
	if err.Error() != success {
		return err
	}

	pathfile = path.Join(handlerDirectory, "handler_template", "handler.tmpl")
	namefile = "" + handlername + ".go"
	err = createfile(pathfile, namefile, myhandler, newHandlerDirectory) // here creating file by template.tmpl
	if err.Error() != success {
		return err
	}
	return fmt.Errorf(success)
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
	return fmt.Errorf(success)
}

func main() {
	rootCmd.AddCommand(generate)
	generate.AddCommand(handler)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
