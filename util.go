package main

import (
	"os"
)

type fileInfoSlice []os.FileInfo

func (s fileInfoSlice) Len() int {
	return len(s)
}

func (s fileInfoSlice) Less(i, j int) bool {
	return s[i].ModTime().Before(s[j].ModTime())
}

func (s fileInfoSlice) Swap(i, j int) {
	temp := s[i]
	s[i] = s[j]
	s[j] = temp
}
