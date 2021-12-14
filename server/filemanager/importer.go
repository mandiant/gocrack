package filemanager

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type extlookup map[string]storage.EngineFileType

var extensionsToType = extlookup{
	".hcmask":     storage.EngineFileMasks,
	".masks":      storage.EngineFileMasks,
	".dict":       storage.EngineFileDictionary,
	".dictionary": storage.EngineFileDictionary,
	".rule":       storage.EngineFileRules,
	".rules":      storage.EngineFileRules,
}

func (s *Context) importDirectory() error {
	if s.cfg.ImportPath == nil {
		return ErrCannotImport
	}

	filepath.Walk(*s.cfg.ImportPath, func(path string, info os.FileInfo, err error) error {
		var nonFatalError error

		if info.IsDir() {
			return nil
		}

		extension := filepath.Ext(path)
		fileType, ok := extensionsToType[extension]
		if !ok {
			log.Warn().Str("path", path).Msg("Skipping import due to invalid file extension")
			return nil
		}

		log.
			Info().
			Str("path", path).
			Uint8("type", uint8(fileType)).
			Str("extension", extension).
			Msg("Importing file")

		fileUUID := uuid.NewString()
		fd, err := os.Open(path)
		if err != nil {
			nonFatalError = err
			goto Done
		}

		if savedInfo, err := s.SaveFile(fd, info.Name(), fileUUID, fileType); err != nil {
			nonFatalError = err
			goto Done
		} else {
			txn, err := s.stor.NewEngineFileTransaction()
			if err != nil {
				nonFatalError = err
				goto Done
			}
			defer txn.Rollback()

			if err := txn.SaveEngineFile(storage.EngineFile{
				FileID:          fileUUID,
				FileName:        info.Name(),
				FileSize:        savedInfo.Size,
				UploadedBy:      "System",
				UploadedByUUID:  SystemUserUUID,
				UploadedAt:      time.Now().UTC(),
				FileType:        fileType,
				NumberOfEntries: savedInfo.NumberOfLines,
				IsShared:        true,
				SHA1Hash:        savedInfo.SHA1,
				SavedAt:         savedInfo.SavedTo,
			}); err != nil {
				nonFatalError = err
				goto Done
			}

			if err := txn.Commit(); err != nil {
				nonFatalError = err
				goto Done
			}

		}

	Done:
		if fd != nil {
			fd.Close()
			os.Remove(fd.Name())
		}

		if nonFatalError != nil {
			log.Error().Err(nonFatalError).Str("path", path).Msg("Failed to import file")
			return nil
		}

		log.Info().Str("path", path).Str("file_id", fileUUID).Msg("Imported file successfully")
		return nil
	})

	return nil
}
