package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"transifex"
	"transifex/format"
)

type LocalizationFile struct {
	transifex.BaseResource
	Filename     string
	Structure    format.FileLocator
	Format       format.Format
	Translations map[string]string
}

func ReadConfig(configFile, rootDir, sourceLang string, t transifex.TransifexAPI) (files []LocalizationFile, err error) {
	if sourceLang == "" {
		return nil, fmt.Errorf("Source lang is empty.")
	}
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Unable to read %s", configFile)
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
					lang := strings.Split(filepath.Base(name), filename)[0]
					translations[lang] = dir + name
				}
			}

			if _, has := translations[sourceLang]; !has {
				log.Fatalf("%s translations file is required for translation resource: %s/%s", sourceLang, dir, filename)
			}

			name := nextFile["name"].(string)
			slug := nextFile["slug"].(string)
			priority := nextFile["priority"].(string)
			structure := format.FileLocators[nextFile["structure"].(string)]
			fname := nextFile["filename"].(string)
			var categories []string
			for _, c := range nextFile["categories"].([]interface{}) {
				categories = append(categories, c.(string))
			}
			resource := LocalizationFile{
				transifex.BaseResource{slug, name, i18nType, string(priority), strings.Join(categories, " ")},
				fname, structure, format.Formats[i18nType], translations}
			files = append(files, resource)
		}
	}
	return files, nil

}
