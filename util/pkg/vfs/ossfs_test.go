/*
Copyright 2018 The Kubernetes Authors.

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

package vfs

import (
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/denverdino/aliyungo/oss"
)

var (
	client     *oss.Client
	testBucket = "kops-test-bucket"
	testKey    = "kops-test-key"
	TestRegion = oss.Region("oss-cn-hangzhou")

	testData = []string{
		"The quick brown fox jumps over the lazy dog",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat",
		"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur",
	}
	testDirKeys = []string{
		"testdir/item_0",
		"testdir/item_1",
		"testdir/dir_0/item_2",
		"testdir/dir_0/dir_1/item3",
		"testdir/dir_0/dir_1/item4",
	}
	testLargeDirKeys = func() []string {
		baseDir := "testlargedir/"
		keys := make([]string, 2018)
		for i := 0; i < 2018; i++ {
			id := strconv.Itoa(i)
			if i%4 == 0 {
				keys[i] = baseDir + "dir_0/dir_1/item_" + id
			} else if i%2 == 0 {
				keys[i] = baseDir + "dir_0/item_" + id
			} else {
				keys[i] = baseDir + "item_" + id
			}
		}
		return keys
	}()
)

func init() {
	AccessKeyId := os.Getenv("ALIYUN_ACCESS_KEY_ID")
	AccessKeySecret := os.Getenv("ALIYUN_ACCESS_KEY_SECRET")

	if len(AccessKeyId) != 0 && len(AccessKeySecret) != 0 {
		client = oss.NewOSSClient(TestRegion, false, AccessKeyId, AccessKeySecret, false)
	} else {
		// TODO: error handling
		log.Fatalf("Unable to initialize client")
	}

	if !isBucketExists(testBucket) {
		createBucket(testBucket)
	}
}

func Test_OSSPath_Parse(t *testing.T) {
	grid := []struct {
		Input          string
		ExpectError    bool
		ExpectedBucket string
		ExpectedPath   string
	}{
		{
			Input:          "oss://bucket",
			ExpectedBucket: "bucket",
			ExpectedPath:   "",
		},
		{
			Input:          "oss://bucket/path",
			ExpectedBucket: "bucket",
			ExpectedPath:   "path",
		},
		{
			Input:          "oss://bucket2/path/subpath",
			ExpectedBucket: "bucket2",
			ExpectedPath:   "path/subpath",
		},
		{
			Input:       "oss:///bucket/path/subpath",
			ExpectError: true,
		},
	}
	for _, g := range grid {
		osspath, err := Context.buildOSSPath(g.Input)
		t.Logf("%v, %v\n", osspath, err)
		if !g.ExpectError {
			if err != nil {
				t.Fatalf("unexpected error parsing OSS path: %v", err)
			}
			if osspath.bucket != g.ExpectedBucket {
				t.Fatalf("unexpected OSS path: %v", osspath)
			}
			if osspath.key != g.ExpectedPath {
				t.Fatalf("unexpected oss path: %v", osspath)
			}
		} else {
			if err == nil {
				t.Fatalf("unexpected error parsing %q", g.Input)
			}
		}
	}
}

func Test_ReadDir(t *testing.T) {
	putMultiFiles(testBucket, testDirKeys)
	defer delMultiFiles(testBucket, testDirKeys)

	expectedKeys := []string{
		"testdir/item_0",
		"testdir/item_1",
		"testdir/dir_0/",
	}
	p := &OSSPath{
		client: client,
		bucket: testBucket,
	}

	p.key = "testdir/"
	paths, err := p.ReadDir()
	if err != nil {
		t.Fatalf("Unable to read directory: %v", err)
	} else if len(paths) != len(expectedKeys) {
		t.Fatalf("Number of items conflicts: expect %d, get %d", len(expectedKeys), len(paths))
	} else {
		var foundKeys []string
		for _, path := range paths {
			osspath := path.(*OSSPath)
			foundKeys = append(foundKeys, osspath.key)
		}
		// TODO: any better compare method?
		sort.Strings(expectedKeys)
		sort.Strings(foundKeys)
		if !reflect.DeepEqual(foundKeys, expectedKeys) {
			t.Fatal("Read directory failed")
		}
	}
}

func Test_ReadTree(t *testing.T) {
	putMultiFiles(testBucket, testDirKeys)
	defer delMultiFiles(testBucket, testDirKeys)

	p := &OSSPath{
		client: client,
		bucket: testBucket,
	}

	p.key = "testdir/"
	paths, err := p.ReadTree()
	if err != nil {
		t.Fatalf("Unable to read directory: %v", err)
	} else if len(paths) != len(testDirKeys) {
		t.Fatalf("Number of items conflicts: expect %d, get %d", len(testDirKeys), len(paths))
	} else {
		var foundKeys []string
		for _, path := range paths {
			osspath := path.(*OSSPath)
			foundKeys = append(foundKeys, osspath.key)
		}
		// TODO: any better compare method?
		sort.Strings(testDirKeys)
		sort.Strings(foundKeys)
		if !reflect.DeepEqual(foundKeys, testDirKeys) {
			t.Fatal("Read directory failed")
		}
	}
}

func Test_ReadLargeDir(t *testing.T) {
	if !isObjectExists(testBucket, "testlargedir/item_1") {
		putMultiFiles(testBucket, testLargeDirKeys)
		defer delMultiFiles(testBucket, testLargeDirKeys)
	}

	expectedKeys := []string{
		"testlargedir/dir_0/",
	}
	for i := 0; i < len(testLargeDirKeys); i++ {
		if i%2 != 0 {
			expectedKeys = append(expectedKeys, "testlargedir/item_"+strconv.Itoa(i))
		}
	}

	p := &OSSPath{
		client: client,
		bucket: testBucket,
		key:    "testlargedir/",
	}

	paths, err := p.ReadDir()
	if err != nil {
		t.Fatalf("Unable to read directory: %v", err)
	} else if len(paths) != len(expectedKeys) {
		t.Fatalf("Number of items conflicts: expect %d, get %d", len(expectedKeys), len(paths))
	} else {
		var foundKeys []string
		for _, path := range paths {
			osspath := path.(*OSSPath)
			foundKeys = append(foundKeys, osspath.key)
		}
		// TODO: any better compare method?
		sort.Strings(expectedKeys)
		sort.Strings(foundKeys)
		if !reflect.DeepEqual(foundKeys, expectedKeys) {
			t.Fatal("Read directory failed")
		}
	}
}

func Test_ReadLargeTree(t *testing.T) {
	if !isObjectExists(testBucket, "testlargedir/item_1") {
		putMultiFiles(testBucket, testLargeDirKeys)
		defer delMultiFiles(testBucket, testLargeDirKeys)
	}

	p := &OSSPath{
		client: client,
		bucket: testBucket,
		key:    "testlargedir/",
	}

	paths, err := p.ReadTree()
	if err != nil {
		t.Fatalf("Unable to read directory: %v", err)
	} else if len(paths) != len(testLargeDirKeys) {
		t.Fatalf("Number of items conflicts: expect %d, get %d", len(testLargeDirKeys), len(paths))
	} else {
		var foundKeys []string
		for _, path := range paths {
			osspath := path.(*OSSPath)
			foundKeys = append(foundKeys, osspath.key)
		}
		// TODO: any better compare method?
		sort.Strings(testLargeDirKeys)
		sort.Strings(foundKeys)
		if !reflect.DeepEqual(foundKeys, testLargeDirKeys) {
			t.Fatal("Read directory failed")
		}
	}
}
func Test_CreateFile(t *testing.T) {
	grid := []struct {
		Key           string
		ExpectError   bool
		ExpectContent string
	}{
		{
			Key:           "test-createfile-key",
			ExpectError:   false,
			ExpectContent: testData[2],
		},
		{
			Key:           "test-createfile-key",
			ExpectError:   true,
			ExpectContent: testData[2],
		},
	}
	p := &OSSPath{
		client: client,
		bucket: testBucket,
	}
	acl := oss.Private

	for _, g := range grid {
		p.key = g.Key
		reader := strings.NewReader(g.ExpectContent)
		err := p.CreateFile(reader, acl)
		if g.ExpectError {
			if err == nil || err != os.ErrExist {
				t.Fatalf("Unexpected error creating file: %v", err)
			}
		} else {
			if err != nil {
				t.Fatalf("Unable to create file: %v", err)
			} else {
				// verification
				content := string(getFileByRawInterface(p.bucket, p.key))
				if content != g.ExpectContent {
					t.Fatalf("File contents conflict: expect '%s', get '%s'", g.ExpectContent, content)
				}
			}
		}
	}
}

func Test_Remove(t *testing.T) {
	grid := []struct {
		Key         string
		ExpectError bool
	}{
		{
			Key:         "test-createfile-key",
			ExpectError: false,
		},
		// TODO: aliyun SDK bugs: deleting a non-existent file does not raise error
		// {
		// 	Key:         "test-createfile-key",
		// 	ExpectError: true,
		// },
	}
	p := &OSSPath{
		client: client,
		bucket: testBucket,
	}

	for _, g := range grid {
		p.key = g.Key
		err := p.Remove()
		if g.ExpectError {
			if err == nil {
				t.Fatalf("Unexpected error removing file: %v", err)
			}
		} else {
			if err != nil {
				t.Fatalf("Unable to remove file: %v", err)
			}
		}
	}
}

func Test_WriteFile(t *testing.T) {
	p := &OSSPath{
		client: client,
		bucket: testBucket,
		key:    testKey,
	}

	// TODO: any test on ACL mode?
	acl := oss.Private
	reader := strings.NewReader(testData[0])
	err := p.WriteFile(reader, acl)
	if err != nil {
		t.Fatalf("Unable to write file: %v", err)
	}

	// verfication
	content := string(getFileByRawInterface(p.bucket, p.key))
	if content != testData[0] {
		t.Fatalf("File contents conflict: expect '%s', get '%s'", testData[0], content)
	}
}

func Test_ReadFile(t *testing.T) {
	p := &OSSPath{
		client: client,
		bucket: testBucket,
		key:    testKey,
	}

	// pre write
	putFileByRawInterface(p.bucket, p.key, []byte(testData[1]))

	data, err := p.ReadFile()
	if err != nil {
		t.Fatalf("Unable to read file: %v", err)
	} else if string(data) != testData[1] {
		t.Fatalf("File contents conflict: expect '%s', get '%s'", testData[1], string(data))
	}
}

func getFileByRawInterface(bucket string, key string) []byte {
	data, err := client.Bucket(bucket).Get(key)
	if err != nil {
		log.Fatalf("error in reading file: %v", err)
	}
	return data
}

func putFileByRawInterface(bucket string, key string, data []byte) {
	contType := "application/octet-stream"
	perm := oss.Private
	err := client.Bucket(bucket).Put(key, data, contType, perm, oss.Options{})
	if err != nil {
		log.Fatalf("error in writing file: %v", err)
	}
}

func putMultiFiles(bucket string, files []string) {
	acl := oss.Private
	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, k := range files {
		go func(key string) {
			path := &OSSPath{
				client: client,
				bucket: bucket,
				key:    key,
			}
			reader := strings.NewReader("")
			err := path.CreateFile(reader, acl)
			wg.Done()
			if err != nil {
				log.Fatalf("error in creating files: %v", err)
			}
		}(k)
	}
	wg.Wait()
}

func delMultiFiles(bucket string, files []string) {
	// teardown
	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, k := range files {
		go func(key string) {
			path := &OSSPath{
				client: client,
				bucket: bucket,
				key:    key,
			}
			err := path.Remove()
			wg.Done()
			if err != nil {
				log.Fatalf("error in teardown: %v", err)
			}
		}(k)
	}
	wg.Wait()
}

func createBucket(bucket string) {
	// Create bucket on aliyun
	b := client.Bucket(bucket)
	err := b.PutBucket(oss.Private)
	if err != nil {
		log.Fatalf("Unable to create bucket: %v", err)
	}
}

func deleteBucket(bucket string) {
	// Delete bucket on aliyun
	b := client.Bucket(bucket)
	err := b.DelBucket()
	if err != nil {
		log.Fatalf("Unable to delete bucket: %v", err)
	}
}

func isBucketExists(bucket string) bool {
	b := client.Bucket(bucket)
	_, err := b.Info()
	if err == nil {
		return true
	}
	ossErr, ok := err.(*oss.Error)
	// TODO: maybe it's subjective to say only 404 represents NOT_EXISTS
	return !(ok && ossErr.StatusCode == 404)
}

func isObjectExists(bucket string, path string) bool {
	exists, err := client.Bucket(bucket).Exists(path)
	if err != nil {
		log.Fatalf("error in checking object existence: %v", err)
	}
	return exists
}
