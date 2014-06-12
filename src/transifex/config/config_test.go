package config

import (
	tu "testutil"
	"testing"
	"path/filepath"
	"strings"
)

func TestReadConfig_LangDir(t *testing.T) {
	configText := `
[{
	"type": "FLATTENXMLTOJSON",
    "structure": "3-CHAR-LOC-DIR",
    "resources": [{
      "dir": "web/src/main/webapp/WEB-INF/data/config/schema_plugins/iso19139/loc",
      "fname": "strings",
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

	files, err := ReadConfig(filepath.Join(root, "config.json"), root, "en")

	if err != nil {
		t.Errorf("Error reading config. %v", err)
		panic(err)

	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file: %v", files)
	}

	if len(files[0].Translations) != 2{
		t.Errorf("Expected 2 translations: %v", files[0].Translations)
	}

	if v, ok := files[0].Translations["en"]; !ok {
		t.Errorf("Expected en translation: %v", files[0].Translations)
	} else if !strings.HasSuffix(v, filepath.Join("root", "eng", "strings.xml")) {
		t.Errorf("Path of en translation was unexpected: %s", v)
	}

	if v, ok := files[0].Translations["fr"]; !ok {
		t.Errorf("Expected fr translation: %v", files[0].Translations)
	} else if !strings.HasSuffix(v, filepath.Join("root", "fre", "strings.xml")) {
		t.Errorf("Path of fr translation was unexpected: %s", v)
	}


}
