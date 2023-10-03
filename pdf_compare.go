package main

import (
	"fmt"
	"os"
)

// not a nice way to compare pdfs but it works for testing and refactoring
// pdfs are considered to be not equal when the size doenst match or when a
// threshold of bytes differ
const MAX_DIFF_THRESHOLD = 500 // empirisch ermittelt

func compareWithReferenceFile(filename string, byteArray []byte) error {
	referenceFileContent, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load reference file: %v", err)
	}

	if len(byteArray) != len(referenceFileContent) {
		return fmt.Errorf("size mismatch: Expected %d bytes but got %d bytes", len(referenceFileContent), len(byteArray))
	}

	diffs := getDifferences(byteArray, referenceFileContent)
	if len(diffs) > MAX_DIFF_THRESHOLD {
		for _, diff := range diffs {
			fmt.Printf("Difference at index %d: Expected %v but got %v\n", diff.index, diff.expected, diff.actual)
		}
		return fmt.Errorf("generated content differs from the reference by %d bytes, which exceeds the threshold", len(diffs))
	}

	return nil
}

type Difference struct {
	index    int
	expected byte
	actual   byte
}

func getDifferences(a, b []byte) []Difference {
	var diffs []Difference
	length := len(a)
	if len(b) < length {
		length = len(b)
	}

	for i := 0; i < length; i++ {
		if a[i] != b[i] {
			diffs = append(diffs, Difference{
				index:    i,
				expected: b[i],
				actual:   a[i],
			})
		}
	}

	// Handle case where one byte slice is longer than the other
	for i := length; i < len(a); i++ {
		diffs = append(diffs, Difference{
			index:  i,
			actual: a[i],
		})
	}

	for i := length; i < len(b); i++ {
		diffs = append(diffs, Difference{
			index:    i,
			expected: b[i],
		})
	}

	return diffs
}
