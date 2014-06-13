package config

import (
	"path/filepath"
	"testing"
	tu "testutil"
)

func TestReadConfig_LangDir(t *testing.T) {
	configText := `
[{
	"type": "FLATTENXMLTOJSON",
    "structure": "3-CHAR-LOC-DIR",
    "resources": [{
      "dir": "loc",
      "fname": "string",
      "name": "Iso19139 Strings",
      "slug": "iso19139-strings-xml",
      "priority": "0",
      "categories": ["schemaplugin", "iso19139"]
    }]
}]`
	root := tu.CreateFileTree(
		tu.Dir("xyz",
			tu.FileAndData("config.json", []byte(configText)),
			tu.Dir("loc",
				tu.Dir("eng", tu.File("string.xml")),
				tu.Dir("fre", tu.File("string.xml")),
				tu.Dir("ger", tu.File("xyz")))))

	files, err := ReadConfig(filepath.Join(root, "config.json"), root, "en")

	if err != nil {
		t.Errorf("Error reading config. %v", err)
		panic(err)

	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file: %v", files)
	}

	if len(files[0].Translations) != 2 {
		t.Errorf("Expected 2 translations: %v", files[0].Translations)
	}

	if v, ok := files[0].Translations["en"]; !ok {
		t.Errorf("Expected en translation: %v", files[0].Translations)
	} else if v != filepath.Join(root, "loc", "eng", "string.xml") {
		t.Errorf("Path of en translation was unexpected: %s", v)
	}

	if v, ok := files[0].Translations["fr"]; !ok {
		t.Errorf("Expected fr translation: %v", files[0].Translations)
	} else if v != filepath.Join(root, "loc", "fre", "string.xml") {
		t.Errorf("Path of fr translation was unexpected: %s", v)
	}
}

func Test_UnmappedLang(t *testing.T) {
	configText := `
[{
	"type": "FLATTENXMLTOJSON",
    "structure": "3-CHAR-LOC-DIR",
    "resources": [{
      "dir": "loc",
      "fname": "string",
      "name": "Iso19139 Strings",
      "slug": "iso19139-strings-xml",
      "priority": "0",
      "categories": ["schemaplugin", "iso19139"]
    }]
}]`
	root := tu.CreateFileTree(
		tu.Dir("xyz",
			tu.FileAndData("config.json", []byte(configText)),
			tu.Dir("loc",
				tu.Dir("eng", tu.File("string.xml")),
				tu.Dir("fre", tu.File("string.xml")),
				tu.Dir("ger", tu.File("xyz")),
				tu.Dir("xxz", tu.File("string.xml")))))

	defer func() {
		if r := recover(); r != nil {
			// panic is expected
		}
	}()

	ReadConfig(filepath.Join(root, "config.json"), root, "en")

	t.Errorf("A panic should have occurred because xxz is not a valid lang code")
}
