package github_files_provider

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/subtle-byte/ghloc/internal/server/rest"
	"github.com/subtle-byte/ghloc/internal/service/github_stat"
	"github.com/subtle-byte/ghloc/internal/util"
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

func (r Github) GetContent(user, repo, branch string, tempStorage github_stat.TempStorage) (_ []github_stat.FileForPath, close func() error, _ error) {
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

	closer := func() error { return nil }

	readerAt := io.ReaderAt(nil)
	readerLen := 0
	if tempStorage == github_stat.TempStorageFile {
		tempFile, err := NewTempFile(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		closer = tempFile.Close
		readerAt = tempFile
		readerLen = tempFile.Len()
	} else {
		r, err := ReadIntoMemory(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		readerAt = r
		readerLen = r.Len()
		log.Print("github.GetContent: use memory for temp data")
	}

	util.LogIOBlocking("github.GetContent", start, fmt.Sprintf("%v %.2fMiB", url, float64(readerLen)/1024.0/1024.0))

	zipReader, err := zip.NewReader(readerAt, int64(readerLen))
	if err != nil {
		closer()
		return nil, nil, err
	}

	filesForPaths := make([]github_stat.FileForPath, 0, len(zipReader.File))
	for _, file := range zipReader.File {
		filesForPaths = append(filesForPaths, github_stat.FileForPath{
			Path:   file.Name[strings.Index(file.Name, "/")+1:],
			Opener: file.Open,
		})
	}
	return filesForPaths, closer, nil
}
