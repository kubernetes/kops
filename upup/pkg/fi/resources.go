package fi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

type Resource interface {
	Open() (io.ReadSeeker, error)
}

type TemplateResource interface {
	Resource
	Curry(args []string) TemplateResource
}

func HashForResource(r Resource, hashAlgorithm HashAlgorithm) (string, error) {
	hasher := NewHasher(hashAlgorithm)
	_, err := CopyResource(hasher, r)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
		}
		return "", fmt.Errorf("error while hashing resource: %v", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func HashesForResource(r Resource, hashAlgorithms []HashAlgorithm) (map[HashAlgorithm]string, error) {
	hashers := make(map[HashAlgorithm]hash.Hash)
	var writers []io.Writer
	for _, hashAlgorithm := range hashAlgorithms {
		if hashers[hashAlgorithm] != nil {
			continue
		}
		hasher := NewHasher(hashAlgorithm)
		hashers[hashAlgorithm] = hasher
		writers = append(writers, hasher)
	}

	w := io.MultiWriter(writers...)

	_, err := CopyResource(w, r)
	if err != nil {
		return nil, fmt.Errorf("error while hashing resource: %v", err)
	}

	hashes := make(map[HashAlgorithm]string)
	for k, hasher := range hashers {
		hashes[k] = hex.EncodeToString(hasher.Sum(nil))
	}

	return hashes, nil
}

func ResourcesMatch(a, b Resource) (bool, error) {
	aReader, err := a.Open()
	if err != nil {
		return false, err
	}
	defer SafeClose(aReader)

	bReader, err := b.Open()
	if err != nil {
		return false, err
	}
	defer SafeClose(bReader)

	const size = 8192
	aData := make([]byte, size)
	bData := make([]byte, size)

	for {
		aN, aErr := io.ReadFull(aReader, aData)
		if aErr != nil && aErr != io.EOF && aErr != io.ErrUnexpectedEOF {
			return false, aErr
		}

		bN, bErr := io.ReadFull(bReader, bData)
		if bErr != nil && bErr != io.EOF && bErr != io.ErrUnexpectedEOF {
			return false, bErr
		}

		if aErr == nil && bErr == nil {
			if aN != size || bN != size {
				panic("violation of io.ReadFull contract")
			}
			if !bytes.Equal(aData, bData) {
				return false, nil
			}
			continue
		}

		if aN != bN {
			return false, nil
		}

		return bytes.Equal(aData[0:aN], bData[0:bN]), nil
	}
}

func CopyResource(dest io.Writer, r Resource) (int64, error) {
	in, err := r.Open()
	if err != nil {
		if os.IsNotExist(err) {
			return 0, err
		}
		return 0, fmt.Errorf("error opening resource: %v", err)
	}
	defer SafeClose(in)

	n, err := io.Copy(dest, in)
	if err != nil {
		return n, fmt.Errorf("error copying resource: %v", err)
	}
	return n, nil
}

func ResourceAsString(r Resource) (string, error) {
	buf := new(bytes.Buffer)
	_, err := CopyResource(buf, r)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ResourceAsBytes(r Resource) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := CopyResource(buf, r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type StringResource struct {
	s string
}

var _ Resource = &StringResource{}

func NewStringResource(s string) *StringResource {
	return &StringResource{s: s}
}

func (s *StringResource) Open() (io.ReadSeeker, error) {
	r := bytes.NewReader([]byte(s.s))
	return r, nil
}

func (s *StringResource) WriteTo(out io.Writer) error {
	_, err := out.Write([]byte(s.s))
	return err
}

type BytesResource struct {
	data []byte
}

var _ Resource = &BytesResource{}

func NewBytesResource(data []byte) *BytesResource {
	return &BytesResource{data: data}
}

func (r *BytesResource) Open() (io.ReadSeeker, error) {
	reader := bytes.NewReader([]byte(r.data))
	return reader, nil
}

type FileResource struct {
	Path string
}

var _ Resource = &FileResource{}

func NewFileResource(path string) *FileResource {
	return &FileResource{Path: path}
}

func (r *FileResource) Open() (io.ReadSeeker, error) {
	in, err := os.Open(r.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("error opening file %q: %v", r.Path, err)
	}
	return in, err
}
