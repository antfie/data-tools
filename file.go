package main

import (
	"os"
	"os/exec"
	"strings"
)

func IsDir(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}

	return false
}

func IsFile(path string) bool {
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return true
	}

	return false
}

func GetTypeOfFile(file string) (string, error) {
	command := exec.Command("file", "-b", "--mime-type", file)
	output, err := command.Output()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
