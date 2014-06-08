package format

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// A Format is responsible for cleaning up the translation file.  For example:
// * A json file with empty string key should be deleted
// * A json file with empty values should be converted to a string with a single space
// * An xml file with nested tags should be flattened
type Format interface {
	// takes the raw data read from file
	// returns:
	// * cleaned bytes
	// * new i18nType - it is possible that the format decided that in order to make the format useable it needed to convert it to a new type
	// * an error or nil if no errors occurred
	Clean([]byte) ([]byte, string, error)
	// Write a new translation to the correct translation file
	// * rootDir - path to the root of the translation files directory tree
	// * langCode - the language code of the translation
	// * srcLang - the source language 
	// * translation - the translation text
	// * fileLocator - strategy for creating the path to the translation file
	Write(rootDir, langCode, srcLang, filename, translation string, fileLocator FileLocator) error
}

var Formats = map[string]Format{"KEYVALUEJSON": new(KeyValueJson), "FLATTENXMLTOJSON": new(FlattenXmlToJson)}

// Strategy for creating the path to a translation file
type FileLocator interface {
	// path - the path to the directory containing the translation files (or root of the tree for looking up the files)
	// lang - the language of the translation file
	// name - the name of the file (or base name)
	// ext - the extension of the file
	Find(path, lang, name, ext string) string
	List(path, name, ext string) ([]string, error)
}

var FileLocators = map[string]FileLocator{"LANG-NAME": new(LangNameLocator)}

// All translation files in same directory and have pattern: lang-name.ext
type LangNameLocator struct{}

func (l LangNameLocator) Find(path, lang, name, ext string) string {
	fileName := fmt.Sprintf("%s-%s.%s", lang, name, ext)
	return filepath.Join(path, fileName)
}

func (l LangNameLocator) List(path, name, ext string) ([]string, error) {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	filename := fmt.Sprintf("-%s.%s", name, ext)

	candidates, readErr := ioutil.ReadDir(path)

	if readErr != nil {
		return nil, readErr
	}

	translationFiles := []string{}

	for _, f := range candidates {
		fname := f.Name()
		if strings.HasSuffix(fname, filename) {
			translationFiles = append(translationFiles, fname)
		}
	}

	return translationFiles, nil
}
