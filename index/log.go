package index

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type Log struct {
	mu   sync.Mutex
	file *os.File
}

func OpenLog(path string) (*Log, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)

	if err != nil {
		return nil, fmt.Errorf("index: open log: %w", err)
	}

	return &Log{file: f}, nil
}

func (l *Log) Append(record Record) error {
	data := EncodeRecord(record)

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := writeAll(l.file, data); err != nil {
		return fmt.Errorf("index: append record: %w", err)
	}

	return nil
}

func (l *Log) Sync() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("index: sync log: %w", err)
	}

	return nil
}

func (l *Log) Replay(index *Index) error {

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("index: seek log: %w", err)
	}

	buf := make([]byte, RecordSize)

	for {

		n, err := io.ReadFull(l.file, buf)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			if errors.Is(err, io.ErrUnexpectedEOF) {
				// Partial final record from a crash.
				break
			}

			return fmt.Errorf("index: read log record: %w", err)
		}

		if n != RecordSize {
			return fmt.Errorf("index: invalid record size: got %d want %d", n, RecordSize)
		}

		record, err := DecodeRecord(buf)
		if err != nil {
			return fmt.Errorf("index: decode log record: %w", err)
		}

		index.Put(record.Hash, Location{PackID: record.PackID, Offset: record.Offset})
	}

	// Restore file position to end after replay.
	if _, err := l.file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("index: seek log end: %w", err)
	}

	return nil
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.file.Close(); err != nil {
		return fmt.Errorf("index: close log: %w", err)
	}

	return nil
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}

		if n == 0 {
			return io.ErrShortWrite
		}

		data = data[n:]
	}

	return nil
}
