package github_files_provider

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/subtle-byte/ghloc/internal/server/rest"
	"github.com/subtle-byte/ghloc/internal/service/github_stat"
)

type Github struct {
	maxZipSizeBytes int
}

func New(maxZipSizeMB int) *Github {
	return &Github{
		maxZipSizeBytes: maxZipSizeMB * 1024 * 1024,
	}
}

func buildGithubUrl(user, repo, branch string) string {
	return fmt.Sprintf("https://github.com/%v/%v/archive/refs/heads/%v.zip", user, repo, branch)
}

func (g *Github) readIntoMemory(r io.Reader) (*bytes.Reader, error) {
	buf := &bytes.Buffer{}

	lr := &LimitedReader{Reader: r, Remaining: g.maxZipSizeBytes}
	_, err := io.Copy(buf, lr)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func (g *Github) GetContent(ctx context.Context, user, repo, branch string, tempStorage github_stat.TempStorage) (_ []github_stat.FileForPath, close func() error, _ error) {
	url := buildGithubUrl(user, repo, branch)

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("do request: %w", err)
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
		tempFile, err := NewTempFile(resp.Body, g.maxZipSizeBytes)
		if err != nil {
			return nil, nil, err
		}
		zerolog.Ctx(ctx).Info().
			Str("fileName", tempFile.File.Name()).
			Msgf("Using disk file to temporary store repo archive")
		closer = tempFile.Close
		readerAt = tempFile
		readerLen = tempFile.Len()
	} else {
		r, err := g.readIntoMemory(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		readerAt = r
		readerLen = r.Len()
		zerolog.Ctx(ctx).Info().Msgf("Using RAM to temporary store repo archive")
	}

	zerolog.Ctx(ctx).Info().
		Float64("durationSec", time.Since(start).Seconds()).
		Str("url", url).
		Int("sizeBytes", readerLen).
		Msg("Downloaded repo zip archive")

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
