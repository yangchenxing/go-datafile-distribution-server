package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestSaveAndBackup(t *testing.T) {
	f := &file{
		path:      "data/test.dat",
		backupDir: "backup",
		backups:   1,
		filename:  "test.dat",
	}
	// test save
	content := []byte("test content")
	if err := f.save(content); err != nil {
		t.Error("save file fail:", err.Error())
		return
	}
	defer os.Remove("data/test.dat")
	if !bytes.Equal(f.content, content) {
		t.Errorf("unexpected content: expected=%q, actual=%q", string(content), string(f.content))
		return
	}
	// test backup
	if err := f.backup(); err != nil {
		t.Errorf("backup file fail:", err.Error())
		return
	}
	backupPath := "backup/test.dat." + f.etag
	if backupContent, err := ioutil.ReadFile(backupPath); err != nil {
		t.Error("read backup file fail:", err.Error())
		return
	} else if !bytes.Equal(backupContent, content) {
		t.Errorf("unexpected backup file content: expected=%q, actual=%q", string(content), string(backupContent))
		return
	}
}

func TestMain(m *testing.M) {
	os.Mkdir("data", 0755)
	os.Mkdir("backup", 0755)
	code := m.Run()
	os.Remove("backup")
	os.Remove("data")
	os.Exit(code)
}
