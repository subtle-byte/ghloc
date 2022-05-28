package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"

	"github.com/pkg/browser"
	"github.com/subtle-byte/ghloc/internal/file_provider/files_in_dir"
	"github.com/subtle-byte/ghloc/internal/rest"
	"github.com/subtle-byte/ghloc/internal/stat"
)

//go:embed server_static
var serverStatic embed.FS

func main() {
	files, err := files_in_dir.GetFilesInDir(".")
	if err != nil {
		panic(err)
	}
	locCounter := stat.NewLOCCounter()
	for _, file := range files {
		func() {
			fileReader, err := file.Opener()
			if err != nil {
				panic(err)
			}
			defer fileReader.Close()
			err = locCounter.AddFile(file.Path, fileReader)
			if err != nil {
				panic(err)
			}
		}()
	}
	locsForPaths := locCounter.GetLOCsForPaths()

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		var filter *string
		if filters := r.Form["filter"]; len(filters) >= 1 {
			filter = &filters[0]
		}
		var matcher *string
		if matchers := r.Form["match"]; len(matchers) >= 1 {
			matcher = &matchers[0]
		}
		statTree := stat.BuildStatTree(locsForPaths, filter, matcher)
		rest.WriteResponse(w, (*rest.Stat)(statTree))
	})

	serverStatic, err := fs.Sub(serverStatic, "server_static")
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(http.FS(serverStatic)))

	socket, err := net.Listen("tcp", "localhost:0") // :0 means random free port
	if err != nil {
		panic(err)
	}
	url := fmt.Sprintf("http://%v", socket.Addr())
	fmt.Println("Web UI:", url)
	fmt.Println("API:   ", url+"/api")
	browser.OpenURL(url)
	http.Serve(socket, nil)
}
