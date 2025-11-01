package project

import (
	"bufio"
	"fmt"
	"os"
)

type BatchReader struct {
	f         *os.File
	scanner   *bufio.Scanner
	batchSize int
	lines     []string
}

func NewBatchReader(src string, batchSize int) (*BatchReader, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("opening source file: %w", err)
	}
	scanner := bufio.NewScanner(f)
	return &BatchReader{
		f:         f,
		scanner:   scanner,
		batchSize: batchSize,
	}, nil
}

func (b *BatchReader) Next() bool {
	for i := 0; i < b.batchSize; i++ {
		if !b.scanner.Scan() {
			return b.lines != nil
		}
		b.lines = append(b.lines, b.scanner.Text())
	}
	return true
}

func (b *BatchReader) Lines() []string {
	defer func() {
		b.lines = nil
	}()
	return b.lines
}

func (b *BatchReader) Err() error {
	return b.scanner.Err()
}

func (b *BatchReader) Close() error {
	return b.f.Close()
}
