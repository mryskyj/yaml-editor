package sampleschema

type sampleJSONImport struct {
	Version string           `json:"version"`
	Source  string           `json:"source"`
	Items   []sampleJSONItem `json:"items"`
}

type sampleJSONItem struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type sampleXMLImport struct {
	Version string           `xml:"version,attr"`
	Source  string           `xml:"source"`
	Entries []sampleXMLEntry `xml:"entry"`
}

type sampleXMLEntry struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}
