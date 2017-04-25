package db

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

type DB interface {
	Load(string, interface{}) (bool, error)
	Save(string, interface{}) error
}

type BoltDB struct {
	db *bolt.DB
}

var (
	versionIdent       = []byte("version")
	persistenceVersion = []byte{1, 0} // major.minor
	topBucket          = []byte("top")
)

const (
	NameIdent = "peername"
	fileName  = "data.db"
)

func Pathname(dbPrefix string) string {
	return dbPrefix + fileName
}

func NewBoltDB(dbPrefix string) (*BoltDB, error) {
	dbPathname := Pathname(dbPrefix)
	db, err := bolt.Open(dbPathname, 0660, nil)
	if err != nil {
		return nil, fmt.Errorf("[boltDB] Unable to open %s: %s", dbPathname, err)
	}
	err = db.Update(checkVersion(false))
	return &BoltDB{db: db}, err
}

func NewBoltDBReadOnly(dbPrefix string) (*BoltDB, error) {
	options := bolt.Options{Timeout: time.Millisecond * 50, ReadOnly: true}
	dbPathname := Pathname(dbPrefix)
	db, err := bolt.Open(dbPathname, 0660, &options)
	if err != nil {
		return nil, fmt.Errorf("[boltDB] Unable to open %s: %s", dbPathname, err)
	}
	err = db.View(checkVersion(true))
	if err != nil {
		return nil, fmt.Errorf("[boltDB] Cannot use persistence file %s: %s", dbPathname, err)
	}
	return &BoltDB{db: db}, nil
}

func checkVersion(readOnly bool) func(tx *bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		if top := tx.Bucket(topBucket); top == nil {
			if readOnly {
				return fmt.Errorf("no top bucket")
			}
			top, err := tx.CreateBucket(topBucket)
			if err != nil {
				return err
			}
			if err := top.Put(versionIdent, persistenceVersion); err != nil {
				return err
			}
		} else {
			if checkVersion := top.Get(versionIdent); checkVersion != nil {
				if checkVersion[0] != persistenceVersion[0] {
					return fmt.Errorf("mismatched version %x", checkVersion)
				}
			}
		}
		return nil
	}
}

func (d *BoltDB) Load(ident string, data interface{}) (bool, error) {
	found := true
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(topBucket)
		v := b.Get([]byte(ident))
		if v == nil {
			found = false
			return nil
		}
		reader := bytes.NewReader(v)
		decoder := gob.NewDecoder(reader)
		err := decoder.Decode(data)
		return err
	})
	return found, err
}

func (d *BoltDB) Save(ident string, data interface{}) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(topBucket)
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(data); err != nil {
			return err
		}
		return b.Put([]byte(ident), buf.Bytes())
	})
}

func (d *BoltDB) Close() error {
	return d.db.Close()
}
