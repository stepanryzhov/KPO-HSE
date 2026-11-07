package main

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed demo/students.json
var demoStudents []byte

//go:embed demo/gradebook.json
var demoGradebook []byte

func prepareDemo() (string, error) {
	directory, err := os.MkdirTemp("", "pos-demo-")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(directory, "gradebooks"), 0700); err != nil {
		os.RemoveAll(directory)
		return "", err
	}
	for name, data := range map[string][]byte{
		"students.json": demoStudents,
		filepath.Join("gradebooks", "default.json"): demoGradebook,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), data, 0600); err != nil {
			os.RemoveAll(directory)
			return "", err
		}
	}
	return directory, nil
}
