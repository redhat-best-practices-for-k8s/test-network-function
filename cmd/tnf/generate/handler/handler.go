package handler

import (
	"bufio"
	"os"
	"path"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type myHandler struct {
	UpperHandlername string
	LowerHandlername string
}

const (
	envHandlersFolder  = "TNF_HANDLERS_SRC"
	docFileName        = "doc.go"
	handlerFolderPerms = 0755
)

var (
	handler = &cobra.Command{
		Use:   "handler",
		Short: "adding new handler.",
		RunE:  generateHandlerFiles,
	}
	defaultHandlersFolder = path.Join("pkg", "tnf", "handlers")
)

func getHandlersDirectory() (string, error) {
	handlersDirectory := os.Getenv(envHandlersFolder)

	if handlersDirectory == "" {
		log.Warnf("Environment variable %s not set. Handlers base folder will be set to ./%s",
			envHandlersFolder, defaultHandlersFolder)

		handlersDirectory = defaultHandlersFolder
	} else {
		log.Infof("Env var %s found. Handlers directory: %s", envHandlersFolder, handlersDirectory)
	}

	// Convert to absolute path.
	if !path.IsAbs(handlersDirectory) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		handlersDirectory = path.Join(cwd, handlersDirectory)
	}

	return handlersDirectory, nil
}

func generateHandlerFilesFromTemplates(handlerTemplatesDirectory, newHandlerDirectory string, myhandler myHandler) error {
	type fileToRender struct {
		templatePath     string
		renderedFileName string
	}

	filesToRender := []fileToRender{
		{templatePath: path.Join(handlerTemplatesDirectory, "doc.tmpl"), renderedFileName: docFileName},
		{templatePath: path.Join(handlerTemplatesDirectory, "handler_test.tmpl"), renderedFileName: myhandler.LowerHandlername + "_test.go"},
		{templatePath: path.Join(handlerTemplatesDirectory, "handler.tmpl"), renderedFileName: myhandler.LowerHandlername + ".go"},
	}

	for _, renderedFileName := range filesToRender {
		if err := createfile(renderedFileName.templatePath, renderedFileName.renderedFileName, myhandler, newHandlerDirectory); err != nil {
			log.Errorf("Unable to create rendered file %s on %s", renderedFileName, newHandlerDirectory)
			return err
		}
	}

	return nil
}

func generateHandlerFiles(cmd *cobra.Command, args []string) error {
	handlername := args[0]
	myhandler := myHandler{LowerHandlername: strings.ToLower(handlername), UpperHandlername: strings.Title(handlername)}

	handlersDirectory, err := getHandlersDirectory()
	if err != nil {
		log.Fatalf("Unable to get handlers path.")
		return err
	}

	handlerTemplatesDirectory := path.Join(handlersDirectory, "handler_template")

	log.Infof("Using absolute path for tnf handlers directory: %s", handlersDirectory)
	newHandlerDirectory := path.Join(handlersDirectory, myhandler.LowerHandlername)

	err = os.Mkdir(newHandlerDirectory, handlerFolderPerms)
	if err != nil {
		log.Fatal("Unable to create handler directory " + newHandlerDirectory)
		os.Exit(1)
	}

	err = generateHandlerFilesFromTemplates(handlerTemplatesDirectory, newHandlerDirectory, myhandler)
	if err != nil {
		return err
	}

	log.Infof("Handler files for %s successfully created in %s\n", myhandler.UpperHandlername, path.Join(newHandlerDirectory))
	return nil
}

func createfile(templateFilePath, outputFileName string, myhandler myHandler, newHandlerDirectory string) error {
	ftpl, err := template.ParseFiles(templateFilePath)
	if err != nil {
		return err
	}

	temp := path.Join(newHandlerDirectory, outputFileName)
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

func Execute() *cobra.Command {
	return handler
}
