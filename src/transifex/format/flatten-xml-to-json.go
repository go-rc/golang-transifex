package format

import (
	"transifex/config"
	"strings"
	"encoding/xml"
	"encoding/json"
	"io"
	"bytes"
)

// A format which takes nested xml that has the translations as the text of the leaf nodes
// since the xml can be nested and the nodes can have attributes, the path to the leaf node is encoded
// as the key of the Json.  
// the format that is uploaded to transifex will be a valid Json key value formatted file
type FlattenXmlToJson struct {}

func (f FlattenXmlToJson) Clean(content []byte) ([]byte, string, error) {

	r := strings.NewReader(string(content))
	parser := xml.NewDecoder(r)

	contentJson := make(map[string]string)

	key := []string{}
	for {
		token, err := parser.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, "", err
		}
		switch t := token.(type) {
		case xml.StartElement:
			elmt := xml.StartElement(t)
			name := elmt.Name.Local
			nodeRep := bytes.Buffer{}
			nodeRep.WriteString(name)
			if len(elmt.Attr) > 0 {
				nodeRep.WriteString("[")
				for i, att := range elmt.Attr {
					if i > 0 {
						nodeRep.WriteString(" and ")
					}
					nodeRep.WriteString(att.Name.Local)
					nodeRep.WriteString("=")
					nodeRep.WriteString(att.Value)
				}
				nodeRep.WriteString("]")
			}
			key = append(key, nodeRep.String())
		case xml.EndElement:
			key = key[:len(key) - 1]
		case xml.CharData:
			text := string(xml.CharData(t))
			fkey := strings.Join(key, " ")
			contentJson[fkey] = text
			
		}
	}

	var err error
	content, err = json.Marshal(contentJson)
	if err != nil {
		return nil, "", err
	}
	return content, "KEYVALUEJSON", nil
}

func (f FlattenXmlToJson) Write(rootDir, langCode, translation string, file config.LocalizationFile) error {
	return nil
}