package format

import (
	"fmt"
	"encoding/json"
	"transifex/config"
	"path/filepath"
	"io/ioutil"
)
type KeyValueJson struct{}

func (f KeyValueJson) Clean(content []byte) ([]byte, string, error) {
	var data map[string]string
	jsonErr := json.Unmarshal(content, &data)
	if jsonErr != nil {
		return nil, "", fmt.Errorf("Not valid json: %s", jsonErr)
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

	return content, "KEYVALUEJSON", nil

}
func (f KeyValueJson) Write(rootDir, langCode, translation string, file config.LocalizationFile) error {
	path, ok := file.Translations[langCode]
	if !ok {
		for _, path := range file.Translations {
			basicPath := filepath.Dir(path)
			fileName := fmt.Sprintf("%s-%s.json", langCode, file.Slug)
			path = filepath.Join(basicPath, fileName)
			break;
		}
	}
	path = filepath.Join(rootDir, path)
	fmt.Println("Updating translations file: " + path)
	return ioutil.WriteFile(path, []byte(translation), 0644)
}