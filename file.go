package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultBackups = 5
)

type file struct {
	sync.RWMutex
	path         string
	backupDir    string
	backups      uint
	filename     string
	content      []byte
	md5          [md5.Size]byte
	etag         string
	lastModified time.Time
}

func (f *file) save(content []byte) error {
	tempPath := f.path + ".tmp"
	if err := ioutil.WriteFile(tempPath, content, 0755); err != nil {
		return fmt.Errorf("save temp file fail: %s", err.Error())
	}
	if err := f.backup(); err != nil {
		return fmt.Errorf("backup fail: %s", err.Error())
	}
	if err := os.Rename(tempPath, f.path); err != nil {
		return fmt.Errorf("rename temp file %q to formal file %q fail: %s", tempPath, f.path, err.Error())
	}
	info, err := os.Stat(f.path)
	if err != nil {
		return fmt.Errorf("stat data file %q fail: %s", f.path, err.Error())
	}
	f.content = content
	f.md5 = md5.Sum(content)
	f.etag = fmt.Sprintf("%x", f.md5)
	f.lastModified = info.ModTime()
	return nil
}

func (f *file) backup() error {
	if f.backups == 0 {
		return nil
	}
	infos, err := ioutil.ReadDir(f.backupDir)
	for err != nil {
		return fmt.Errorf("read dir %q fail: %s", f.backupDir, err.Error())
	}
	sort.Sort(sort.Reverse(fileInfoSlice(infos)))
	backups := f.backups - 1
	for _, info := range infos {
		if strings.HasPrefix(info.Name(), f.filename) && len(info.Name()) == len(f.filename)+65 {
			if backups > 0 {
				backups--
			} else {
				os.Remove(filepath.Join(f.backupDir, info.Name()))
			}
		}
	}
	backupPath := filepath.Join(f.backupDir, f.filename+"."+f.etag)
	if err := ioutil.WriteFile(backupPath, f.content, 0755); err != nil {
		return fmt.Errorf("write backup file %q fail: %s", backupPath, err.Error())
	}
	return nil
}
