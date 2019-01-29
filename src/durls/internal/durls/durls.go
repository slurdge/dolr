package durls

import (
	"github.com/martinlindhe/base36"
	"log"
	"strings"

	skip32 "github.com/dgryski/go-skip32"

	"encoding/binary"
	"errors"
	"github.com/etcd-io/bbolt"
)

var gdb *bbolt.DB
var bucketName = []byte("urls")
var key = []byte("0123456789")
var obfsucator *skip32.Skip32
var errNotFound = errors.New("durls: not found")

// OpenDB Opens the main DB
func OpenDB(filename string) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bbolt.Open(filename, 0600, nil)
	gdb = db
	if err != nil {
		log.Fatal(err)
	}
	obfsucator, err = skip32.New(key)
	if err != nil {
		log.Fatal(err)
	}
	err = gdb.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()
}

// Shorten shortens URL
func Shorten(URL string) string {
	var id32o uint32
	gdb.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		id, _ := bucket.NextSequence()
		id32 := uint32(id)
		id32o = obfsucator.Obfus(id32)
		array := make([]byte, 4)
		binary.LittleEndian.PutUint32(array, id32)
		return bucket.Put(array, []byte(URL))
	})
	return strings.ToLower(base36.Encode(uint64(id32o)))
}

// Lookup lookups URL
func Lookup(shortURL string) (string, error) {
	var url string
	var err error
	gdb.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		id32o := uint32(base36.Decode(shortURL))
		id32 := obfsucator.Unobfus(id32o)
		array := make([]byte, 4)
		binary.LittleEndian.PutUint32(array, id32)
		urlBytes := bucket.Get(array)
		if urlBytes != nil {
			url = string(urlBytes)
		} else {
			err = errNotFound
		}
		return nil
	})
	return url, err
}
