package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestServeHead(t *testing.T) {
	resp := httptest.NewRecorder()
	s = &server{
		dataPath:   "data",
		backupPath: "backup",
		backups:    1,
		files:      make(map[string]*file),
	}
	content := []byte("test content")
	if err := ioutil.WriteFile("data/test.dat", content, 0755); err != nil {
		t.Error("write test.dat fail:", err.Error())
		return
	}
	req := &http.Request{}
}

func TestServeGet(t *testing.T) {

}

func TestServePost(t *testing.T) {

}

func TestMain(m *testing.M) {
	os.Mkdir("data")
	os.Mkdir("backup")
	code = m.Run()
	os.Remove("data")
	os.Remote("backup")
	os.Exit(code)
}
