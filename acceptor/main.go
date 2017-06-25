package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/nats-io/go-nats"
	"io"
	"net/http"
	"strconv"
	"strings"
	// "os"
	"time"
)

func main() {
	router := httprouter.New()

	addr := "0.0.0.0:8080"
	router.GET("/_/*p", accept)
	router.POST("/_/*p", accept)

	router.ServeFiles("/_s/*filepath", http.Dir("../responder/bin/public"))
	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    time.Duration(10 * int64(time.Second)),
		WriteTimeout:   time.Duration(10 * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("Polyglot Acceptor", version(), "started at", addr)
	server.ListenAndServe()
}

// default handler
func accept(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {

	// the multipart contains the multipart data
	multipart := make(map[string][]Multipart)

	// if request is a POST, parse the multipartform for stuff in the forms
	if request.Method == "POST" {
		request.ParseMultipartForm(3 << 20)
		if request.MultipartForm != nil {
			for mk, mv := range request.MultipartForm.File {
				var parts []Multipart
				for _, v := range mv {
					f, err := v.Open()
					if err != nil {
						danger("Cannot read multipart message", err)
					}
					var buf bytes.Buffer
					_, err = io.Copy(&buf, f)
					if err != nil {
						danger("Cannot copy multipart message into buffer", err)
					}
					content := base64.StdEncoding.EncodeToString(buf.Bytes())
					part := Multipart{
						Filename:    v.Filename,
						ContentType: v.Header["Content-Type"][0],
						Content:     content,
					}
					parts = append(parts, part)
				}
				multipart[mk] = parts
			}
		}
	}

	// the form contains data from the URL as well as the POST form
	params := make(map[string][]string)
	err := request.ParseForm()
	if err != nil {
		danger("Failed to parse form", err)
	}

	for fk, fv := range request.Form {
		params[fk] = fv
	}

	reqInfo := RequestInfo{
		Method: request.Method,
		URL: URLInfo{
			Scheme:   request.URL.Scheme,
			Opaque:   request.URL.Opaque,
			Host:     request.URL.Host,
			Path:     request.URL.Path,
			RawQuery: request.URL.RawQuery,
			Fragment: request.URL.Fragment,
		},
		Proto:            request.Proto,
		Header:           request.Header,
		Body:             request.Body,
		ContentLength:    request.ContentLength,
		TransferEncoding: request.TransferEncoding,
		Host:             request.Host,
		Params:           params,
		Multipart:        multipart,
		RemoteAddr:       request.RemoteAddr,
		RequestURI:       request.RequestURI,
	}

	// marshal the RequestInfo struct into JSON
	reqJson, err := json.Marshal(reqInfo)
	if err != nil {
		danger("Failed to marshal the request into JSON", err)
	}
	routeId := request.Method + request.URL.Path

	// send request

	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		danger("Cannot connect to NATS server", err)
	}
	defer conn.Close()

	response, err := conn.Request(routeId, reqJson, 1*time.Second)

	if err := conn.LastError(); err != nil {
		danger("Cannot send request to NATS server", err)
	} else {
		info("Sent request to NATS server")
	}

	// set the identity to the route ID eg GET/_/path
	respInfo := ResponseInfo{}
	err = json.Unmarshal(response.Data, &respInfo)
	if err != nil {
		danger("Failed to unmarshal the response JSON into ResponseInfo", err)
	}

	// write headers
	for k, v := range respInfo.Header {
		for _, val := range v {
			writer.Header().Add(k, val)
		}
	}

	// get status
	status, err := strconv.Atoi(respInfo.Status)
	if err != nil {
		reply(writer, 500, []byte(err.Error()))
	}

	var data []byte
	// get content type
	ctype, hasCType := respInfo.Header["Content-Type"]
	if hasCType == true {
		if is_text_mime_type(ctype[0]) {
			data = []byte(respInfo.Body)
		} else {
			data, _ = base64.StdEncoding.DecodeString(respInfo.Body)
		}
	} else {
		data = []byte(respInfo.Body) // if not given the content type, assume it's text
	}
	// write status and body to response
	reply(writer, status, data)

}

func reply(writer http.ResponseWriter, status int, body []byte) {
	writer.WriteHeader(status)
	writer.Write(body)
}

func is_text_mime_type(ctype string) bool {
	if strings.HasPrefix(ctype, "text") ||
		strings.HasPrefix(ctype, "application/json") {
		return true
	} else {
		return false
	}

}
