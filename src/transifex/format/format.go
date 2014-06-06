package format

import "transifex/config"
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
	Write(rootDir, langCode, translation string, file config.LocalizationFile) error
}

var Formats = map[string]Format{"KEYVALUEJSON": new(KeyValueJson), "FLATTENXMLTOJSON": new(FlattenXmlToJson)}
