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

	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
)

// Doc struct is a generic document
type Doc struct {
	DocNr    string
	Title    string
	Path     string
	Revision string
	Markdown bool
	Files    []File
}

// File struct
type File struct {
	FilePath string
	RelPath  string
}

// Global variables
var docMap map[string]*Doc
var projectNumber string
var projectName string
var tmpl packr.Box
var static packr.Box

func main() {
	if len(os.Args) > 1 {
		viper.SetConfigFile(os.Args[1]) // Prefer to use config file provided as argument
	} else {
		viper.SetConfigName("config") // name of config file (without extension)
		viper.AddConfigPath(".")      // optionally look for config in the working directory
	}

	// Add default values to options if not set in config file
	viper.SetDefault("pn_start_row", 4)
	viper.SetDefault("title_col", 1)
	viper.SetDefault("docnr_col", 2)
	viper.SetDefault("number_log", "testfiles/Nummerliggare.xlsm")

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
	t.Execute(w, docMap)
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("base").Parse(tmpl.String("base.html")))
	_, err := t.Parse(tmpl.String("details.html"))
	if err != nil {
		log.Fatalf("Cant parse the template %v", err)
	}
	t.Execute(w, docMap)
}

func otherHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("base").Parse(tmpl.String("base.html")))
	_, err := t.Parse(tmpl.String("other.html"))
	if err != nil {
		log.Fatalf("Cant parse the template %v", err)
	}
	t.Execute(w, docMap)
}

// Files handler opens a local file and then returns to the overview page.
func filesHandler(w http.ResponseWriter, r *http.Request) {
	go open(r.URL.Query().Get("path"))
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
						relPath, err := filepath.Rel(rootDir, path)
						if err != nil {
							log.Fatalf("Cant find current working directory %v", err)
						}
						docMap[k].Files = append(v.Files, File{FilePath: path, RelPath: relPath})
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

	for _, row := range sheet.Rows[viper.GetInt("pn_start_row"):] {
		docnr, err := row.Cells[viper.GetInt("docnr_col")].FormattedValue()
		if err != nil {
			log.Fatal(err)
		}
		if docnr != "" {
			// nr := row.Cells[0].String()
			title := row.Cells[viper.GetInt("title_col")].String()

			// Initialize the project document map with the numbers from the number log.
			docMap[docnr] = &Doc{Title: title, DocNr: docnr}
			// fmt.Printf("%s : %s : %s\n", nr, docnr, title)
		}
	}
}
