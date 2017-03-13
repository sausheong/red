package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
)

var logger *log.Logger

type RequestInfo struct {
	Method           string                 `json:"Method"`
	URL              URLInfo                `json:"URL"`
	Proto            string                 `json:"Proto"`
	Header           map[string][]string    `json:"Header"`
	Body             io.ReadCloser          `json:"Body"`
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

func init() {
	file, err := os.OpenFile("acceptor.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	logger = log.New(file, "INFO ", log.Ldate|log.Ltime|log.Lshortfile)
}

// create a random UUID with from RFC 4122
// adapted from http://github.com/nu7hatch/gouuid
func createUUID() (uuid string) {
	u := new([16]byte)
	_, err := rand.Read(u[:])
	danger("Cannot generate UUID", err)
	// 0x40 is reserved variant from RFC 4122
	u[8] = (u[8] | 0x40) & 0x7F
	// Set the four most significant bits (bits 12 through 15) of the
	// time_hi_and_version field to the 4-bit version number.
	u[6] = (u[6] & 0xF) | (0x4 << 4)
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
	return
}

// for logging

func info(args ...interface{}) {
	logger.SetPrefix("INFO ")
	logger.Println(args...)
}

func danger(args ...interface{}) {
	logger.SetPrefix("ERROR ")
	logger.Println(args...)
}

func warning(args ...interface{}) {
	logger.SetPrefix("WARNING ")
	logger.Println(args...)
}
