# Prjbrain

The idea behind Prjbrain is that I by looking through your product development documentation provide valuable insight into the status of the project.

## Current functionality

0. It reads a `config.yml` located at the root of the project directory.
1. It reads document numbers and titles from an .XLSX file provided in the config.yml.
2. It recursively goes through all the files in the directory and sub directories and try to match the files to the document number.
3. It fires up a web server that serves the Prjbrain dashboard at http://localhost:8000
4. From the dashboard you can see all documents and all files attached to that document.

## Planned functionality

* Check for common naming errors, where documents don't follow the projects preffered conventions.
* Do sanity tests of the files and visually show in the dashboard when something looks wrong.
* Load the summary from .docx report into the dashboard for easy visualization of the entire project.
* The ability to parse the `project requirement specifications` and cross reference those to the reports to see where the requirements are fulfilled.
* Watch the directory in realtime for changes and automatically reload files.
* Ability to export a zip file of all the latest release files.

## Rationale to technical decicisons

Prjbrain is implemented in [go](https://www.golang.org). Go is well suited for the job.

* It has an large standard library and is well suited for building web server applications.
* It has a large eco system of third part libraries.
* It compiles down to a static binary that is easy to distribute.
* It is trivial to cross-compile for different OS and architechtures.
* It is a fun and easy-to-learn language.
* It is FAST!

## To build

Build using the current go way

``` bash
go get -u github.com/gntech/prjbrain

cd $GOPATH/src/github.com/gntech/prjbrain

go build
```

Or use [vgo](https://github.com/golang/vgo) even though it is not standardized yet.

``` bash
vgo build
```

## To build snapshot release

Prefferably use [goreleaser](https://goreleaser.com/). This will build 64bit binaries for Linux and Windows.

`goreleaser --snapshot --rm-dist`

See `.goreleaser.yml` for options.

## Usage

1. Put the prjbrain/prjbrain.exe binary together with the config.yml at the root of your project folder.
2. Edit the config.yml to point to your number log XLSX/XLSM file.
    * Possibly also adjust the cells and rows used for the different values.
3. Run the prjbrain/prjbrain.exe binary. The dashboard will automatically open your web browser.

## Licenses 

Due to the gooxml dependency this project currently uses [AGPL-3.0](https://opensource.org/licenses/AGPL-3.0)
