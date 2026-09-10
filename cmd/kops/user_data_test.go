/*
Copyright The Kubernetes Authors.

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
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"k8s.io/kops/pkg/testutils/golden"
)

var compressedKubeEnv = regexp.MustCompile(`(?m)^echo ("[^"\n]*") \| base64 -d \| gzip -d > conf/kube_env\.yaml$`)

// normalizeUserData compares the bytes installed by the bootstrap script, since
// Go's gzip encoder may produce different bytes across toolchain versions.
// Only the encoded literal is replaced; the shell commands remain part of the comparison.
func normalizeUserData(input string) (string, error) {
	input = strings.ReplaceAll(strings.TrimSpace(input), "\r\n", "\n")
	var result strings.Builder
	end := 0
	for _, match := range compressedKubeEnv.FindAllStringSubmatchIndex(input, -1) {
		start, stop := match[2], match[3]
		compressed, err := base64.StdEncoding.DecodeString(input[start+1 : stop-1])
		if err != nil {
			return "", fmt.Errorf("decoding kube_env base64: %w", err)
		}
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return "", fmt.Errorf("opening kube_env gzip: %w", err)
		}
		// Read through EOF to verify the checksum and detect truncated gzip data.
		payload, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return "", fmt.Errorf("reading kube_env gzip: %w", err)
		}
		result.WriteString(input[end:start])
		result.WriteString(strconv.Quote(string(payload)))
		end = stop
	}
	result.WriteString(input[end:])
	return result.String(), nil
}

func assertUserDataMatchesFile(t *testing.T, actual, expectedPath string) {
	t.Helper()
	actualNormalized, err := normalizeUserData(actual)
	if err != nil {
		t.Fatalf("invalid actual user-data %q: %v", expectedPath, err)
	}
	if expected, err := os.ReadFile(expectedPath); err == nil {
		expectedNormalized, err := normalizeUserData(string(expected))
		if err != nil {
			t.Fatalf("invalid expected user-data %q: %v", expectedPath, err)
		}
		if actualNormalized == expectedNormalized {
			return
		}
	}
	// Keep the original generated script in golden files, including when updating them.
	golden.AssertMatchesFile(t, actual, expectedPath)
}

func TestNormalizeUserData(t *testing.T) {
	encode := func(payload string, level int) string {
		t.Helper()
		var b bytes.Buffer
		writer, err := gzip.NewWriterLevel(&b, level)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write([]byte(payload)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return base64.StdEncoding.EncodeToString(b.Bytes())
	}
	script := func(encoded string) string {
		return "#!/bin/bash\n" + `echo "` + encoded + `" | base64 -d | gzip -d > conf/kube_env.yaml` + "\necho done\n"
	}
	payload := "clusterName: test.example.com\nnodeAddresses:\n" + strings.Repeat("  - 10.0.0.1\n", 16)
	encoded := encode(payload, gzip.DefaultCompression)
	expected := script(encoded)
	normalized, err := normalizeUserData(expected)
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.TrimSpace(script("")); strings.Replace(normalized, strconv.Quote(payload), `""`, 1) != want {
		t.Fatal("normalization changed the surrounding script or decoded payload")
	}
	otherEncoding := encode(payload, gzip.NoCompression)
	if otherEncoding == encoded {
		t.Fatal("test requires different gzip encodings")
	}
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	corrupt := bytes.Clone(compressed)
	corrupt[len(corrupt)-8] ^= 1

	for _, tc := range []struct {
		name    string
		input   string
		equal   bool
		wantErr bool
	}{
		{name: "same encoding", input: expected, equal: true},
		{name: "different encoding", input: script(otherEncoding), equal: true},
		{name: "Windows fixture line endings", input: strings.ReplaceAll(expected, "\n", "\r\n"), equal: true},
		{name: "changed YAML", input: script(encode(payload+"changed: true\n", gzip.DefaultCompression))},
		{name: "changed YAML whitespace", input: script(encode(strings.TrimSuffix(payload, "\n"), gzip.DefaultCompression))},
		{name: "invalid base64", input: script("%%%"), wantErr: true},
		{name: "invalid gzip header", input: script(base64.StdEncoding.EncodeToString([]byte("not gzip"))), wantErr: true},
		{name: "bad checksum", input: script(base64.StdEncoding.EncodeToString(corrupt)), wantErr: true},
		{name: "truncated gzip", input: script(base64.StdEncoding.EncodeToString(compressed[:len(compressed)-1])), wantErr: true},
		{name: "changed destination", input: strings.Replace(expected, "conf/kube_env.yaml", "conf/wrong.yaml", 1)},
		{name: "missing decompression", input: strings.Replace(expected, " | gzip -d", "", 1)},
		{name: "missing payload", input: compressedKubeEnv.ReplaceAllString(expected, "")},
		{name: "duplicate payload", input: expected + compressedKubeEnv.FindString(expected)},
		{name: "changed shell command", input: strings.Replace(expected, "echo done", "echo failed", 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeUserData(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("normalizeUserData error = %v, want error = %v", err, tc.wantErr)
			}
			if err == nil && (actual == normalized) != tc.equal {
				t.Errorf("normalized equality = %v, want %v", actual == normalized, tc.equal)
			}
		})
	}
}
