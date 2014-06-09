package format

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"path/filepath"
	"testing"
	tu "testutil"
)

const xmlData = `<a><b at="atv" at2="atv2"><l>value1</l><d at="atv" at2="atv2">dv</d></b><c>value2</c></a>`

type a struct {
	B b `xml:"b"`
	C c `xml:"c"`
}

type c struct {
	Text string `xml:",chardata"`
}
type b struct {
	At  string `xml:"at,attr"`
	At2 string `xml:"at2,attr"`
	L   l      `xml:"l"`
	D   d      `xml:"d"`
}
type l struct {
	Text string `xml:",chardata"`
}
type d struct {
	At   string `xml:"at,attr"`
	At2  string `xml:"at2,attr"`
	Text string `xml:",chardata"`
}

func Test_Clean(t *testing.T) {
	cleaned, i18, err := FlattenXmlToJson{}.Clean([]byte(xmlData))

	if err != nil {
		t.Error(err)
	}
	if i18 != "KEYVALUEJSON" {
		t.Error("Expecte i18 to be KEYVALUEJSON")
	}
	var prop map[string]string
	json.Unmarshal(cleaned, &prop)

	if len(prop) != 3 {
		t.Errorf("Expected 3 items but found %s in %s", len(prop), prop)
	}

	for key, item := range prop {
		switch key {
		case "a b[at=atv and at2=atv2] l":
			tu.AssertEquals("itemValue0", "value1", item, t)
		case "a b[at=atv and at2=atv2] d[at=atv and at2=atv2]":
			tu.AssertEquals("itemValue1", "dv", item, t)
		case "a c":
			tu.AssertEquals("itemValue2", "value2", item, t)
		default:
			t.Errorf("Unexpected key: %s in %s", key, prop)
		}
	}
}

func Test_WriteUpdateExisting(t *testing.T) {
	translation := make(map[string]string)
	translation["a b[at=atv and at2=atv2] l"] = "newvalue1"
	translation["a b[at=atv and at2=atv2] d[at=atv and at2=atv2]"] = "newdv"
	translation["a c"] = "newvalue2"

	translationAsBytes, _ := json.Marshal(translation)

	tmpDir, _ := ioutil.TempDir("", "test")
	fileToUpdate := filepath.Join(tmpDir, "en-name.xml")
	if err := ioutil.WriteFile(fileToUpdate, []byte(xmlData), 644); err != nil {
		panic(err)
	}

	err := FlattenXmlToJson{}.Write(tmpDir, "en", "en", "name", string(translationAsBytes), FileLocators["LANG-NAME"])
	if err != nil {
		t.Errorf("Unexpected error occurred. %s", err)
	}

	xmlData, _ := ioutil.ReadFile(fileToUpdate)

	var data a
	if err = xml.Unmarshal(xmlData, &data); err != nil {
		t.Errorf("Unable to parse result xml: %q\n\n%s", err, string(xmlData))
	}

	tu.AssertEquals("", "newvalue1", data.B.L.Text, t)
	tu.AssertEquals("", "newdv", data.B.D.Text, t)
	tu.AssertEquals("", "newvalue2", data.C.Text, t)
}

func Test_WriteCreateNew(t *testing.T) {
	translation := make(map[string]string)
	translation["a b[at=atv and at2=atv2] l"] = "newvalue1"
	translation["a b[at=atv and at2=atv2] d[at=atv and at2=atv2]"] = "newdv"
	translation["a c"] = "newvalue2"

	translationAsBytes, _ := json.Marshal(translation)

	tmpDir, _ := ioutil.TempDir("", "test")
	fileToUpdate := filepath.Join(tmpDir, "en-name.xml")
	if err := ioutil.WriteFile(fileToUpdate, []byte(xmlData), 644); err != nil {
		panic(err)
	}

	err := FlattenXmlToJson{}.Write(tmpDir, "fr", "en", "name", string(translationAsBytes), FileLocators["LANG-NAME"])
	if err != nil {
		t.Errorf("Unexpected error occurred. %s", err)
	}

	xmlData, _ := ioutil.ReadFile(filepath.Join(tmpDir, "fr-name.xml"))

	var data a
	if err = xml.Unmarshal(xmlData, &data); err != nil {
		t.Errorf("Unable to parse result xml: %q\n\n%s", err, string(xmlData))
	}

	tu.AssertEquals("", "newvalue1", data.B.L.Text, t)
	tu.AssertEquals("", "newdv", data.B.D.Text, t)
	tu.AssertEquals("", "newvalue2", data.C.Text, t)
}
