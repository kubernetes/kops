package ipset

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Name string

type Type string

const (
	ListSet = Type("list:set")
	HashIP  = Type("hash:ip")
)

type Interface interface {
	Create(ipsetName Name, ipsetType Type) error
	AddEntry(ipsetName Name, entry string) error
	DelEntry(ipsetName Name, entry string) error
	Flush(ipsetName Name) error
	Destroy(ipsetName Name) error

	List(prefix string) ([]Name, error)

	FlushAll() error
	DestroyAll() error
}

type ipset struct {
	refCount
}

func New() Interface {
	return &ipset{refCount: newRefCount()}
}

func (i *ipset) Create(ipsetName Name, ipsetType Type) error {
	return doExec("create", string(ipsetName), string(ipsetType))
}

func (i *ipset) AddEntry(ipsetName Name, entry string) error {
	if i.inc(ipsetName, entry) > 1 { // already in the set
		return nil
	}
	return doExec("add", string(ipsetName), entry)
}

func (i *ipset) DelEntry(ipsetName Name, entry string) error {
	if i.dec(ipsetName, entry) > 0 { // still needed
		return nil
	}
	return doExec("del", string(ipsetName), entry)
}

func (i *ipset) Flush(ipsetName Name) error {
	i.removeSet(ipsetName)
	return doExec("flush", string(ipsetName))
}

func (i *ipset) FlushAll() error {
	i.refCount = newRefCount()
	return doExec("flush")
}

func (i *ipset) Destroy(ipsetName Name) error {
	i.removeSet(ipsetName)
	return doExec("destroy", string(ipsetName))
}

func (i *ipset) DestroyAll() error {
	i.refCount = newRefCount()
	return doExec("destroy")
}

// Fetch a list of all existing sets with a given prefix
func (i *ipset) List(prefix string) ([]Name, error) {
	output, err := exec.Command("ipset", "list", "-name", "-output", "plain").Output()
	if err != nil {
		return nil, err
	}

	var selected []Name
	sets := strings.Split(string(output), "\n")
	for _, v := range sets {
		if strings.HasPrefix(v, prefix) {
			selected = append(selected, Name(v))
		}
	}

	return selected, err
}

func doExec(args ...string) error {
	if output, err := exec.Command("ipset", args...).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "ipset %v failed: %s", args, output)
	}
	return nil
}

// Reference-counting
type key struct {
	ipsetName Name
	entry     string
}

// note no locking is required as all operations are serialised in the controller
type refCount struct {
	ref map[key]int
}

func newRefCount() refCount {
	return refCount{ref: make(map[key]int)}
}

func (rc *refCount) inc(ipsetName Name, entry string) int {
	k := key{ipsetName, entry}
	rc.ref[k]++
	return rc.ref[k]
}

func (rc *refCount) dec(ipsetName Name, entry string) int {
	k := key{ipsetName, entry}
	rc.ref[k]--
	return rc.ref[k]
}

func (rc *refCount) removeSet(ipsetName Name) {
	for k := range rc.ref {
		if k.ipsetName == ipsetName {
			delete(rc.ref, k)
		}
	}
}
