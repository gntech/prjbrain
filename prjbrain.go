/*
Prjbrain - dashboard for product development folders.
Copyright (C) 2018  Gustav Näslund

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"baliance.com/gooxml/spreadsheet"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// Doc struct is a generic document
type Doc struct {
	DocNr    string
	Title    string
	Path     string
	Revision string
	Files    []File
}

// File struct
type File struct {
	FilePath string
	RelPath  string
	Rev      string
	DocNr    string
	Comment  string
}

// Data struct contains all the data sent to the pages in the web interface
type Data struct {
	ProjectName   *string
	ProjectNumber *string
	DocMap        *map[string]*Doc
	OrphanFiles   *[]File
}

// Global variables
var docMap map[string]*Doc
var orphanFiles []File
var projectNumber string
var projectName string
var tmpl packr.Box
var static packr.Box
var data Data

func main() {
	if len(os.Args) > 1 {
		viper.SetConfigFile(os.Args[1]) // Prefer to use config file provided as argument
	} else {
		viper.SetConfigName("config") // name of config file (without extension)
		viper.AddConfigPath(".")      // optionally look for config in the working directory
	}

	// Add default values to options if not set in config file
	viper.SetDefault("prjtitle", "")
	viper.SetDefault("prjnr", "")
	viper.SetDefault("prjnr_cell", "C1")
	viper.SetDefault("prjtitle_cell", "C2")
	viper.SetDefault("pn_start_row", 5)
	viper.SetDefault("title_col", "B")
	viper.SetDefault("docnr_col", "C")
	viper.SetDefault("number_log", "testfiles/Nummerliggare.xlsm")
	viper.SetDefault("subdirs_to_skip", []string{".git"})

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatalf("fatal error config file: %s", err)
	}

	rootDir := getInputDir(viper.ConfigFileUsed())
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	initDocMap(viper.GetString("number_log"))
	searchForDocs(rootDir)

	data = Data{ProjectName: &projectName, ProjectNumber: &projectNumber, DocMap: &docMap, OrphanFiles: &orphanFiles}

	tmpl = packr.NewBox("./templates")
	static = packr.NewBox("./static")

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(static)))
	r.HandleFunc("/", overviewHandler)
	r.HandleFunc("/details", detailsHandler)
	r.HandleFunc("/other", otherHandler)
	r.HandleFunc("/files", filesHandler)
	http.Handle("/", r)

	addr := "localhost:8000"
	srv := &http.Server{
		Handler: r,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Prjbrain is monitoring " + absRootDir + " and serving at " + addr)
	fmt.Println("Press Ctrl+C to quit")
	go open("http://" + addr)
	log.Fatal(srv.ListenAndServe())
}

// open opens the specified URL or file from the file system in the systems default application.
// From github.com/icza/gowut
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func overviewHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("base").Parse(tmpl.String("base.html")))
	_, err := t.Parse(tmpl.String("overview.html"))
	if err != nil {
		log.Fatalf("Cant parse the template %v", err)
	}
	t.Execute(w, data)
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("base").Parse(tmpl.String("base.html")))
	_, err := t.Parse(tmpl.String("details.html"))
	if err != nil {
		log.Fatalf("Cant parse the template %v", err)
	}
	t.Execute(w, data)
}

func otherHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("base").Parse(tmpl.String("base.html")))
	_, err := t.Parse(tmpl.String("other.html"))
	if err != nil {
		log.Fatalf("Cant parse the template %v", err)
	}
	t.Execute(w, data)
}

// Files handler opens a local file and then returns to the overview page.
func filesHandler(w http.ResponseWriter, r *http.Request) {
	go open(r.URL.Query().Get("path"))
	// Use redirect code 303 to avoid that the browser caches away the actual GET request when opening a file.
	http.Redirect(w, r, "/", 303)
}

func getInputDir(cf string) string {
	p := path.Dir(cf)
	if path.IsAbs(p) {
		return p
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cant find current working directory %v", err)
	}
	return path.Join(wd, p)
}

// parseFile tries to create a formal File from the given path.
// If it encounters anything strange it will add that to a comment.
func parseFile(path string, rootDir string) {
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		log.Fatalf("%s is not in %s %v", path, rootDir, err)
	}
	name := filepath.Base(path)

	// Check if the filename is in the docMap retrieved from the number log file
	for k, v := range docMap {
		if strings.HasPrefix(name, k) {
			docMap[k].Files = append(v.Files, File{FilePath: path, RelPath: relPath})
			return
		}
		// If not found try to convert filenames to all lower case when comparing to work around how filenames work in Windows.
		// If a match is found now, add a comment to fix file naming.
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(k)) {
			docMap[k].Files = append(v.Files, File{FilePath: path, RelPath: relPath, Comment: "The case of the filename and docnr does not match."})
			return
		}
	}

	// Check if the filename starts with the project number, indicating that this is a project file.
	if strings.HasPrefix(name, projectNumber) {
		// This is considered an orphan file and added to the orphan file list.
		orphanFiles = append(orphanFiles, File{FilePath: path, RelPath: relPath})
	}

	return
}

func searchForDocs(rootDir string) {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", rootDir, err)
			return err
		}
		if info.IsDir() {
			for _, subDir := range viper.GetStringSlice("subdirs_to_skip") {
				if info.Name() == subDir {
					return filepath.SkipDir
				}
			}
			// fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
		}
		if !info.IsDir() {
			// Try to parse the file/filename to add it to the project files.
			parseFile(path, rootDir)
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
	xlFile, err := spreadsheet.Open(nrLogFile)
	if err != nil {
		log.Fatal(err)
	}

	// Choose the first sheet
	sheet := xlFile.Sheets()[0]

	// Get project number from a certain cell if number is not set explicitly in the config
	if viper.GetString("prjnr") == "" {
		projectNumber = sheet.Cell(viper.GetString("prjnr_cell")).GetFormattedValue()
	} else {
		projectNumber = viper.GetString("prjnr")
	}

	// Get project name from a certain cell if name is not set explicitly in the config
	if viper.GetString("prjname") == "" {
		projectName = sheet.Cell(viper.GetString("prjtitle_cell")).GetFormattedValue()
	} else {
		projectName = viper.GetString("prjname")
	}

	// Initialize the map that will hold all the docs in the project
	docMap = make(map[string]*Doc)

	for _, row := range sheet.Rows()[viper.GetInt("pn_start_row")-2:] {
		docnr := row.Cell(viper.GetString("docnr_col")).GetFormattedValue()
		if err != nil {
			log.Fatal(err)
		}
		if docnr != "" {
			// nr := row.Cells[0].String()
			title := row.Cell(viper.GetString("title_col")).GetFormattedValue()

			// Initialize the project document map with the numbers from the number log.
			docMap[docnr] = &Doc{Title: title, DocNr: docnr}
			// fmt.Printf("%s : %s : %s\n", nr, docnr, title)
		}
	}
}
