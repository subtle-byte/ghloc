package repository

import (
	"archive/zip"
	"bytes"
	"fmt"
	"ghloc/internal/model"
	"ghloc/internal/service"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Github struct {
}

const maxZipSize = 100 * 1024 * 1024 // 100 MiB

func BuildGithubUrl(user, repo, branch string) string {
	return fmt.Sprintf("https://github.com/%v/%v/archive/refs/heads/%v.zip", user, repo, branch)
}

func ReadIntoMemory(r io.Reader) (*bytes.Reader, error) {
	buf := &bytes.Buffer{}

	lr := &LimitedReader{Reader: r, Remaining: maxZipSize}
	_, err := io.Copy(buf, lr)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func (r Github) GetContent(user, repo, branch string, tempStorage service.TempStorage) (*model.Content, error) {
	url := BuildGithubUrl(user, repo, branch)

	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		log.Println(url, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, model.NotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%v %v %v", url, "unexpected status code", resp.StatusCode)
	}

	contents := &model.Content{}

	readerAt := io.ReaderAt(nil)
	len := 0
	if tempStorage == service.File {
		tempFile, err := NewTempFile(resp.Body)
		if err != nil {
			return nil, err
		}
		contents.Closer = tempFile.Close
		readerAt = tempFile
		len = tempFile.Len()
	} else {
		r, err := ReadIntoMemory(resp.Body)
		if err != nil {
			return nil, err
		}
		readerAt = r
		len = r.Len()
		log.Print("github.GetContent: use memory for temp data")
	}

	logIOBlocking("github.GetContent", start, fmt.Sprintf("%v %.2fMiB", url, float64(len)/1024.0/1024.0))

	reader, err := zip.NewReader(readerAt, int64(len))
	if err != nil {
		return nil, err
	}

	for _, file := range reader.File {
		contents.ByPath = append(contents.ByPath, model.ContentForPath{
			Path:          file.Name[strings.Index(file.Name, "/")+1:],
			ContentOpener: file.Open,
		})
	}
	return contents, nil
}
