package vfs

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi/hashing"
	"os"
)

// VFSScan scans a source Path for changes files
type VFSScan struct {
	Base   Path
	hashes map[string]*hashing.Hash
}

func NewVFSScan(base Path) *VFSScan {
	return &VFSScan{Base: base}
}

type ChangeType string

const ChangeType_Added ChangeType = "ADDED"
const ChangeType_Removed ChangeType = "REMOVED"
const ChangeType_Modified ChangeType = "MODIFIED"

type Change struct {
	ChangeType ChangeType
	Path       Path
	Hash       *hashing.Hash
}

// Scans for changes files.  On the first call will return all files as ChangeType_Added.
// On subsequent calls will return any changed files (using their hashes)
func (v *VFSScan) Scan() ([]Change, error) {
	allFiles, err := v.Base.ReadTree()
	if err != nil {
		return nil, fmt.Errorf("Error reading dir %q: %v", v.Base, err)
	}

	files := make(map[string]Path)
	hashes := make(map[string]*hashing.Hash)
	for _, f := range allFiles {
		key := f.Path()
		files[key] = f
		hasHash, ok := f.(HasHash)
		if !ok {
			return nil, fmt.Errorf("Source must support hashing: %T", f)
		}
		hash, err := hasHash.PreferredHash()
		if err != nil {
			return nil, fmt.Errorf("Error hashing %q: %v", key, err)
		}

		hashes[key] = hash
	}

	if v.hashes == nil {
		v.hashes = hashes
		var changes []Change
		for k, f := range files {
			hash := hashes[k]
			changes = append(changes, Change{ChangeType: ChangeType_Added, Path: f, Hash: hash})
		}
		return changes, nil
	}

	var changes []Change
	for k, f := range files {
		oldHash := v.hashes[k]
		newHash := hashes[k]

		if oldHash == nil {
			changes = append(changes, Change{ChangeType: ChangeType_Added, Path: f, Hash: newHash})
		} else if !oldHash.Equal(newHash) {
			changes = append(changes, Change{ChangeType: ChangeType_Modified, Path: f, Hash: newHash})
		}
	}

	for k := range v.hashes {
		newHash := hashes[k]
		f := files[k]
		if newHash == nil {
			changes = append(changes, Change{ChangeType: ChangeType_Removed, Path: f, Hash: newHash})
		}
	}

	v.hashes = hashes
	return changes, nil
}

func SyncDir(src *VFSScan, destBase Path) error {
	changes, err := src.Scan()
	if err != nil {
		return fmt.Errorf("Error scanning source dir %q: %v", src, err)
	}

	for _, change := range changes {
		f := change.Path

		relativePath, err := RelativePath(f, src.Base)
		if err != nil {
			return err
		}

		destFile := destBase.Join(relativePath)

		switch change.ChangeType {
		case ChangeType_Removed:
			err := destFile.Remove()
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("error removing file %q: %v", destFile, err)
				}
			}
			continue

		case ChangeType_Modified, ChangeType_Added:
			break

		default:
			return fmt.Errorf("unknown change type: %q", change.ChangeType)
		}

		hashMatch, err := hashesMatch(f, destFile)
		if err != nil {
			return err
		}
		if hashMatch {
			glog.V(2).Infof("File hashes match: %s and %s", f, destFile)
			continue
		}

		srcData, err := f.ReadFile()
		if err != nil {
			return fmt.Errorf("error reading source file %q: %v", f, err)
		}

		destData, err := destFile.ReadFile()
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("error reading dest file %q: %v", f, err)
			}
		}

		if destData == nil || !bytes.Equal(srcData, destData) {
			glog.V(2).Infof("Copying data from %s to %s", f, destFile)
			err = destFile.WriteFile(srcData)
			if err != nil {
				return fmt.Errorf("error writing dest file %q: %v", f, err)
			}
		}
	}

	return nil
}

func hashesMatch(src, dest Path) (bool, error) {
	sh, ok := src.(HasHash)
	if !ok {
		return false, nil
	}

	dh, ok := dest.(HasHash)
	if !ok {
		return false, nil
	}

	{
		srcHash, err := sh.PreferredHash()
		if err != nil {
			glog.Warningf("error getting hash of source file %s: %v", src, err)
		} else if srcHash != nil {
			destHash, err := dh.Hash(srcHash.Algorithm)
			if err != nil {
				glog.Warningf("error comparing hash of dest file %s: %v", dest, err)
			} else if destHash != nil {
				return destHash.Equal(srcHash), nil
			}
		}
	}

	{
		destHash, err := dh.PreferredHash()
		if err != nil {
			glog.Warningf("error getting hash of dest file %s: %v", src, err)
		} else if destHash != nil {
			srcHash, err := dh.Hash(destHash.Algorithm)
			if err != nil {
				glog.Warningf("error comparing hash of src file %s: %v", dest, err)
			} else if srcHash != nil {
				return srcHash.Equal(destHash), nil
			}
		}
	}

	glog.Infof("No compatible hash: %s and %s", src, dest)
	return false, nil
}
