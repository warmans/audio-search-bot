package importer

import (
	"context"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/jmoiron/sqlx"
	"github.com/warmans/audio-search-bot/internal/audiometa"
	"github.com/warmans/audio-search-bot/internal/metadata"
	"github.com/warmans/audio-search-bot/internal/search"
	"github.com/warmans/audio-search-bot/internal/store"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"
)

const filePollingInterval = time.Second * 10

type pendingFile struct {
	srtFilePath string
	modTime     time.Time
}

func (f pendingFile) inferredMediaPath() string {
	return fmt.Sprintf("%s.mp3", strings.TrimSuffix(f.srtFilePath, path.Ext(f.srtFilePath)))
}

func NewIncrementalImporter(
	srtDir string,
	metadataDir string,
	conn *store.Conn,
	searcher *search.BlugeSearch,
	logger *slog.Logger,
	useFilePolling bool,
) *Incremental {
	return &Incremental{
		srtDir:         srtDir,
		metadataDir:    metadataDir,
		conn:           conn,
		searcher:       searcher,
		logger:         logger,
		useFilePolling: useFilePolling,
	}
}

type Incremental struct {
	srtDir         string
	metadataDir    string
	conn           *store.Conn
	searcher       *search.BlugeSearch
	logger         *slog.Logger
	useFilePolling bool
}

func (i *Incremental) Start(ctx context.Context) error {

	i.logger.Info("Starting initial file sync...")
	if err := i.importAllNew(ctx); err != nil {
		return err
	}

	i.logger.Info("Starting incremental file sync...", slog.Bool("polling", i.useFilePolling))
	if i.useFilePolling {
		return i.startFilePolling(ctx)
	}
	return i.startFileWatch(ctx)
}

func (i *Incremental) startFilePolling(ctx context.Context) error {
	for {
		time.Sleep(filePollingInterval)
		if err := i.importAllNew(ctx); err != nil {
			return err
		}
	}
}

func (i *Incremental) startFileWatch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// since files are typically added in batches
	// group up as many pending files as are detected in a 2s period
	// instead of dispatching an import for each file.
	ticker := time.NewTicker(time.Second * 2)
	var pendingFiles []pendingFile

	// Start listening for events.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok || !strings.HasSuffix(event.Name, ".srt") {
					return
				}
				if event.Has(fsnotify.Create) {
					stat, err := os.Stat(event.Name)
					if err != nil {
						i.logger.Error("failed stat file", slog.String("err", err.Error()))
						continue
					}

					pending := pendingFile{srtFilePath: event.Name, modTime: stat.ModTime()}

					mediaExists, err := i.mediaFileExists(pending)
					if err != nil {
						i.logger.Error("failed to locate associated media file", slog.String("err", err.Error()), slog.String("srtPath", event.Name))
						continue
					}
					if !mediaExists {
						i.logger.Info("no media file for SRT, skipping for now...", slog.String("srtPath", event.Name))
						continue
					}
					pendingFiles = append(pendingFiles, pending)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				i.logger.Error("watcher error", slog.String("err", err.Error()))
			case <-ticker.C:
				if len(pendingFiles) > 0 {
					if err := i.importNew(ctx, pendingFiles); err != nil {
						i.logger.Error(
							"Failed to import pending files",
							slog.String("err", err.Error()),
						)
					}
					pendingFiles = []pendingFile{}
				}
			}
		}
	}()

	// Add a path.
	err = watcher.Add(i.srtDir)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (i *Incremental) importAllNew(ctx context.Context) error {
	manifest, err := store.NewSRTStore(i.conn.Db).GetManifest()
	if err != nil {
		return err
	}

	dirEntries, err := os.ReadDir(i.srtDir)
	if err != nil {
		return err
	}

	toImport := []pendingFile{}
	for _, v := range dirEntries {
		if !strings.HasSuffix(v.Name(), ".srt") {
			continue
		}
		inf, err := v.Info()
		if err != nil {
			return err
		}
		addToImport := false
		if oldModTime, ok := manifest[path.Join(i.srtDir, v.Name())]; ok {
			if inf.ModTime().After(oldModTime) {
				i.logger.Info("file older than existing",
					slog.String("path", path.Join(i.srtDir, v.Name())),
					slog.Time("old", oldModTime),
					slog.Time("new", inf.ModTime()),
				)
				addToImport = true
			}
		} else {
			addToImport = true
		}
		if addToImport {
			toImport = append(toImport, pendingFile{srtFilePath: path.Join(i.srtDir, v.Name()), modTime: inf.ModTime()})
		}
	}
	if len(toImport) == 0 {
		return nil
	}

	i.logger.Info("Importing files...", slog.Int("num_files", len(toImport)), slog.String("dir", i.srtDir))
	if err := i.importNew(ctx, toImport); err != nil {
		i.logger.Error(
			"Failed to import pending files",
			slog.String("err", err.Error()),
		)
	}
	return nil
}

func (i *Incremental) importNew(ctx context.Context, pendingFiles []pendingFile) error {

	for k, pending := range pendingFiles {

		if err := audiometa.DumpMeta(pending.inferredMediaPath()); err != nil {
			return fmt.Errorf("failed to dump metadata for file: %s: %w", pending.inferredMediaPath(), err)
		}

		err := i.conn.WithTx(func(tx *sqlx.Tx) error {
			meta, err := metadata.CreateMetadataFromSRT(pending.srtFilePath, i.metadataDir)
			if err != nil {
				return fmt.Errorf("failed to create metadata: %w", err)
			}
			logger := i.logger.With(slog.String("media_id", meta.ID()), slog.Time("modtime", pending.modTime))

			s := store.NewSRTStore(tx)
			result, err := s.ManifestAdd(pending.srtFilePath, pending.modTime)
			if err != nil {
				return fmt.Errorf("failed to add to manifest: %w", err)
			}
			if result == store.UpsertResultNoop {
				// nothing to do
				logger.Info("File already processed, skipped")
				return nil
			}

			if err := s.ImportMedia(*meta); err != nil {
				return err
			}

			logger.Info("Import to index...", slog.String("result", string(result)), slog.Float64("progress", float64(k)/float64(len(pendingFiles))*100))
			return i.searcher.Import(ctx, meta, result == store.UpsertResultUpdated)
		})
		if err != nil {
			return err
		}
		if k%100 == 0 {
			if err := i.searcher.RefreshIndex(); err != nil {
				return err
			}
		}
	}
	return i.searcher.RefreshIndex()
}

func (i *Incremental) mediaFileExists(pending pendingFile) (bool, error) {
	_, err := os.Stat(pending.inferredMediaPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat media file %s: %w", pending.inferredMediaPath(), err)
	}

	return true, nil
}
