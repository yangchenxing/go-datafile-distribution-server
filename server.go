package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type server struct {
	dataPath   string
	backupPath string
	backups    uint
	files      map[string]*file
}

func (dds *server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	name := strings.Trim(req.URL.Path, "/")
	if strings.Contains(name, "/") {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	switch req.Method {
	case "GET":
		dds.serveGet(name, resp, req)
	case "POST":
		dds.servePost(name, resp, req)
	case "HEAD":
		dds.serveHead(name, resp, req)
	}
}

func (dds *server) serveHead(name string, resp http.ResponseWriter, req *http.Request) {
	file := dds.files[name]
	if file == nil {
		if file = dds.loadFile(name); file == nil {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
	}
	file.RLock()
	defer file.RUnlock()
	header := resp.Header()
	header.Set("ETag", file.etag)
	header.Set("Last-Modified", file.lastModified.Format("Mon, 2 Jan 2006 15:04:05 MST"))
	if expectedTag := req.Header.Get("If-None-Match"); expectedTag != "" {
		if expectedTag == file.etag {
			resp.WriteHeader(http.StatusNotModified)
			return
		}
	}
	if expectedModifiedTime := req.Header.Get("If-Modified-Since"); expectedModifiedTime != "" {
		modTime, err := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", expectedModifiedTime)
		if err == nil && !modTime.Before(file.lastModified) {
			resp.WriteHeader(http.StatusNotModified)
			return
		}
	}
	resp.WriteHeader(http.StatusNoContent)
}

func (dds *server) serveGet(name string, resp http.ResponseWriter, req *http.Request) {
	file := dds.files[name]
	if file == nil {
		if file = dds.loadFile(name); file == nil {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
	}
	file.RLock()
	defer file.RUnlock()
	header := resp.Header()
	header.Set("ETag", fmt.Sprintf("%x", file.md5))
	header.Set("Last-Modified", file.lastModified.Format("Mon, 2 Jan 2006 15:04:05 MST"))
	if expectedTag := req.Header.Get("If-None-Match"); expectedTag != "" {
		if expectedTag == file.etag {
			resp.WriteHeader(http.StatusNotModified)
			return
		}
	}
	if expectedModifiedTime := req.Header.Get("If-Modified-Since"); expectedModifiedTime != "" {
		modTime, err := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", expectedModifiedTime)
		if err == nil && !modTime.Before(file.lastModified) {
			resp.WriteHeader(http.StatusNotModified)
			return
		}
	}
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.filename))
	header.Set("Content-Length", strconv.Itoa(len(file.content)))
	header.Set("Content-MD5", base64.StdEncoding.EncodeToString(file.md5[:]))
	resp.Write(file.content)
}

func (dds *server) servePost(name string, resp http.ResponseWriter, req *http.Request) {
	file := dds.files[name]
	if file == nil {
		file = dds.newFile(name)
	}
	content, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	file.Lock()
	defer file.Unlock()
	if err := file.save(content); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp.WriteHeader(http.StatusNoContent)
}

func (dds *server) newFile(name string) *file {
	file := &file{
		path:      filepath.Join(dds.dataPath, name),
		backupDir: dds.backupPath,
		backups:   dds.backups,
		filename:  name,
	}
	dds.files[name] = file
	return file
}

func (dds *server) loadFile(name string) *file {
	path := filepath.Join(dds.dataPath, name)
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	file := &file{
		path:         path,
		backupDir:    dds.backupPath,
		backups:      dds.backups,
		filename:     name,
		content:      content,
		md5:          md5.Sum(content),
		lastModified: info.ModTime(),
	}
	file.etag = fmt.Sprintf("%x", file.md5)
	dds.files[name] = file
	return file
}
