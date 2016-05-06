package fi

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"net/http"
	"os"
	"path"
)

func DownloadURL(url string, dest string, hash string) (string, error) {
	if hash != "" {
		match, err := fileHasHash(dest, hash)
		if err != nil {
			return "", err
		}
		if match {
			return hash, nil
		}
	}

	dirMode := os.FileMode(0755)
	err := downloadURLAlways(url, dest, dirMode)
	if err != nil {
		return "", err
	}

	if hash != "" {
		match, err := fileHasHash(dest, hash)
		if err != nil {
			return "", err
		}
		if !match {
			return "", fmt.Errorf("downloaded from %q but hash did not match expected %q", url, hash)
		}
	} else {
		hash, err = HashFile(dest, HashAlgorithmSHA256)
		if err != nil {
			return "", err
		}
	}

	return hash, nil
}

func downloadURLAlways(url string, destPath string, dirMode os.FileMode) error {
	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return fmt.Errorf("error creating directories for destination file %q: %v", destPath, err)
	}

	output, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating file for download %q: %v", destPath, err)
	}
	defer output.Close()

	glog.Infof("Downloading %q", url)

	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error doing HTTP fetch of %q: %v", url, err)
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return fmt.Errorf("error downloading HTTP content from %q: %v", url, err)
	}
	return nil
}
