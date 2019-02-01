package dorl

import (
	"log"
	"net/url"
	"strings"

	"github.com/martinlindhe/base36"

	skip32 "github.com/dgryski/go-skip32"

	"encoding/binary"
	"errors"

	"github.com/etcd-io/bbolt"
)

// Session The main holder of information for current DB session
type Session struct {
	db         *bbolt.DB
	obfuscator *skip32.Skip32
}

var bucketName = []byte("urls")
var errNotFound = errors.New("dorl: not found")
var errInvalid = errors.New("dorl: invalid URL")

// OpenSession Opens the main session
func OpenSession(dbName string, obsKey []byte) *Session {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	session := new(Session)
	db, err := bbolt.Open(dbName, 0600, nil)
	session.db = db
	if err != nil {
		log.Fatal(err)
	}
	obfuscator, err := skip32.New(obsKey)
	if err != nil {
		log.Fatal(err)
	}
	session.obfuscator = obfuscator
	err = session.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	return session
	//defer db.Close()
}

// Shorten shortens URL
func (session *Session) Shorten(URL string) (string, error) {
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return "", err
	}
	var id32o uint32
	session.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		id, _ := bucket.NextSequence()
		id32 := uint32(id)
		id32o = session.obfuscator.Obfus(id32)
		array := make([]byte, 4)
		binary.LittleEndian.PutUint32(array, id32)
		return bucket.Put(array, []byte(URL))
	})
	return strings.ToLower(base36.Encode(uint64(id32o))), nil
}

// Lookup lookups URL
func (session *Session) Lookup(shortURL string) (string, error) {
	var url string
	var err error
	session.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		id32o := uint32(base36.Decode(shortURL))
		id32 := session.obfuscator.Unobfus(id32o)
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
