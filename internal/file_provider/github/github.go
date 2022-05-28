package github

import (
	"archive/zip"
	"bytes"
	"fmt"
	"ghloc/internal/file_provider"
	"ghloc/internal/github_service"
	"ghloc/internal/rest"
	"ghloc/internal/util"
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

func (r Github) GetContent(user, repo, branch string, tempStorage github_service.TempStorage) (_ []file_provider.FileForPath, close func() error, _ error) {
	url := BuildGithubUrl(user, repo, branch)

	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		log.Println(url, err)
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, rest.NotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("%v %v %v", url, "unexpected status code", resp.StatusCode)
	}

	filesForPaths := []file_provider.FileForPath(nil)
	closer := func() error { return nil }

	readerAt := io.ReaderAt(nil)
	len := 0
	if tempStorage == github_service.TempStorageFile {
		tempFile, err := NewTempFile(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		closer = tempFile.Close
		readerAt = tempFile
		len = tempFile.Len()
	} else {
		r, err := ReadIntoMemory(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		readerAt = r
		len = r.Len()
		log.Print("github.GetContent: use memory for temp data")
	}

	util.LogIOBlocking("github.GetContent", start, fmt.Sprintf("%v %.2fMiB", url, float64(len)/1024.0/1024.0))

	zipReader, err := zip.NewReader(readerAt, int64(len))
	if err != nil {
		closer()
		return nil, nil, err
	}

	for _, file := range zipReader.File {
		filesForPaths = append(filesForPaths, file_provider.FileForPath{
			Path:   file.Name[strings.Index(file.Name, "/")+1:],
			Opener: file.Open,
		})
	}
	return filesForPaths, closer, nil
}
