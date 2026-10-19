package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Charan010/stash/binio"
)

const (
	magic   = "STSHMAN1"
	version = uint32(1)

	headerSize = 8 + 4 + 8 + 8
)

func WriteAtomic(path string, m Manifest) error {
	if err := m.Validate(); err != nil {

		return fmt.Errorf("manifest: validate before write: %w", err)
	}

	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("manifest: create directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".manifest-*.tmp")

	if err != nil {
		return fmt.Errorf("manifest: create temp file: %w", err)
	}

	tmpPath := tmp.Name()

	committed := false

	defer func() {
		_ = tmp.Close()

		if !committed {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := writeManifest(tmp, m); err != nil {
		return err
	}

	/*
		flushing the data to actual disk to prevent any data loss when the tool crashes.
		dirty pages stay on OS page cache and gets flushed occasionally . So, to force correctness,
		we flush to actual disk every time we write a manifest.
	*/
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("manifest: sync temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("manifest: close temp file: %w", err)
	}

	/*
		renames are atomic in unix systems and .tmp would be removed
		from <manifest_name>.tmp
	*/
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("manifest: rename temp file: %w", err)
	}

	if err := syncDirectory(dir); err != nil {
		return err
	}

	committed = true

	return nil
}

func writeManifest(file *os.File, m Manifest) error {
	w := binio.NewWriter(headerSize)

	w.Bytes([]byte(magic))
	w.Uint32(version)
	w.Uint64(m.Size)
	w.Uint64(uint64(len(m.Chunks)))

	if err := writeAll(file, w.Data()); err != nil {
		return fmt.Errorf("manifest: write header: %w", err)
	}

	for i, hash := range m.Chunks {
		if err := writeAll(file, hash[:]); err != nil {
			return fmt.Errorf("manifest: write chunk %d: %w", i, err)
		}
	}

	return nil
}

func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("manifest: open directory for sync: %w", err)
	}

	defer dir.Close()

	if err := dir.Sync(); err != nil {
		return fmt.Errorf("manifest: sync directory: %w", err)
	}

	return nil
}

func writeAll(file *os.File, data []byte) error {
	for len(data) > 0 {
		n, err := file.Write(data)

		if err != nil {
			return err
		}

		if n == 0 {
			return fmt.Errorf("manifest: short write")
		}

		data = data[n:]
	}

	return nil
}
