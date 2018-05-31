package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/gohugoio/hugo/parser"
	"github.com/icza/gowut/gwu"
	"github.com/russross/blackfriday"
)

func buildHTML(doc string) gwu.Comp {
	r, err := os.Open(doc)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	page, err := parser.ReadFrom(r)
	if err != nil {
		log.Fatal(err)
	}

	metadata, err := page.Metadata()
	if err != nil {
		log.Fatal(err)
	}

	p := gwu.NewPanel()
	p.Add(gwu.NewTextBox(metadata["title"].(string)))
	p.Add(gwu.NewTextBox(metadata["date"].(string)))
	output := blackfriday.Run(page.Content())
	html := string(output[:])
	p.Add(gwu.NewHTML(html))

	return p
}

func main() {
	// Create and start a GUI server (omitting error check)
	server := gwu.NewServer("prjbrain", "localhost:8081")

	docsDir := "docs/"
	docs, err := ioutil.ReadDir(docsDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, doc := range docs {
		fmt.Println(doc.Name())
		win := gwu.NewWindow(doc.Name(), doc.Name())
		win.Add(buildHTML(path.Join(docsDir, doc.Name())))
		server.AddWin(win)
	}

	server.Start("") // Also opens windows list in browser
}
