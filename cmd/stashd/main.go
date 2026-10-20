package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Charan010/stash/index"
	"github.com/Charan010/stash/pack"
	"github.com/Charan010/stash/packfile"
	"github.com/Charan010/stash/server"
	"github.com/Charan010/stash/store"
)

const (
	dataDir      = ".stash"
	activePackID = uint64(1)
)

func openStore() (*store.Store, func(), error) {

	packDir := filepath.Join(dataDir, "packs")
	manifestDir := filepath.Join(dataDir, "manifests")

	indexPath := filepath.Join(dataDir, "index.log")

	if err := os.MkdirAll(packDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create pack dir : %w", err)
	}

	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create manifest dir: %w", err)
	}

	idx := index.New()

	logFile, err := index.OpenLog(indexPath)

	if err != nil {
		return nil, nil, fmt.Errorf("open index log : %w", err)
	}

	if err := logFile.Replay(idx); err != nil {

		_ = logFile.Close()
		return nil, nil, fmt.Errorf("replay index log : %w", err)
	}

	packs := packfile.NewManager(packDir)
	packPath := filepath.Join(packDir, fmt.Sprintf("%016x.pack", activePackID))

	file, err := os.OpenFile(packPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)

	if err != nil {
		_ = logFile.Close()
		return nil, nil, fmt.Errorf("open active pack: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		_ = logFile.Close()

		return nil, nil, err
	}

	var writer *pack.Writer

	if info.Size() == 0 {

		writer = pack.NewWriter(file)

		header := pack.Header{
			ObjectCount: 0,
			Version:     pack.Version,
		}

		if err := writer.WriteHeader(header); err != nil {
			_ = file.Close()
			_ = logFile.Close()

			return nil, nil, err
		}

	} else {
		writer = pack.NewWriterAt(file, uint64(info.Size()))
	}

	s := store.New(idx, logFile, packs, activePackID, writer, manifestDir)

	cleanup := func() {
		_ = logFile.Sync()
		_ = logFile.Close()
		_ = packs.Close()
		_ = file.Close()
	}

	return s, cleanup, nil
}

func main() {

	s, cleanup, err := openStore()

	if err != nil {
		log.Fatal(err)
	}

	defer cleanup()

	srv := server.New(s)
	addr := ":8080"

	fmt.Println("stash server listening on", addr)

	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
