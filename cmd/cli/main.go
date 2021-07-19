package main

import (
	"bufio"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/spf13/cobra"
)

type handler struct {
	X string
}

var (
	pathrelativetoroot string
	handlerDirectory   string
	handlername        string
	rootCmd            = &cobra.Command{
		Use:   "kind",
		Short: "kind is a tool for managing local Kubernetes clusters",
		Long:  "kind creates and manages local Kubernetes clusters using Docker container 'nodes'",
		Run: func(cmd *cobra.Command, args []string) {

			handlername = args[len(args)-1]
			pathrelativetoroot = path.Join("..", "..")
			handlerDirectory = path.Join(pathrelativetoroot, "pkg", "tnf", "handlers")
			new_handler_directory := path.Join(handlerDirectory, handlername)

			os.Mkdir(new_handler_directory, 0755)

			// create 3 files by template

			my_handler := handler{X: handlername}

			path_file := path.Join(handlerDirectory, "handler_template", "doc.tmpl")
			namefile := "doc.tmpl"
			createfile(path_file, namefile, my_handler, new_handler_directory)

			path_file = path.Join(handlerDirectory, "handler_template", "template_test.tmpl")
			namefile = "" + handlername + "_test.tmpl"
			createfile(path_file, namefile, my_handler, new_handler_directory)

			path_file = path.Join(handlerDirectory, "handler_template", "template.tmpl")
			namefile = "" + handlername
			createfile(path_file, namefile, my_handler, new_handler_directory)

		},
	}
)

func createfile(path_file string, namefile string, my_handler handler, new_handler_directory string) {
	f_tpl, err := template.ParseFiles(path_file)
	if err != nil {
		log.Fatalln(err)
	}

	temp := path.Join(new_handler_directory, namefile)

	f, err := os.Create(temp)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	err = f_tpl.Execute(w, my_handler)
	if err != nil {
		panic(err)
	}
	w.Flush()
}

func main() {
	rootCmd.SetArgs(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

}
