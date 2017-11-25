package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sync"
)

type RFCidx struct {
	RFCs []struct {
		Name    string   `xml:"doc-id"`
		Authors []string `xml:"author>name"`
	} `xml:"rfc-entry"`
}

type ietf_object struct {
	id    uint
	dist  uint64
	paths [][]*ietf_object
	mutex sync.Mutex
	name  string
	links []*ietf_object
}

func do_work(wg *sync.WaitGroup, dist uint64, obj *ietf_object, path []*ietf_object) {
	var done bool

	obj.mutex.Lock()
	if dist < obj.dist {
		obj.dist = dist
		obj.paths = [][]*ietf_object{path}
	} else if dist == obj.dist {
		obj.paths = append(obj.paths, path)
	} else {
		done = true
	}
	obj.mutex.Unlock()

	if !done {
		for _, doc_obj := range obj.links {
			for _, next_auth_obj := range doc_obj.links {
				wg.Add(1)
				go do_work(wg, dist+1, next_auth_obj, append(path, obj, doc_obj))
			}
		}
	}

	wg.Done()
}

func main() {
	var idx RFCidx
	var author_count, document_count uint
	var authors, documents []ietf_object
	var wg sync.WaitGroup

	d, _ := ioutil.ReadFile("rfc-index.xml") // https://www.rfc-editor.org/rfc-index.xml
	xml.Unmarshal(d, &idx)

	document_map := make(map[string]*ietf_object)
	author_map := make(map[string]*ietf_object)

	for _, doc := range idx.RFCs {
		doc_obj := ietf_object{id: document_count, name: doc.Name}
		document_count++
		documents = append(documents, doc_obj)
		document_map[doc_obj.name] = &doc_obj

		for _, author := range doc.Authors {
			if _, ok := author_map[author]; !ok {
				auth_obj := ietf_object{id: author_count, name: author, dist: math.MaxUint64}
				author_count++
				authors = append(authors, auth_obj)
				author_map[author] = &auth_obj
			}

			doc_obj.links = append(doc_obj.links, author_map[author])
			author_map[author].links = append(author_map[author].links, &doc_obj)
		}
	}

	fmt.Printf("Ingested file rfc-index.xml\nProcessed %d RFCs\nProcessed %d unique authors\n", document_count, author_count)

	_, ok := author_map[os.Args[1]]
	if !ok {
		fmt.Printf("%s not found\n", os.Args[1])
		os.Exit(1)
	}

	wg.Add(1)

	go do_work(&wg, 0, author_map["J. Postel"], []*ietf_object{})

	wg.Wait()

	obj := author_map[os.Args[1]]
	fmt.Printf("Name: %s -- Postel #: %d (%d)\n", obj.name, obj.dist, len(obj.paths))
	for _, path := range obj.paths {
		for _, path_obj := range path {
			fmt.Printf("%s -> ", path_obj.name)
		}
		fmt.Printf("%s\n", obj.name)
	}
}
