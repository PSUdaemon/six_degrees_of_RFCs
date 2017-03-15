package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

type RFCidx struct {
	RFCs []struct {
		Name    string   `xml:"doc-id"`
		Authors []string `xml:"author>name"`
	} `xml:"rfc-entry"`
}

func main() {
	var idx RFCidx
	var author_count, rfc_count uint

	d, _ := ioutil.ReadFile("rfc-index.xml") // https://www.rfc-editor.org/rfc-index.xml
	xml.Unmarshal(d, &idx)

	vertices := make(map[string]map[string]bool)
	edges := make(map[string]map[string]bool)

	for _, rfc := range idx.RFCs {
		edges[rfc.Name] = make(map[string]bool)
		rfc_count++
		for _, author := range rfc.Authors {
			if _, ok := vertices[author]; !ok {
				vertices[author] = make(map[string]bool)
				author_count++
			}
			vertices[author][rfc.Name] = true
			edges[rfc.Name][author] = true
		}
	}

	fmt.Printf("Ingested file rfc-index.xml\nProcessed %d RFCs\nProcessed %d unique authors\n", rfc_count, author_count)
	fmt.Printf("%#v\n", vertices["J. Postel"])
}
