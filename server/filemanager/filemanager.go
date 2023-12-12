/* package filemanager is responsible for the saving of files along with the calculation of metadata */

package filemanager

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mandiant/gocrack/server/storage"
)

// Context contains the state of the filemanager and provides public methods to interact
// with all filesystem actions within gocrack
type Context struct {
	stor FileManagerStor
	cfg  Config
}

type FileSaveResponse struct {
	Size          int64
	SavedTo       string
	NumberOfLines int64
	SHA1          string
}

type FileManagerStor interface {
	NewEngineFileTransaction() (storage.EngineFileTxn, error)
	DeleteEngineFile(string) error
	DeleteTaskFile(string) error
}

var (
	// ErrCannotImport is returned whenever we are told to reload but cannot due to missing import_directory
	ErrCannotImport = errors.New("filemanager: cannot import because import_directory is not set in the config")
	// SystemUserUUID is the UUID of files uploaded by the GoCrack system
	SystemUserUUID = "b2c9e661-74e5-4ba3-b1ce-624894e85622"
)

func splitFilePath(rootPath, filename string) string {
	filename = strings.Replace(filename, "-", "", -1)
	var parts = []string{rootPath}

	for i := 0; i < 14; i += 2 {
		parts = append(parts, filename[i:i+2])
	}
	parts = append(parts, filename)

	return filepath.Join(parts...)
}

// New creates a new FileManager for GoCrack
func New(stor FileManagerStor, cfg Config) *Context {
	return &Context{
		stor: stor,
		cfg:  cfg,
	}
}

// Refresh checks the import directory for new files and saves them into the database
func (s *Context) Refresh() error {
	return s.importDirectory()
}

// SaveFile returns metadata about the file and saves the file to disk
func (s *Context) SaveFile(src io.ReadCloser, filename string, fileUUID string, filetype interface{}) (*FileSaveResponse, error) {
	var rootpath string
	var finalSavePath string

	switch t := filetype.(type) {
	case storage.TaskFileEngine:
		rootpath = s.cfg.TaskUploadPath
	case storage.EngineFileType:
		rootpath = s.cfg.EngineFilePath
	default:
		return nil, fmt.Errorf("unknown filetype `%v`", t)
	}

	// XXX(cschmitt): This is kinda messy how we have to handle file uploads.. we should create something like python's buffered temp file
	fd, err := ioutil.TempFile(s.cfg.TempPath, "")
	if err != nil {
		return nil, err
	}

	szrec := &WriteSizeLineRecorder{}
	hasher := sha1.New()

	// Write the file to the temp file, hasher, and our size/line recorder
	mw := io.MultiWriter(fd, hasher, szrec)
	if _, err = io.Copy(mw, src); err != nil {
		fd.Close()
		return nil, nil
	}

	// Close the file we just wrote to disk
	if err := fd.Close(); err != nil {
		return nil, err
	}

	fileHash := fmt.Sprintf("%x", hasher.Sum(nil))
	finalSavePath = splitFilePath(rootpath, fileUUID)
	rootPath, _ := filepath.Split(finalSavePath)

	// Create the nested directory
	if err := os.MkdirAll(rootPath, 0770); err != nil {
		return nil, err
	}

	// Move the tempfile over to the new path
	if err := os.Rename(fd.Name(), finalSavePath); err != nil {
		return nil, err
	}

	return &FileSaveResponse{
		SHA1:          fileHash,
		SavedTo:       finalSavePath,
		Size:          szrec.Size(),
		NumberOfLines: szrec.Lines(),
	}, nil
}

// DeleteFile removes the file from the system
func (s *Context) DeleteFile(doc interface{}) error {
	switch t := doc.(type) {
	case storage.EngineFile:
		if err := s.stor.DeleteEngineFile(t.FileID); err != nil {
			return err
		}
		return os.Remove(t.SavedAt)
	case storage.TaskFile:
		if err := s.stor.DeleteTaskFile(t.FileID); err != nil {
			return err
		}
		return os.Remove(t.SavedAt)
	default:
		return fmt.Errorf("unknown filetype `%v`", t)
	}
}
