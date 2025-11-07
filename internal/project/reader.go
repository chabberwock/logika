package project

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
)

type FilterFunc func(map[string]any) (bool, error)

type LogReader struct {
	line       int
	scanner    *bufio.Scanner
	Filter     FilterFunc
	current    map[string]any
	currentErr error
}

func NewLogReader(f *os.File) *LogReader {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024*10), 1024*1024*10)
	return &LogReader{
		scanner: scanner,
	}
}

func (l *LogReader) Next() bool {
	for l.scanner.Scan() {
		l.line++
		l.current, l.currentErr = l.toRow(l.scanner.Bytes(), l.line)
		l.currentErr = errors.Join(l.currentErr, l.scanner.Err())
		if l.currentErr != nil {
			return false
		}
		if l.Filter == nil {
			continue
		}
		ok, err := l.Filter(l.current)
		if err != nil {
			break
		}
		if ok {
			return true
		}
	}
	l.currentErr = errors.Join(l.currentErr, l.scanner.Err())
	return false
}

func (l *LogReader) Err() error {
	return l.currentErr
}

func (l *LogReader) Row() map[string]any {
	return l.current
}

func (l *LogReader) toRow(b []byte, line int) (map[string]any, error) {
	data := make(map[string]any)
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return map[string]any{
		"line": line,
		"data": data,
	}, nil
}
