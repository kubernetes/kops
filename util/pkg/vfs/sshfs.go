/*
Copyright 2019 The Kubernetes Authors.

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
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
)

type SSHPath struct {
	client *ssh.Client
	sudo   bool
	server string
	path   string
}

type SSHAcl struct {
	Mode os.FileMode
}

var _ Path = &SSHPath{}

func NewSSHPath(client *ssh.Client, server string, path string, sudo bool) *SSHPath {
	return &SSHPath{
		client: client,
		server: server,
		path:   path,
		sudo:   sudo,
	}
}

func (p *SSHPath) newClient(ctx context.Context) (*sftp.Client, error) {
	if !p.sudo {
		sftpClient, err := sftp.NewClient(p.client)
		if err != nil {
			return nil, fmt.Errorf("error creating sftp client: %w", err)
		}

		return sftpClient, nil
	}
	s, err := p.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("error creating sftp client (in new-session): %w", err)
	}

	stdin, err := s.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating sftp client (at stdin pipe): %w", err)
	}
	stdout, err := s.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating sftp client (at stdout pipe): %w", err)
	}

	err = s.Start("sudo /usr/lib/openssh/sftp-server")
	if err != nil {
		return nil, fmt.Errorf("error creating sftp client (executing 'sudo /usr/lib/openssh/sftp-server'): %w", err)
	}

	c, err := sftp.NewClientPipe(stdout, stdin)
	if err != nil {
		return nil, fmt.Errorf("error starting sftp (executing 'sudo /usr/lib/openssh/sftp-server'): %w", err)
	}
	return c, nil
}

func (p *SSHPath) Path() string {
	return "ssh://" + p.server + p.path
}

func (p *SSHPath) String() string {
	return p.Path()
}

func (p *SSHPath) Remove() error {
	ctx := context.TODO()

	sftpClient, err := p.newClient(ctx)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	err = sftpClient.Remove(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("error deleting %s: %w", p, err)
	}

	return nil
}

func (p *SSHPath) RemoveAll() error {
	tree, err := p.ReadTree()
	if err != nil {
		return err
	}

	for _, filePath := range tree {
		err := filePath.Remove()
		if err != nil {
			return fmt.Errorf("error removing file %s: %w", filePath, err)
		}
	}

	return nil
}

func (p *SSHPath) RemoveAllVersions() error {
	return p.Remove()
}

func (p *SSHPath) Join(relativePath ...string) Path {
	args := []string{p.path}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return NewSSHPath(p.client, p.server, joined, p.sudo)
}

func mkdirAll(sftpClient *sftp.Client, dir string) error {
	if dir == "/" {
		// Must always exist
		return nil
	}

	stat, err := sftpClient.Lstat(dir)
	if err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("not a directory: %q", dir)
		}
		return nil
	}

	parent := path.Dir(dir)
	err = mkdirAll(sftpClient, parent)
	if err != nil {
		return err
	}

	err = sftpClient.Mkdir(dir)
	if err != nil {
		return fmt.Errorf("error creating directory %q over sftp: %w", dir, err)
	}
	return nil
}

func (p *SSHPath) WriteFile(ctx context.Context, data io.ReadSeeker, acl ACL) error {
	sftpClient, err := p.newClient(ctx)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	dir := path.Dir(p.path)
	err = mkdirAll(sftpClient, dir)
	if err != nil {
		return err
	}

	tempfile := path.Join(dir, fmt.Sprintf(".tmp-%d", rand.Int63()))
	f, err := sftpClient.Create(tempfile)
	if err != nil {
		// TODO: Retry if concurrently created?
		return fmt.Errorf("error creating temp file in %q: %w", dir, err)
	}

	// Note from here on in we have to close f and delete or rename the temp file

	_, err = io.Copy(f, data)

	if closeErr := f.Close(); err == nil {
		err = closeErr
	}

	if err == nil {
		if acl != nil {
			sshACL, ok := acl.(*SSHAcl)
			if !ok {
				err = fmt.Errorf("unexpected acl type %T", acl)
			} else {
				err = sftpClient.Chmod(tempfile, sshACL.Mode)
				if err != nil {
					err = fmt.Errorf("error during chmod of %q: %w", tempfile, err)
				}
			}
		}
	}

	if err == nil {
		// posix rename will replace the destination (normal sftp rename does not)
		usePosixRename := true
		if usePosixRename {
			err = sftpClient.Rename(tempfile, p.path)
			if err != nil {
				err = fmt.Errorf("error renaming file %q -> %q (with posix rename): %w", tempfile, p.path, err)
			}
		} else {
			var session *ssh.Session
			session, err = p.client.NewSession()
			if err != nil {
				err = fmt.Errorf("error creating session for rename: %w", err)
			} else {
				defer session.Close()

				cmd := "mv " + tempfile + " " + p.path
				if p.sudo {
					cmd = "sudo " + cmd
				}
				err = session.Run(cmd)
				if err != nil {
					err = fmt.Errorf("error renaming file %q -> %q (with %q): %w", tempfile, p.path, cmd, err)
				}
			}
		}

	}

	if err == nil {
		return nil
	}

	// Something went wrong; try to remove the temp file
	if removeErr := sftpClient.Remove(tempfile); removeErr != nil {
		klog.Warningf("unable to remove temp file %q: %v", tempfile, removeErr)
	}

	return err
}

// To prevent concurrent creates on the same file while maintaining atomicity of writes,
// we take a process-wide lock during the operation.
// Not a great approach, but fine for a single process (with low concurrency)
var createFileLockSSH sync.Mutex

func (p *SSHPath) CreateFile(ctx context.Context, data io.ReadSeeker, acl ACL) error {
	createFileLockSSH.Lock()
	defer createFileLockSSH.Unlock()

	// Check if exists
	_, err := p.ReadFile(ctx)
	if err == nil {
		return os.ErrExist
	}

	if !os.IsNotExist(err) {
		return err
	}

	return p.WriteFile(ctx, data, acl)
}

// ReadFile implements Path::ReadFile
func (p *SSHPath) ReadFile(ctx context.Context) ([]byte, error) {
	var b bytes.Buffer
	_, err := p.WriteTo(&b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// WriteTo reads the file (in a streaming way)
// This implements io.WriterTo
func (p *SSHPath) WriteTo(out io.Writer) (int64, error) {
	ctx := context.TODO()

	sftpClient, err := p.newClient(ctx)
	if err != nil {
		return 0, fmt.Errorf("error creating sftp client: %w", err)
	}
	defer sftpClient.Close()

	f, err := sftpClient.Open(p.path)
	if err != nil {
		return 0, fmt.Errorf("error opening file %s over sftp: %w", p, err)
	}
	defer f.Close()

	return f.WriteTo(out)
}

func (p *SSHPath) ReadDir() ([]Path, error) {
	ctx := context.TODO()

	sftpClient, err := p.newClient(ctx)
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()

	files, err := sftpClient.ReadDir(p.path)
	if err != nil {
		return nil, err
	}
	var children []Path
	for _, f := range files {
		child := NewSSHPath(p.client, p.server, path.Join(p.path, f.Name()), p.sudo)

		children = append(children, child)
	}
	return children, nil
}

func (p *SSHPath) ReadTree() ([]Path, error) {
	ctx := context.TODO()

	sftpClient, err := p.newClient(ctx)
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()

	var paths []Path
	err = readSFTPTree(sftpClient, p, &paths)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func readSFTPTree(sftpClient *sftp.Client, p *SSHPath, dest *[]Path) error {
	files, err := sftpClient.ReadDir(p.path)
	if err != nil {
		return err
	}
	for _, f := range files {
		child := NewSSHPath(p.client, p.server, path.Join(p.path, f.Name()), p.sudo)

		*dest = append(*dest, child)

		if f.IsDir() {
			err = readSFTPTree(sftpClient, child, dest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *SSHPath) Base() string {
	return path.Base(p.path)
}
