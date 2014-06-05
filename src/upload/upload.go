package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"transifex"
	"transifex/cli"
	"transifex/config"
)

var sourceLang string
var rootDir string
var transifexApi transifex.TransifexAPI
var existingResources = make(map[string]bool)

func main() {
	transifexCLI := cli.NewCLI()
	transifexApi = transifex.NewTransifexAPI(transifexCLI.ProjectSlug(), transifexCLI.Username(), transifexCLI.Password())
	rootDir = transifexCLI.RootDir()
	// transifexApi.Debug = true
	var err error
	if err = transifexApi.ValidateConfiguration(); err != nil {
		log.Fatalf(err.Error())
	}

	if sourceLang, err = transifexApi.SourceLanguage(); err != nil {
		log.Fatalf("Error loading the transifext project data.")
	}

	files, readFilesErr := config.ReadConfig(transifexCLI.ConfigFile(), rootDir, sourceLang, transifexApi)

	if readFilesErr != nil {
		fmt.Println(rootDir)
		log.Fatalf("Error reading reading language files: \n\n%s", readFilesErr)
	}


	readExistingResources()

	doneChannel := make(chan string, len(files))
	defer close(doneChannel)

	for _, file := range files {
		go func() {
			uploadFile(file)
			addTranslations(file)
			doneChannel <- ""
		}()
	}

	for done := 0; done < len(files); {
		<-doneChannel

		done++
	}
}

func readBody(resp http.Response) []byte {
	bytes, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatalf("Failed to read response %s\n", readErr)
	}
	return bytes
}

func uploadFile(file config.LocalizationFile) {
	slug := file.Slug
	filename := file.Translations[sourceLang]
	content, fileErr := ioutil.ReadFile(rootDir + filename)
	if fileErr != nil {
		log.Fatalf("Unable to load file: %s", fileErr)
	}
	req := transifex.UploadResourceRequest{file.BaseResource, string(content), "true"}

	if _, has := existingResources[slug]; !has {
		fmt.Printf("Creating new resource: '%s' '%s'\n", filename, slug)
		err := transifexApi.CreateResource(req)
		if err != nil {
			log.Fatalf("Error encountered sending the request to transifex: \n%s'n", err)
		}

		fmt.Printf("Finished Adding '%s'\n", slug)
	} else {
		if err := transifexApi.UpdateResourceContent(slug, string(content)); err != nil {
			log.Fatalf("Error updating content")
		}

		fmt.Printf("Finished Updating '%s'\n", slug)
	}
}

func readExistingResources() {
	resources, err := transifexApi.ListResources()
	if err != nil {
		log.Fatalf("Unable to load resources: %s", err)
	}
	for _, res := range resources {
		existingResources[res.Slug] = true
	}
}

func addTranslations(file config.LocalizationFile) {
	for lang, translationFile := range file.Translations {
		if lang != sourceLang {
			content, fileErr := ioutil.ReadFile(rootDir + translationFile)
			if fileErr != nil {
				log.Fatalf("Unable to load file: %s", fileErr)
			}

			transifexApi.UploadTranslationFile(file.Slug, lang, string(content))
		}
	}
}
