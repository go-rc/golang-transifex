package format

import (
	"path/filepath"
	"testing"
	tu "testutil"
)

func Test_LangNameLocator_Find(t *testing.T) {
	l := LangNameLocator{}
	tu.AssertEquals("", filepath.Join("xyz", "fr-name.json"), l.Find("xyz", "fr", "name", "json"), t)
	tu.AssertEquals("", filepath.Join("xyz", "en-name.json"), l.Find("xyz", "en", "name", "json"), t)
	tu.AssertEquals("", filepath.Join("xyz", "abc", "en-name.xml"), l.Find("xyz/abc", "en", "name", "xml"), t)
	tu.AssertEquals("", "en-name.xml", l.Find("", "en", "name", "xml"), t)

}
