package main

import (
	"bufio"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/spf13/cobra"
)

type myHandler struct {
	Handlername string
}

var (
	pathrelativetoroot string
	handlerDirectory   string
	handlername        string

	rootCmd = &cobra.Command{
		Use:   "rootCmd",
		Short: "A CLI for creating, validating, and test-network-function tests.",
	}

	generate = &cobra.Command{
		Use:   "generate",
		Short: "generator tool for various tnf artifacts.",
	}

	handler = &cobra.Command{
		Use:   "handler",
		Short: "adding new handler.",
		Run:   generateHandlerFiles,
	}
)

func generateHandlerFiles(cmd *cobra.Command, args []string) {
	handlername = args[0]
	pathrelativetoroot = path.Join("..", "..")
	handlerDirectory = path.Join(pathrelativetoroot, "pkg", "tnf", "handlers")
	newHandlerDirectory := path.Join(handlerDirectory, handlername)

	err := os.Mkdir(newHandlerDirectory, 0755)
	if err != nil {
		log.Fatal(err)
	}
	myhandler := myHandler{Handlername: handlername}

	// pathfile this is the path of the file from template file that will creat

	pathfile := path.Join(handlerDirectory, "handler_template", "doc.tmpl")
	namefile := "" + "doc.go"
	createfile(pathfile, namefile, myhandler, newHandlerDirectory) // here creating file by doc.tmpl

	pathfile = path.Join(handlerDirectory, "handler_template", "handler_test.tmpl")
	namefile = "" + handlername + "_test.go"
	createfile(pathfile, namefile, myhandler, newHandlerDirectory) // here creating file by template_test.tmpl

	pathfile = path.Join(handlerDirectory, "handler_template", "handler.tmpl")
	namefile = "" + handlername + ".go"
	createfile(pathfile, namefile, myhandler, newHandlerDirectory) // here creating file by template.tmpl
}

func createfile(pathfile, namefile string, myhandler myHandler, newHandlerDirectory string) {
	ftpl, err := template.ParseFiles(pathfile)
	if err != nil {
		log.Fatalln(err)
	}

	temp := path.Join(newHandlerDirectory, namefile)
	f, err := os.Create(temp)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	err = ftpl.Execute(w, myhandler)
	if err != nil {
		panic(err)
	}

	w.Flush()
}

func main() {
	rootCmd.AddCommand(generate)
	generate.AddCommand(handler)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
