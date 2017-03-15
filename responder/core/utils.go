package responder

// Request and response info structs, utility functions and methods
// Uses boltdb for

import (
	"crypto/rand"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"strings"
)

// for logging
var logger *log.Logger

func info(args ...interface{}) {
	logger.SetPrefix("INFO [" + ROUTEID + "] - ")
	logger.Println(args...)
}

func danger(args ...interface{}) {
	logger.SetPrefix("ERROR [" + ROUTEID + "] - ")
	logger.Println(args...)
}

func warning(args ...interface{}) {
	logger.SetPrefix("WARNING [" + ROUTEID + "] - ")
	logger.Println(args...)
}

type RequestInfo struct {
	Method           string                 `json:"Method"`
	URL              URLInfo                `json:"URL"`
	Proto            string                 `json:"Proto"`
	Header           map[string][]string    `json:"Header"`
	Body             interface{}            `json:"Body"`
	ContentLength    int64                  `json:"ContentLength"`
	TransferEncoding []string               `json:"TransferEncoding"`
	Host             string                 `json:"Host"`
	Params           map[string][]string    `json:"Params"`
	Multipart        map[string][]Multipart `json:"Multipart"`
	RemoteAddr       string                 `json:"RemoteAddr"`
	RequestURI       string                 `json:"RequestURI"`
}

type Multipart struct {
	Filename    string `json:"Filename"`
	ContentType string `json:"ContentType"`
	Content     string `json:"Content"`
}

type URLInfo struct {
	Scheme   string `json:"Scheme"`
	Opaque   string `json:"Opaque"`
	Host     string `json:"Host"`
	Path     string `json:"Path"`
	RawQuery string `json:"RawQuery"`
	Fragment string `json:"Fragment"`
}

type ResponseInfo struct {
	Status string              `json:"status"`
	Header map[string][]string `json:"header"`
	Body   string              `json:"body"`
}

// methods on RequestInfo

func (req *RequestInfo) GetHeader(key string) (value string) {
	if req.Header[key] != nil && len(req.Header[key]) > 0 {
		value = req.Header[key][0]
	} else {
		value = ""
	}
	return
}

func (req *RequestInfo) GetCookie(key string) (value string) {
	list := req.GetHeader("Cookie")
	cookies := strings.Split(list, ";")
	for _, c := range cookies {
		kv := strings.SplitN(c, "=", 2)
		if key == strings.TrimSpace(kv[0]) {
			return kv[1]
		}
	}
	return
}

func (req *RequestInfo) GetParam(key string) (value string) {
	if req.Params[key] != nil && len(req.Params[key]) > 0 {
		value = req.Params[key][0]
	} else {
		value = ""
	}
	return
}

// methods on ResponseInfo

func (resp *ResponseInfo) AddHeader(key string, value string) {
	if resp.Header[key] == nil {
		resp.Header[key] = []string{}
	}
	resp.Header[key] = append(resp.Header[key], value)
}

func (resp *ResponseInfo) SetCookie(key string, value string) {
	resp.AddHeader("Set-Cookie", key+"="+value)
}

func (resp *ResponseInfo) RedirectTo(url string) {
	resp.Status = "302"
	resp.AddHeader("Location", url)
}

func (resp *ResponseInfo) SetHTML() {
	response.AddHeader("Content-Type", "text/html; charset=utf-8")
}

func (resp *ResponseInfo) SetJSON() {
	response.AddHeader("Content-Type", "application/json; charset=utf-8")
}

// embedded db methods
// store a byte array with a bucket and a key
func Store(bucket string, key string, value []byte) (err error) {
	db, err := bolt.Open("red.db", 0600, nil)
	defer db.Close()
	if err != nil {
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(key), value)
		if err != nil {
			return err
		}
		return err
	})
	return
}

// get a byte array with a bucket and a key
func Get(bucket string, key string) (value string, err error) {
	db, err := bolt.Open("red.db", 0600, nil)
	defer db.Close()
	if err != nil {
		return
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("Bucket %q not found!", bucket)
		}
		value = string(b.Get([]byte(key)))
		return nil
	})
	return
}

func Delete(bucket string, key string) (err error) {
	db, err := bolt.Open("red.db", 0600, nil)
	defer db.Close()
	if err != nil {
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("Bucket %q not found!", bucket)
		}
		b.Delete([]byte(key))
		return nil
	})
	return
}

// create a random UUID with from RFC 4122
// adapted from http://github.com/nu7hatch/gouuid
func CreateUUID() (uuid string) {
	u := new([16]byte)
	_, err := rand.Read(u[:])
	if err != nil {
		log.Fatalf("%s: %s", "Cannot generate UUID", err)
		panic(fmt.Sprintf("%s: %s", "Cannot generate UUID", err))
	}

	// 0x40 is reserved variant from RFC 4122
	u[8] = (u[8] | 0x40) & 0x7F
	// Set the four most significant bits (bits 12 through 15) of the
	// time_hi_and_version field to the 4-bit version number.
	u[6] = (u[6] & 0xF) | (0x4 << 4)
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
	return
}
