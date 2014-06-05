package main

import (
	"encoding/json"
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
	transifexApi.Debug = transifexCLI.Debug()
	var err error

	if sourceLang, err = transifexApi.SourceLanguage(); err != nil {
		log.Fatalf("\n\nError loading the transifext project data: \n%s", err)
	}


	files, readFilesErr := config.ReadConfig(transifexCLI.ConfigFile(), rootDir, sourceLang, transifexApi)

	if readFilesErr != nil {
		fmt.Println(rootDir)
		log.Fatalf("\n\nError reading reading language files: \n\n%s", readFilesErr)
	}

	readExistingResources()

	doneChannel := make(chan string, len(files))
	defer close(doneChannel)

	for _, file := range files {
		go upload(doneChannel, file)
	}

	for done := 0; done < len(files); {
		slug := <-doneChannel
		fmt.Printf("\nFINISHED %s\n", slug)
		done++
	}
}

func upload(doneChannel chan string, file config.LocalizationFile) {
	uploadFile(file)
	doneChannel <- file.Slug
}

func readBody(resp http.Response) []byte {
	bytes, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatalf("Failed to read response %s\n", readErr)
	}
	return bytes
}

func loadContent(lang string, file config.LocalizationFile) string {
	filename := file.Translations[lang]
	content, fileErr := ioutil.ReadFile(rootDir + filename)
	if fileErr != nil {
		log.Fatalf("Unable to load file: %s", fileErr)
	}
	switch file.I18nType {
	case transifex.KeyValueJson:
		var data map[string]string
		jsonErr := json.Unmarshal(content, &data)
		if jsonErr != nil {
			log.Fatalf("%s is identifies as %s in the configuration file but is not valid json: %s", filename, file.I18nType, jsonErr)
		}
		for key, value := range data {
			if key == "" {
				delete(data, key)
			}
			if value == "" {
				data[key] = " "
			}
			content, jsonErr = json.Marshal(data)
			if jsonErr != nil {
				panic("An error occurred when encoding json after updating json so that transifex can use it")
			}
		}
	}
	return string(content)
}

func uploadFile(file config.LocalizationFile) {
	slug := file.Slug
	filename := file.Translations[sourceLang]
	content := loadContent(sourceLang, file)
	req := transifex.UploadResourceRequest{file.BaseResource, string(content), "true"}

	if _, has := existingResources[slug]; !has {
		fmt.Printf("Creating new resource: '%s' '%s'\n", filename, slug)
		err := transifexApi.CreateResource(req)
		if err != nil {
			log.Fatalf("Error encountered sending the request to transifex: \n%s'n", err)
		}

		addTranslations(file)

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
	for lang, _ := range file.Translations {
		if lang != sourceLang {
			content := loadContent(lang, file)

			transifexApi.UploadTranslationFile(file.Slug, lang, content)
		}
	}
}
