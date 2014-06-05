package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"transifex"
)

type LocalizationFile struct {
	transifex.BaseResource
	translations map[string]string
}

var projectSlug = flag.String("project", "", "REQUIRED - the transifex project slug")
var configFile = flag.String("config", "", "REQUIRED - The location of the configuration file")
var sourceLang string
var username = flag.String("username", "", "The transifex username")
var password = flag.String("password", "", "The transifex password")

var rootDir string
var client = &http.Client{}
var transifexApi transifex.TransifexAPI
var existingResources = make(map[string]bool)

func init() {
	flag.Parse()
	if *configFile == "" {
		fmt.Printf("The 'config' flag is required.  \n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *projectSlug == "" {
		fmt.Printf("The 'project' flag is required.  \n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	rootDir = filepath.Dir(*configFile)

	if !strings.HasSuffix(rootDir, "/") {
		rootDir = rootDir + "/"
	}
}

func readFiles() (files []LocalizationFile, err error) {
	bytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Unable to read %s", *configFile)
		return nil, err
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(bytes, &jsonData); err != nil {
		return nil, err
	}

	for i18nType, array := range jsonData {
		for _, nextFileRaw := range array.([]interface{}) {
			nextFile := nextFileRaw.(map[string]interface{})
			dir := nextFile["dir"].(string)
			if !strings.HasSuffix(dir, "/") {
				dir += "/"
			}
			filename := "-" + nextFile["filename"].(string) + ".json"

			candidates, readErr := ioutil.ReadDir(rootDir + dir)

			if readErr != nil {
				return nil, readErr
			}

			translations := make(map[string]string)
			for _, file := range candidates {
				name := file.Name()
				if !file.IsDir() && strings.HasSuffix(name, filename) {
					translations[strings.Split(filepath.Base(name), "-")[0]] = dir + name
				}
			}

			if _, has := translations[sourceLang]; !has {
				log.Fatalf("%s translations file is required for translation resource: %s/%s", sourceLang, dir, filename)
			}

			name := nextFile["name"].(string)
			slug := nextFile["slug"].(string)
			priority := nextFile["priority"].(string)
			var categories []string
			for _, c := range nextFile["categories"].([]interface{}) {
				categories = append(categories, c.(string))
			}
			resource := LocalizationFile{
				transifex.BaseResource{slug, name, i18nType, string(priority), strings.Join(categories, " ")},
				translations}
			files = append(files, resource)
		}
	}
	return files, nil
}

func readBody(resp http.Response) []byte {
	bytes, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatalf("Failed to read response %s\n", readErr)
	}
	return bytes
}

func uploadFile(file LocalizationFile) {
	slug := file.Slug
	filename := file.translations[sourceLang]
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

func readAuth(field *string, prompt string) {

	if *field == "" {
		var line string
		var readlineErr error
		in := bufio.NewReader(os.Stdin)
		fmt.Printf("Enter your %s: ", prompt)
		if line, readlineErr = in.ReadString('\n'); readlineErr != nil {
			log.Fatalf("Failed to read %s", prompt)
		}

		*field = strings.TrimSpace(line)
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

func addTranslations(file LocalizationFile) {
	for lang, translationFile := range file.translations {
		if lang != sourceLang {
			content, fileErr := ioutil.ReadFile(rootDir + translationFile)
			if fileErr != nil {
				log.Fatalf("Unable to load file: %s", fileErr)
			}

			transifexApi.UploadTranslationFile(file.Slug, lang, string(content))
		}
	}
}

func main() {
	readAuth(username, "username")
	readAuth(password, "password")

	transifexApi = transifex.NewTransifexAPI(*projectSlug, *username, *password)
	// transifexApi.Debug = true
	var err error
	if err = transifexApi.ValidateConfiguration(); err != nil {
		log.Fatalf(err.Error())
	}

	if sourceLang, err = transifexApi.SourceLanguage(); err != nil {
		log.Fatalf("Error loading the transifext project data.")
	}

	files, readFilesErr := readFiles()

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
