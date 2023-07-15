package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/subtle-byte/ghloc/internal/infrastructure/local_files_provider"
	"github.com/subtle-byte/ghloc/internal/server/rest"
	"github.com/subtle-byte/ghloc/internal/service/loc_count"
)

//go:embed server_static
var serverStatic embed.FS

func main() {
	var matcher *string
	flag.Func("m", "sets `match expression` used to filter results, sets flag -c implicitly", func(s string) error {
		matcher = &s
		return nil
	})
	cli := flag.Bool("c", false, "output to` `console")
	flag.Parse()

	locsForPaths := countLOCsForPaths()
	if matcher != nil {
		printInConsole(locsForPaths, *matcher)
	} else if *cli {
		printInConsole(locsForPaths, "")
	} else {
		runServer(locsForPaths)
	}
}

func countLOCsForPaths() []loc_count.LOCForPath {
	fmt.Print("Counting lines of code...")
	counted := make(chan bool, 1)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-counted:
				return
			case <-ticker.C:
				fmt.Print(".")
			}
		}
	}()

	files, err := local_files_provider.GetFilesInDir(".")
	if err != nil {
		panic(err)
	}
	locCounter := loc_count.NewFilesLOCCounter()
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

	counted <- true
	fmt.Println()

	return locsForPaths
}

func printInConsole(locsForPaths []loc_count.LOCForPath, matcher string) {
	statTree := loc_count.BuildStatTree(locsForPaths, nil, &matcher)
	type LocAndLang struct {
		Loc  int
		Lang string
	}
	stats := []LocAndLang{}
	for lang, loc := range statTree.LOCByLangs {
		stats = append(stats, LocAndLang{loc, lang})
	}
	sort.Slice(stats, func(i, j int) bool { return stats[i].Loc > stats[j].Loc })
	firstWidth := 50
	secondWidth := 20
	width := firstWidth + secondWidth + 3
	printPair := func(a, b interface{}, sep string) {
		aStr := fmt.Sprint(a)
		if len(aStr) > firstWidth {
			aStr = aStr[:firstWidth-3] + "..."
		}
		bStr := fmt.Sprint(b)
		sepLen := (firstWidth - len(aStr)) + (secondWidth - len(bStr))
		fmt.Printf(" %v%v%v \n", aStr, strings.Repeat(sep, sepLen+1), bStr)
	}
	fmt.Println(strings.Repeat("=", width))
	printPair("File type", "Lines of code", " ")
	fmt.Println(strings.Repeat("=", width))
	for _, stat := range stats {
		printPair(stat.Lang, stat.Loc, ".")
	}
	fmt.Println(strings.Repeat("=", width))
	printPair("Total", statTree.LOC, " ")
	fmt.Println(strings.Repeat("=", width))
}

func runServer(locsForPaths []loc_count.LOCForPath) {
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
		statTree := loc_count.BuildStatTree(locsForPaths, filter, matcher)
		rest.WriteResponse(w, r, (*rest.SortedStat)(statTree), true)
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
	go browser.OpenURL(url)
	http.Serve(socket, nil)
}
