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
	X string
}

var (
	pathrelativetoroot string
	handlerDirectory   string
	handlername        string

	rootCmd = &cobra.Command{
		Use:   "rootCmd",
		Short: "test network function command line tools",
	}

	generate = &cobra.Command{
		Use:   "generate",
		Short: "generator tool for various tnf artifacts, such as handler code, catalog etc",
	}

	handler = &cobra.Command{
		Use:   "handler",
		Short: "adding new handler",
		Run:   addingHandler,
	}
)

func addingHandler(cmd *cobra.Command, args []string) {

	handlername = args[0]
	pathrelativetoroot = path.Join("..", "..")
	handlerDirectory = path.Join(pathrelativetoroot, "pkg", "tnf", "handlers")
	newHandlerDirectory := path.Join(handlerDirectory, handlername)

	os.Mkdir(newHandlerDirectory, 0755)

	// create 3 files by template

	myhandler := myHandler{X: handlername}

	pathfile := path.Join(handlerDirectory, "handler_template", "doc.tmpl")
	namefile := "doc.tmpl"
	createfile(pathfile, namefile, myhandler, newHandlerDirectory)

	pathfile = path.Join(handlerDirectory, "handler_template", "template_test.tmpl")
	namefile = "" + handlername + "_test.tmpl"
	createfile(pathfile, namefile, myhandler, newHandlerDirectory)

	pathfile = path.Join(handlerDirectory, "handler_template", "template.tmpl")
	namefile = "" + handlername
	createfile(pathfile, namefile, myhandler, newHandlerDirectory)
}

func createfile(pathfile string, namefile string, myhandler myHandler, newHandlerDirectory string) {

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
