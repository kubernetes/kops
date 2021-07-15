package gcs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func ListPrefixes(bucketName string, prefix string) ([]string, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	bucket := client.Bucket(bucketName)

	var prefixes []string

	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix, Delimiter: "/"})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing bucket: %w", err)
		}
		if attrs.Prefix != "" {
			prefixes = append(prefixes, attrs.Prefix)
		}
	}

	return prefixes, nil
}

func ReadObject(bucketName string, name string) ([]byte, error) {
	ctx := context.Background()

	cacheDir := os.Getenv("CACHE_DIR")
	if cacheDir == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("os.UserHomeDir failed: %w", err)
		}
		cacheDir = filepath.Join(homedir, ".cache", "testsoup", "gcs")
	}

	cacheName := bucketName + "_" + name
	cacheName = strings.ReplaceAll(cacheName, "/", "_")
	cachePath := filepath.Join(cacheDir, cacheName)

	existing, err := os.ReadFile(cachePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("os.ReadFile(%q) failed: %w", cachePath, err)
		}
	} else {
		return existing, nil
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("MkdirAll(%q) failed: %w", cacheDir, err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	p := "gs://" + bucketName + "/" + name

	bucket := client.Bucket(bucketName)

	r, err := bucket.Object(name).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("error opening object %q: %w", p, err)
	}
	defer r.Close()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading object %q: %w", p, err)
	}

	if err := os.WriteFile(cachePath, body, 0644); err != nil {
		return nil, fmt.Errorf("os.WriteFile(%q) failed: %w", cachePath, err)
	}

	return body, nil
}
