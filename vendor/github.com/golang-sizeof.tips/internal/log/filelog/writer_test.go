package filelog

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

var (
	bunch1 = map[string]uint32{
		"testing.log":      86400,
		"testing.log.asd":  86300,
		"testing.log.002":  86100,
		"testing.log.":     200,
		"testing.log023":   150,
		"testing.log.1123": 100,
	}
	bunch2 = map[string]uint32{
		"super-test.log": 700,
	}
)

func createTestFiles(bunch map[string]uint32) (dirName string) {
	now := time.Now()
	dirName, err := ioutil.TempDir("", "writer_test_")
	if err != nil {
		log.Panicf("failed to create temp directory: %s", err.Error())
	}
	for fName, modified := range bunch {
		fileName := filepath.Join(dirName, fName)
		if err := ioutil.WriteFile(fileName, []byte("test file"), 0660); err != nil {
			log.Panicf("failed to create log file '%s', reason: %s", fName, err.Error())
		}
		if err := os.Chtimes(fileName,
			now.Add(-1*time.Duration(modified)*time.Second),
			now.Add(-1*time.Duration(modified)*time.Second),
		); err != nil {
			log.Panicf("failed to change times of log file '%s', reason: %s", fName, err.Error())
		}
	}
	return
}

func removeTestFiles(dirName string) {
	if err := os.RemoveAll(dirName); err != nil {
		log.Panicf("failed to remove temp directory '%s', reason: %s", dirName, err.Error())
	}
}

func TestProcessAlreadyRotatedFiles(t *testing.T) {
	test := func(bunch map[string]uint32, filename string, keepSeconds uint64, filesKept map[string]bool, nextFilename string) {
		dir := createTestFiles(bunch)
		defer removeTestFiles(dir)

		w := &Writer{
			filename: filepath.Join(dir, filename),
		}
		w.SetRotatedFilesExpiration(keepSeconds)

		fName := w.processAlreadyRotatedFiles()
		if needName := filepath.Join(dir, nextFilename); fName != needName {
			t.Errorf("fileNameForRotation expected '%s', got '%s'", needName, fName)
		}

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Panicf("failed to read temp directory '%s', reason: %s", dir, err.Error())
		}
		filesNum, i := len(filesKept), 1
		for _, file := range files {
			if i > filesNum {
				t.Errorf("more then %d were left in logs folder", filesNum)
			}
			if _, ok := filesKept[file.Name()]; !ok {
				t.Errorf("file '%s' sould be deleted, but it'is not", file.Name())
			}
			i++
		}
	}

	test(bunch1, "testing.log", 500, map[string]bool{
		"testing.log":      true,
		"testing.log.":     true,
		"testing.log023":   true,
		"testing.log.1123": true,
	}, "testing.log.1124")
	test(bunch2, "super-test.log", 800, map[string]bool{
		"super-test.log": true,
	}, "super-test.log.001")
}

func TestOpenNewFile(t *testing.T) {
	test := func(bunch map[string]uint32, filename, daily_opendate string, lines, size uint64) {
		dir := createTestFiles(bunch)
		defer removeTestFiles(dir)

		w := &Writer{
			filename: filepath.Join(dir, filename),
			waiter:   &sync.WaitGroup{},
		}

		if err := w.openNewFile(); err != nil {
			t.Error("failed to open file")
		}
		defer w.closeCurrentFile()

		if w.maxlines_curlines != lines {
			t.Errorf("maxlines_curlines expected %d, got %d", lines, w.maxlines_curlines)
		}
		if w.maxsize_cursize != size {
			t.Errorf("maxsize_cursize expected %d, got %d", size, w.maxsize_cursize)
		}
		if w.daily_opendate != daily_opendate {
			t.Errorf("daily_opendate expected '%s', got '%s'", daily_opendate, w.daily_opendate)
		}
	}

	test(bunch1, "testing.log", time.Now().Add(-86400*time.Second).Format(dayFormat), 1, 9)
	test(bunch2, "test.log", time.Now().Format(dayFormat), 0, 0)
}

func TestSetWaitOnClose(t *testing.T) {
	dir := createTestFiles(bunch2)
	defer removeTestFiles(dir)

	w := &Writer{
		filename: filepath.Join(dir, "super-test.log"),
		waiter:   &sync.WaitGroup{},
	}
	w.SetWaitOnClose(true)

	if err := w.openNewFile(); err != nil {
		t.Error("failed to open file")
	}
	w.closeCurrentFile()

	if int(w.file.Fd()) != -1 {
		t.Errorf("waiting failed, file is still not closed")
	}
}
