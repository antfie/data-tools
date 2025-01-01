package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
)

const bufferSize = 4096

func CompareFiles(left, right string) (bool, error) {
	f1, err := os.Open(path.Clean(left))

	if err != nil {
		return false, fmt.Errorf("failed to open file for left-hand comparison: %v", err)
	}

	defer f1.Close()

	f2, err := os.Open(path.Clean(right))

	if err != nil {
		return false, fmt.Errorf("failed to open file for right-hand comparison: %v", err)
	}

	defer f2.Close()

	buf1 := make([]byte, bufferSize)
	buf2 := make([]byte, bufferSize)

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)

		if err1 != nil && err1 != io.EOF {
			return false, fmt.Errorf("error reading file for left-hand comparison: %v", err1)
		}

		if err2 != nil && err2 != io.EOF {
			return false, fmt.Errorf("error reading file for right-hand comparison: %v", err2)
		}

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}
	}
}
