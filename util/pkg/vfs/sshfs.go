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

func (p *SSHPath) newClient() (*sftp.Client, error) {
	if !p.sudo {
		sftpClient, err := sftp.NewClient(p.client)
		if err != nil {
			return nil, fmt.Errorf("error creating sftp client: %v", err)
		}

		return sftpClient, nil
	}
	s, err := p.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("error creating sftp client (in new-session): %v", err)
	}

	//if err := s.R("sftp"); err != nil {
	//	return nil, fmt.Errorf("error creating sftp client (in new-session): %v", err)
	//}
	stdin, err := s.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating sftp client (at stdin pipe): %v", err)
	}
	stdout, err := s.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating sftp client (at stdout pipe): %v", err)
	}

	err = s.Start("sudo /usr/lib/openssh/sftp-server")
	if err != nil {
		return nil, fmt.Errorf("error creating sftp client (executing 'sudo /usr/lib/openssh/sftp-server'): %v", err)
	}

	return sftp.NewClientPipe(stdout, stdin)
}
func (p *SSHPath) Path() string {
	return "ssh://" + p.server + p.path
}

func (p *SSHPath) String() string {
	return p.Path()
}

func (p *SSHPath) Remove() error {
	sftpClient, err := p.newClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	err = sftpClient.Remove(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("error deleting %s: %v", p, err)
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
		return fmt.Errorf("error creating directory %q over sftp: %v", dir, err)
	}
	return nil
}

func (p *SSHPath) WriteFile(data io.ReadSeeker, acl ACL) error {
	sftpClient, err := p.newClient()
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
		return fmt.Errorf("error creating temp file in %q: %v", dir, err)
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
					err = fmt.Errorf("error during chmod of %q: %v", tempfile, err)
				}
			}
		}
	}

	if err == nil {
		var session *ssh.Session
		session, err = p.client.NewSession()
		if err != nil {
			err = fmt.Errorf("error creating session for rename: %v", err)
		} else {
			cmd := "mv " + tempfile + " " + p.path
			if p.sudo {
				cmd = "sudo " + cmd
			}
			err = session.Run(cmd)
			if err != nil {
				err = fmt.Errorf("error renaming file %q -> %q: %v", tempfile, p.path, err)
			}
		}
		// sftp rename seems to fail if dest file exists
		//err = sftpClient.Rename(tempfile, p.path)
		//if err != nil {
		//	err = fmt.Errorf("error during file write of %q: rename failed: %v", p.path, err)
		//}
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

func (p *SSHPath) CreateFile(data io.ReadSeeker, acl ACL) error {
	createFileLockSSH.Lock()
	defer createFileLockSSH.Unlock()

	// Check if exists
	_, err := p.ReadFile()
	if err == nil {
		return os.ErrExist
	}

	if !os.IsNotExist(err) {
		return err
	}

	return p.WriteFile(data, acl)
}

// ReadFile implements Path::ReadFile
func (p *SSHPath) ReadFile() ([]byte, error) {
	var b bytes.Buffer
	_, err := p.WriteTo(&b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// WriteTo implements io.WriterTo
func (p *SSHPath) WriteTo(out io.Writer) (int64, error) {
	sftpClient, err := p.newClient()
	if err != nil {
		return 0, err
	}
	defer sftpClient.Close()

	f, err := sftpClient.Open(p.path)
	if err != nil {
		return 0, fmt.Errorf("error opening file %s over sftp: %v", p, err)
	}
	defer f.Close()

	return f.WriteTo(out)
}

func (p *SSHPath) ReadDir() ([]Path, error) {
	sftpClient, err := p.newClient()
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
	sftpClient, err := p.newClient()
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

//
//// scpMkdir executes a mkdir against the SSH target, using SCP
//func (s *SSHPath) scpMkdir(dest string, mode os.FileMode) error {
//	klog.V(4).Infof("Doing SSH SCP mkdir: %q", dest)
//	session, err := s.client.NewSession()
//	if err != nil {
//		return fmt.Errorf("error establishing SSH session: %v", err)
//	}
//	defer session.Close()
//
//	name := path.Base(dest)
//	scpBase := path.Dir(dest)
//	//scpBase = "." + scpBase
//
//	var stdinErr error
//	go func() {
//		w, _ := session.StdinPipe()
//		defer w.Close()
//		_, stdinErr = fmt.Fprintln(w, "D0"+toOctal(mode), 0, name)
//		if stdinErr != nil {
//			return
//		}
//	}()
//	output, err := session.CombinedOutput("/usr/bin/scp -tr " + scpBase)
//	if err != nil {
//		klog.Warningf("Error output from SCP: %s", output)
//		return fmt.Errorf("error doing SCP mkdir: %v", err)
//	}
//	if stdinErr != nil {
//		klog.Warningf("Error output from SCP: %s", output)
//		return fmt.Errorf("error doing SCP mkdir (writing to stdin): %v", stdinErr)
//	}
//
//	return nil
//}
//
//func toOctal(mode os.FileMode) string {
//	return strconv.FormatUint(uint64(mode), 8)
//}
//
//// scpPut copies a file to the SSH target, using SCP
//func (s *SSHPath) scpPut(dest string, length int, content io.Reader, mode os.FileMode) error {
//	klog.V(4).Infof("Doing SSH SCP upload: %q", dest)
//	session, err := s.client.NewSession()
//	if err != nil {
//		return fmt.Errorf("error establishing SSH session: %v", err)
//	}
//	defer session.Close()
//
//	name := path.Base(dest)
//	scpBase := path.Dir(dest)
//	//scpBase = "." + scpBase
//
//	var stdinErr error
//	go func() {
//		w, _ := session.StdinPipe()
//		defer w.Close()
//		_, stdinErr = fmt.Fprintln(w, "C0"+toOctal(mode), length, name)
//		if stdinErr != nil {
//			return
//		}
//		_, stdinErr = io.Copy(w, content)
//		if stdinErr != nil {
//			return
//		}
//		_, stdinErr = fmt.Fprint(w, "\x00")
//		if stdinErr != nil {
//			return
//		}
//	}()
//	output, err := session.CombinedOutput("/usr/bin/scp -tr " + scpBase)
//	if err != nil {
//		klog.Warningf("Error output from SCP: %s", output)
//		return fmt.Errorf("error doing SCP put: %v", err)
//	}
//	if stdinErr != nil {
//		klog.Warningf("Error output from SCP: %s", output)
//		return fmt.Errorf("error doing SCP put (writing to stdin): %v", stdinErr)
//	}
//
//	return nil
//}
