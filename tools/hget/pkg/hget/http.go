package hget

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"k8s.io/klog/v2"
)

// DownloadResult is the result of a download operation.
// It contains the hash of the downloaded file and the length of the file.
type DownloadResult struct {
	Hash   string
	Length int64
}

// DownloadURL downloads the file from the given URL and writes it to the given writer.
// It returns the hash of the downloaded file and the length of the file.
func DownloadURL(ctx context.Context, url string, w io.Writer) (*DownloadResult, error) {
	log := klog.FromContext(ctx)

	// Set up the HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	// Enable compression, it saves everyone time and money
	req.Header.Add("Accept-Encoding", "gzip, deflate")

	info := fmt.Sprintf("%s %s", req.Method, req.URL.String())

	startTime := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing http request %s: %w", info, err)
	}
	defer resp.Body.Close()

	var src io.Reader

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		src, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("creating gzip reader: %w", err)
		}
	case "deflate":
		src = flate.NewReader(resp.Body)

	case "":
		// No compression
		src = resp.Body

	default:
		return nil, fmt.Errorf("unsupported content encoding: %s", resp.Header.Get("Content-Encoding"))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from http request %s: %d", info, resp.StatusCode)
	}

	// Calculate hash while downloading
	hasher := sha256.New()
	writer := io.MultiWriter(w, hasher)

	if _, err := io.Copy(writer, src); err != nil {
		return nil, fmt.Errorf("downloading from %s: %w", info, err)
	}

	elapsed := time.Since(startTime)

	result := &DownloadResult{
		Hash:   hex.EncodeToString(hasher.Sum(nil)),
		Length: resp.ContentLength,
	}
	log.Info("downloaded file", "url", url, "result", result, "elapsed", elapsed, "content-encoding", resp.Header.Get("Content-Encoding"))

	return result, nil
}

func readLines(ctx context.Context, url string, callback func(line string) error) (*DownloadResult, error) {
	// Set up the HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	info := fmt.Sprintf("%s %s", req.Method, req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing http request %s: %w", info, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from http request %s: %d", info, resp.StatusCode)
	}

	// Calculate hash while downloading
	hasher := sha256.New()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) != 0 {
			if _, err := hasher.Write(line); err != nil {
				return nil, fmt.Errorf("writing to hasher: %w", err)
			}
			s := string(line[:len(line)-1])
			if err := callback(s); err != nil {
				return nil, fmt.Errorf("callback failed: %w", err)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("reading from %s: %w", info, err)
		}
	}

	result := &DownloadResult{
		Hash:   hex.EncodeToString(hasher.Sum(nil)),
		Length: resp.ContentLength,
	}

	return result, nil
}
