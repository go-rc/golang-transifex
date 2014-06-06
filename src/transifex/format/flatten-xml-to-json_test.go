package format

import (
	"encoding/json"
	"testing"
)
func Test_Clean (t *testing.T) {
	xmlData := `<a><b at="atv" at="atv2"><l>value1</l><d at="atv" at2="atv2">dv</d></b><c>value2</c>`
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
		case "a b[at=atv and at=atv2] l":
			assertEquals("itemValue0", "value1", item, t)
		case "a b[at=atv and at=atv2] d[at=atv and at2=atv2]":
			assertEquals("itemValue1", "dv", item, t)
		case "a c":
			assertEquals("itemValue2", "value2", item, t)
		default:
			t.Errorf("Unexpected key: %s in %s", key, prop)
		}
	}
}

func assertEquals(msg, expected, actual string, t *testing.T) {
			if actual != expected {
				t.Errorf("%s: Expected/Actual \n%q\n%q", msg, expected, actual)
			}
}