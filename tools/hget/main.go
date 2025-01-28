package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/tools/hget/pkg/hget"
)

type options struct {
	Sha256     string
	Chmod      os.FileMode
	OutputPath string

	Indexes []string
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// // StringSliceVar is like flag.StringSliceVar, but it allows empty strings to be added
// func StringSliceVar(p *[]string, name string, value []string, usage string) {
// 	flag.Var(stringSliceValue(p), name, usage)
// }

type stringSliceValue []string

func (s *stringSliceValue) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceValue) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func run(ctx context.Context) error {
	log := klog.FromContext(ctx)

	opts := &options{}

	fileMode := ""
	flag.StringVar(&fileMode, "chmod", fileMode, "Permissions to set on the output file (octal)")
	flag.StringVar(&opts.Sha256, "sha256", "", "SHA256 hash to verify against")
	flag.StringVar(&opts.OutputPath, "output", "", "Path to write the downloaded file")

	var sha256sums stringSliceValue
	flag.Var(&sha256sums, "sha256sums", "URL to SHA256SUMS file to find the file to download")

	var urls stringSliceValue
	flag.Var(&urls, "url", "URL to download the file from")

	flag.Parse()

	opts.Indexes = sha256sums
	if fileMode != "" {
		parsed, err := strconv.ParseUint(fileMode, 8, 32)
		if err != nil {
			return fmt.Errorf("parsing chmod %q: %v", fileMode, err)
		}
		opts.Chmod = os.FileMode(parsed)
	}

	if opts.Sha256 == "" {
		fmt.Fprintln(os.Stderr, "error: --sha256 is required")
		flag.Usage()
		os.Exit(1)
	}

	if opts.OutputPath == "" {
		fmt.Fprintln(os.Stderr, "error: --output is required")
		flag.Usage()
		os.Exit(1)
	}

	alreadyExists, err := checkExistingFile(ctx, opts.OutputPath, opts)
	if err != nil {
		log.Error(err, "failed to check existing file", "path", opts.OutputPath)
	}
	if alreadyExists {
		log.Info("file already exists", "path", opts.OutputPath)
		return nil
	}

	var errs []error
	for _, url := range urls {
		if err := downloadToFile(ctx, opts, url); err != nil {
			errs = append(errs, err)
		} else {
			return nil
		}
	}

	// Try assets from the index
	{
		assets := hget.NewIndex()

		for _, index := range opts.Indexes {
			if err := assets.AddToIndex(ctx, index); err != nil {
				errs = append(errs, fmt.Errorf("adding %q to asset index: %w", index, err))
			}
		}

		asset, err := assets.Lookup(ctx, opts.Sha256)
		if err != nil {
			errs = append(errs, fmt.Errorf("looking up asset: %w", err))
		}
		if asset != nil {
			if err := downloadToFile(ctx, opts, asset.URL); err != nil {
				errs = append(errs, fmt.Errorf("downloading asset: %w", err))
			} else {
				return nil
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to download file: %w", errors.Join(errs...))
	}

	return nil
}

func downloadToFile(ctx context.Context, opts *options, url string) error {
	log := klog.FromContext(ctx)

	startTime := time.Now()

	dir := filepath.Dir(opts.OutputPath)
	// Create a temporary file to download to
	tmpFile, err := os.CreateTemp(dir, "hget-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %v", err)
	}
	removeTempFile := true
	defer func() {
		if removeTempFile {
			if err := os.Remove(tmpFile.Name()); err != nil {
				log.Error(err, "failed to remove temp file", "path", tmpFile.Name())
			}
		}
	}()

	downloadResults, err := hget.DownloadURL(ctx, url, tmpFile)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Info("downloaded file", "url", url, "results", downloadResults, "elapsed", elapsed)

	// Verify hash
	if downloadResults.Hash != opts.Sha256 {
		return fmt.Errorf("hash mismatch: got %s, want %s", downloadResults.Hash, opts.Sha256)
	}

	// Set permissions if specified
	if opts.Chmod != 0 {
		if err := os.Chmod(tmpFile.Name(), opts.Chmod); err != nil {
			return fmt.Errorf("setting permissions: %w", err)
		}
	}

	// Move to final destination
	if err := os.Rename(tmpFile.Name(), opts.OutputPath); err != nil {
		return fmt.Errorf("moving temp file to destination: %w", err)
	}
	removeTempFile = false
	return nil
}

func checkExistingFile(ctx context.Context, p string, opts *options) (bool, error) {
	log := klog.FromContext(ctx)

	stat, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat file %q: %w", p, err)
	}

	hash, err := hget.GetHashForFile(p)
	if err != nil {
		return false, fmt.Errorf("failed to get hash of %q: %w", p, err)
	}
	if hash != opts.Sha256 {
		log.Info("file already exists but hash is not correct", "path", p, "got", hash, "want", opts.Sha256)
		return false, nil
	}

	if opts.Chmod != 0 && stat.Mode() != opts.Chmod {
		log.Info("file already exists but permissions are not correct", "path", p, "got", stat.Mode(), "want", opts.Chmod)
		if err := os.Chmod(p, opts.Chmod); err != nil {
			return false, fmt.Errorf("setting permissions on %q: %w", p, err)
		}
	}

	return true, nil
}
