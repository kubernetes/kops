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
	"regexp"
	"strings"
	"testing"
)

var compressedUserDataPattern = regexp.MustCompile(`echo "([A-Za-z0-9+/=]+)" \| base64 -d \| gzip -d`)

func normalizeCompressedUserData(content string) (string, error) {
	content = strings.ReplaceAll(strings.TrimSpace(content), "\r\n", "\n")
	var decodeErr error
	normalized := compressedUserDataPattern.ReplaceAllStringFunc(content, func(command string) string {
		payload := compressedUserDataPattern.FindStringSubmatch(command)[1]
		compressed, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			decodeErr = err
			return command
		}
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			decodeErr = err
			return command
		}
		decoded, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			decodeErr = err
			return command
		}
		return fmt.Sprintf(`echo "%s" | base64 -d | gzip -d`, base64.StdEncoding.EncodeToString(decoded))
	})
	return normalized, decodeErr
}

func TestNormalizeCompressedUserData(t *testing.T) {
	encode := func(content string, level int) string {
		t.Helper()
		var buffer bytes.Buffer
		writer, err := gzip.NewWriterLevel(&buffer, level)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("echo \"%s\" | base64 -d | gzip -d > conf/kube_env.yaml\n", base64.StdEncoding.EncodeToString(buffer.Bytes()))
	}
	actual := encode("version: 1.37.0-beta.1\n", gzip.BestSpeed)
	expected := encode("version: 1.37.0-beta.1\n", gzip.BestCompression)
	if actual == expected {
		t.Fatal("test requires different gzip encodings")
	}
	baseline, err := normalizeCompressedUserData(expected)
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name    string
		content string
		equal   bool
		wantErr bool
	}{
		{name: "different encoding", content: actual, equal: true},
		{name: "CRLF encoding", content: strings.ReplaceAll(actual, "\n", "\r\n"), equal: true},
		{name: "outer whitespace", content: "\n" + actual + "\n", equal: true},
		{name: "different payload", content: encode("version: 0.0.0\n", gzip.BestSpeed)},
		{name: "different command", content: actual + "exit 1\n"},
		{name: "invalid gzip", content: `echo "YWJj" | base64 -d | gzip -d`, wantErr: true},
		{name: "invalid base64", content: `echo "A" | base64 -d | gzip -d`, wantErr: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			normalized, err := normalizeCompressedUserData(testCase.content)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if (normalized == baseline) != testCase.equal {
				t.Fatalf("unexpected comparison result: %q", normalized)
			}
		})
	}
}
