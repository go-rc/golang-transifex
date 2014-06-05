package transifex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type TransifexAPI struct {
	ApiUrl, Project, username, password string
	client                              *http.Client
	Debug bool
}

type BaseResource struct {
	Slug       string   `json:"slug"`
	Name       string   `json:"name"`
	I18nType   string   `json:"i18n_type"`
	Priority   string   `json:"priority"`
	Category   string   `json:"category"`
}
type Resource struct {
	BaseResource
	SourceLanguage string `json:"source_language_code"`
}
type UploadResourceRequest struct {
	BaseResource
	Content             string `json:"content"`
	Accept_translations string   `json:"accept_translations"`
}


func NewTransifexAPI(project, username, password string) TransifexAPI {
	return TransifexAPI{"https://www.transifex.com/api/2/", project, username, password, &http.Client{}, false}
}

func (t TransifexAPI) ListResources() ([]Resource, error) {
	resp, err := t.execRequest("GET", t.resourcesUrl(true), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, err
	}
	var resources []Resource
	jsonErr := json.Unmarshal(data, &resources)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return resources, nil
}

func (t TransifexAPI) CreateResource(newResource UploadResourceRequest) error {
	data, marshalErr := json.Marshal(newResource)
	if marshalErr != nil {
		return marshalErr
	}

	resp, err := t.execRequest("POST", t.resourcesUrl(false), bytes.NewReader(data))
	if err != nil {
		return err
	}

	checkData, checkErr := t.checkValidJsonResponse(resp, fmt.Sprintf("Failed to create resource: %s\n", newResource.Slug))
	checkDataArray := checkData.([]interface{})
	fmt.Printf(`Create Resource Summary:

Strings Added: %v
Strings updated: %v
Strings deleted: %v

`, checkDataArray[0], checkDataArray[1], checkDataArray[2])
	return checkErr
}

func (t TransifexAPI) UpdateResourceContent(slug, content string) error {
	data, marshalErr := json.Marshal(map[string]string{"slug":slug, "content":content})
	if marshalErr != nil {
		return marshalErr
	}

	resp, err := t.execRequest("PUT", t.resourceUrl(slug, true) + "content/", bytes.NewReader(data))
	if err != nil {
		return err
	}

	checkData, checkErr := t.checkValidJsonResponse(resp, fmt.Sprintf("Error updating content of %s", slug))
	dataMap := checkData.(map[string]interface{})
	fmt.Printf(`Update Source Language Content Summary:

Strings Added: %v
Strings updated: %v
Strings deleted: %v

`, dataMap["strings_added"], dataMap["strings_updated"], dataMap["strings_delete"])
	return checkErr
}

func (t TransifexAPI) ValidateConfiguration() error {
	msg := "Error occurred when checking credentials. Please check credentials and network connection"
	if _, err := t.SourceLanguage(); err != nil {
		return fmt.Errorf(msg)
	}
	return nil	
}

func (t TransifexAPI) UploadTranslationFile(slug, langCode, content string) error {
	data, marshalErr := json.Marshal(map[string]string{"content":content})
	if marshalErr != nil {
		return marshalErr
	}

	resp, err := t.execRequest("PUT", t.resourceUrl(slug, true) + "translation/" + langCode + "/", bytes.NewReader(data))
	if err != nil {
		return err
	}

	checkData, checkErr := t.checkValidJsonResponse(resp, fmt.Sprintf("Error adding %s translations for %s", langCode, slug))
	dataMap := checkData.(map[string]interface{})
	fmt.Printf(`Update %s Translation summary:

Strings Added: %v
Strings updated: %v
Strings deleted: %v

`, langCode, dataMap["strings_added"], dataMap["strings_updated"], dataMap["strings_delete"])
	return checkErr
}

func (t TransifexAPI) SourceLanguage() (string, error) {
	resp, err := t.execRequest("GET", "https://www.transifex.com/api/2/project/"+t.Project, nil)
	if err != nil {
		return "", err
	}
	jsonData, checkErr := t.checkValidJsonResponse(resp, "Error while reading project description");
	if checkErr != nil {
		return "", checkErr
	}
	lang := jsonData.(map[string]interface{})["source_language_code"]
	if lang == nil {
		return "", fmt.Errorf("An error occurred while reading response. Expected a 'source_language_code' json field:\n%s", jsonData)
	}
	return lang.(string), nil
}
// func (t TransifexAPI) DownloadTranslations(slug string) []string {
// 	t.getJson(fmt.Sprintf(t.ApiUrl+"project/%s/languages"))
// }
func (t TransifexAPI) execRequest(method string, url string, requestData io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, url, requestData)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(t.username, t.password)
	if requestData != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	if t.Debug {
		func () {
			var dump, _ = httputil.DumpRequest(request, true)
			fmt.Println(string(dump))
		}()
	}

	fmt.Printf("\nExecuting http %s request: '%s'\n\n", method, url)
	return t.client.Do(request)
}

func (t TransifexAPI) resourcesUrl(endSlash bool) string {
	url := fmt.Sprintf(t.ApiUrl+"project/%s/resources", t.Project)
	if endSlash {
		return url + "/"
	} 
	return url
}
func (t TransifexAPI) resourceUrl(slug string, endSlash bool) string {
	url := fmt.Sprintf(t.ApiUrl+"project/%s/resource/%s", t.Project, slug)
	if endSlash {
		return url + "/"
	} 
	return url
}

func (t TransifexAPI) checkValidJsonResponse(resp *http.Response, errorMsg string) (interface{}, error) {

	responseData, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		return nil, readErr
	}

	var jsonData interface{}
	if err := json.Unmarshal(responseData, &jsonData); err != nil {
		return nil, fmt.Errorf(errorMsg + "\n\nError:\n" + string(responseData))
	}

	if t.Debug {

			fmt.Printf("Response: %s", jsonData)
		switch jsonData.(type) {
		case map[string]interface{}:
			for key, val := range jsonData.(map[string]interface{}) {
				fmt.Println(key,":",val)
			}
		case []interface{}:
			for _, val := range jsonData.([]interface{}) {
				fmt.Println(val)
			}
		default:
			fmt.Printf("Response: %s", jsonData)
		}
	}

	fmt.Println("\n")
	return jsonData, nil
}
