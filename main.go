package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/tealeg/xlsx"
)

// Doc struct is a generic document
type Doc struct {
	DocNr    string
	Title    string
	Path     string
	Revision string
	Markdown bool
	Files    []string
}

func overviewHandler(w http.ResponseWriter, r *http.Request) {
	p := docMap
	t := template.Must(template.ParseFiles("templates/base.html", "templates/overview.html"))
	t.Execute(w, p)
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	p := docMap
	t := template.Must(template.ParseFiles("templates/base.html", "templates/details.html"))
	t.Execute(w, p)
}

func otherHandler(w http.ResponseWriter, r *http.Request) {
	p := docMap
	t := template.Must(template.ParseFiles("templates/base.html", "templates/other.html"))
	t.Execute(w, p)
}

func searchForDocs(rootDir string) {
	subDirToSkip := ".git" // dir/to/walk/skip

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", rootDir, err)
			return err
		}
		if info.IsDir() && info.Name() == subDirToSkip {
			// fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		if !info.IsDir() {
			//docList = append(docList, Doc{Path: path, Title: info.Name(), Markdown: strings.HasSuffix(info.Name(), ".md")})
			if strings.HasPrefix(info.Name(), projectNumber) {
				for k, v := range docMap {
					if strings.HasPrefix(info.Name(), k) {
						docMap[k].Files = append(v.Files, path)
						// fmt.Println(v.Files)
					}
				}
			}
			// fmt.Printf("Add file to list: %q\n", path)
			return nil
		}

		// fmt.Printf("visited dir: %q\n", path)
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", rootDir, err)
	}
}

func initDocMap(nrLogFile string) {
	// Open the number log.
	xlFile, err := xlsx.OpenFile(nrLogFile)
	if err != nil {
		log.Fatal(err)
	}

	// Choose the first sheet
	sheet := xlFile.Sheets[0]
	// Get project number from a certain cell
	projectNumber = sheet.Cell(0, 1).String()
	// Get project name from a certain cell
	projectName = sheet.Cell(1, 1).String()

	// Initialize the map that will hold all the docs in the project
	docMap = make(map[string]*Doc)

	for _, row := range sheet.Rows[4:] {
		docnr, err := row.Cells[2].FormattedValue()
		if err != nil {
			log.Fatal(err)
		}
		if docnr != "" {
			// nr := row.Cells[0].String()
			title := row.Cells[1].String()

			// Initialize the project document map with the numbers from the number log.
			docMap[docnr] = &Doc{Title: title, DocNr: docnr}
			// fmt.Printf("%s : %s : %s\n", nr, docnr, title)
		}
	}
}

// Global variables
var docMap map[string]*Doc
var projectNumber string
var projectName string

func main() {
	rootDir := "."
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatal(err)
	}
	initDocMap("testfiles/Nummerliggare.xlsm")
	searchForDocs(rootDir)

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/", overviewHandler)
	r.HandleFunc("/details", detailsHandler)
	r.HandleFunc("/other", otherHandler)
	http.Handle("/", r)

	addr := "127.0.0.1:8000"
	srv := &http.Server{
		Handler: r,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Prjbrain is monitoring " + absRootDir)
	fmt.Println("Please go to " + addr + " in your favourite web browser")
	log.Fatal(srv.ListenAndServe())
}
