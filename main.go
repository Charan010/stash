package main

import (
	"fmt"
	"os"
	"path/filepath"

	objecthash "github.com/Charan010/stash/hash"
	"github.com/Charan010/stash/index"
	"github.com/Charan010/stash/manifest"
	"github.com/Charan010/stash/pack"
	"github.com/Charan010/stash/packfile"
	"github.com/Charan010/stash/store"
)

func main() {
	const (
		dataDir      = ".stash"
		inputPath    = "test.bin"
		outputPath   = "restored.bin"
		activePackID = uint64(1)
	)

	packDir := filepath.Join(dataDir, "packs")
	manifestDir := filepath.Join(dataDir, "manifests")
	indexPath := filepath.Join(dataDir, "index.log")

	if err := os.MkdirAll(packDir, 0o755); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		panic(err)
	}

	idx := index.New()

	log, err := index.OpenLog(indexPath)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	if err := log.Replay(idx); err != nil {
		panic(err)
	}

	packs := packfile.NewManager(packDir)
	defer packs.Close()

	packPath := filepath.Join(
		packDir,
		fmt.Sprintf("%016x.pack", activePackID),
	)

	packFile, err := os.OpenFile(
		packPath,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0o644,
	)
	if err != nil {
		panic(err)
	}
	defer packFile.Close()

	info, err := packFile.Stat()
	if err != nil {
		panic(err)
	}

	var writer *pack.Writer

	if info.Size() == 0 {
		writer = pack.NewWriter(packFile)

		if err := writer.WriteHeader(pack.Header{
			Version:     pack.Version,
			ObjectCount: 0,
		}); err != nil {
			panic(err)
		}
	} else {
		writer = pack.NewWriterAt(
			packFile,
			uint64(info.Size()),
		)
	}

	s := store.New(
		idx,
		log,
		packs,
		activePackID,
		writer,
		manifestDir,
	)

	// Dumb test: read the whole file as one chunk.
	data, err := os.ReadFile(inputPath)
	if err != nil {
		panic(err)
	}

	chunkHash, err := s.Put(data)
	if err != nil {
		panic(err)
	}

	fmt.Println("chunk hash:", chunkHash)

	m := manifest.New(
		uint64(len(data)),
		[]objecthash.Hash{
			chunkHash,
		},
	)

	objectID, err := s.CommitManifest(m)
	if err != nil {
		panic(err)
	}

	fmt.Println("object id:", objectID)

	out, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	if err := s.Restore(objectID, out); err != nil {
		panic(err)
	}

	fmt.Println("restored to:", outputPath)
}
