package fi

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"k8s.io/kube-deploy/upup/pkg/fi/hashing"
	"net/http"
	"os"
	"path"
)

func DownloadURL(url string, dest string, hash *hashing.Hash) (*hashing.Hash, error) {
	if hash != nil {
		match, err := fileHasHash(dest, hash)
		if err != nil {
			return nil, err
		}
		if match {
			return hash, nil
		}
	}

	dirMode := os.FileMode(0755)
	err := downloadURLAlways(url, dest, dirMode)
	if err != nil {
		return nil, err
	}

	if hash != nil {
		match, err := fileHasHash(dest, hash)
		if err != nil {
			return nil, err
		}
		if !match {
			return nil, fmt.Errorf("downloaded from %q but hash did not match expected %q", url, hash)
		}
	} else {
		hash, err = hashing.HashAlgorithmSHA256.HashFile(dest)
		if err != nil {
			return nil, err
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
