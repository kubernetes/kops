/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

func main() {
	err := run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	filestoreBase := ""
	prefix := ""
	hashFileURL := ""
	var exclude globList

	flag.StringVar(&filestoreBase, "base", filestoreBase, "base directory")
	flag.StringVar(&prefix, "prefix", prefix, "prefix to fetch")
	flag.StringVar(&hashFileURL, "sums", hashFileURL, "prefix to fetch")
	flag.Var(&exclude, "exclude", "path-globs to exclude from output")

	flag.Parse()

	if hashFileURL == "" {
		hashFileURL = filestoreBase + prefix + "SHA256SUMS"
	}

	httpResponse, err := http.Get(hashFileURL)
	if err != nil {
		return fmt.Errorf("downloading %q: %w", hashFileURL, err)
	}
	if httpResponse.StatusCode != 200 {
		return fmt.Errorf("unexpected status getting %q: %v", hashFileURL, httpResponse.Status)
	}
	defer httpResponse.Body.Close()

	b, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return fmt.Errorf("reading body %q: %w", hashFileURL, err)
	}

	m := &manifest{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "-----BEGIN PGP SIGNED MESSAGE-----" {
			// Start of the PGP boilerplate
			continue
		}
		if line == "-----BEGIN PGP SIGNATURE-----" {
			// Part of the PGP signature; end of signed content
			break
		}
		if line == "Hash: SHA256" {
			// Part of the PGP boilerplate
			continue
		}
		tokens := strings.Fields(line)
		if len(tokens) != 2 {
			return fmt.Errorf("unexpected line %q (expected 2 tokens)", line)
		}
		hash := tokens[0]
		name := tokens[1]
		name = removeBadPrefix(name)

		if exclude.Matches(prefix + name) {
			continue
		}
		m.Files = append(m.Files, file{
			Name:   prefix + name,
			SHA256: hash,
		})
	}

	out, err := yaml.Marshal(&m)
	if err != nil {
		return fmt.Errorf("building yaml: %w", err)
	}
	if _, err := os.Stdout.Write(out); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

type manifest struct {
	// FileStores []fileStore `json:"filestores,omitempty"`
	Files []file `json:"files,omitempty"`
}

type file struct {
	Name   string `json:"name,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

// removeBadPrefix is a hack for some of the older kubernetes sha256sum files,
// while accidentally included an invalid prefix
func removeBadPrefix(name string) string {
	badPrefix := "/workspace/src/k8s.io/kubernetes/_output"
	if !strings.HasPrefix(name, badPrefix) {
		return name
	}
	tokens := strings.Split(name, "/")
	return strings.Join(tokens[8:], "/")
}
