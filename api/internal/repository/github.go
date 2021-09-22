package repository

import (
	"archive/zip"
	"fmt"
	"ghloc/internal/model"
	"ghloc/internal/service"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Github struct {
}

const maxZipSize = 100 * 1024 * 1024 // 100 MiB

func BuildGithubUrl(user, repo, branch string) string {
	return fmt.Sprintf("https://github.com/%v/%v/archive/refs/heads/%v.zip", user, repo, branch)
}

func (r Github) GetContent(user, repo, branch string) ([]service.ContentForPath, error) {
	url := BuildGithubUrl(user, repo, branch)

	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		log.Println(url, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, model.BadRequest{"Repository is not found"}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%v %v %v", url, "unexpected status code", resp.StatusCode)
	}

	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	log.Print("temp file: ", tmpfile.Name())
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	lr := &LimitedReader{Reader: resp.Body, Remaining: maxZipSize}
	_, err = io.Copy(tmpfile, lr)
	if err != nil {
		return nil, err
	}

	zipSize := maxZipSize - lr.Remaining
	logIOBlocking("github.GetContent", start, fmt.Sprintf("%v %.2fMiB", url, float64(zipSize)/1024.0/1024.0))

	reader, err := zip.OpenReader(tmpfile.Name())
	if err != nil {
		return nil, err
	}

	contents := []service.ContentForPath(nil)
	for _, file := range reader.File {
		contents = append(contents, service.ContentForPath{
			Path:          file.Name[strings.Index(file.Name, "/")+1:],
			ContentOpener: file.Open,
		})
	}
	return contents, nil
}
