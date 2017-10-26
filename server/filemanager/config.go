package filemanager

import (
	"fmt"
	"os"

	"github.com/fireeye/gocrack/shared"
)

var (
	errNotValidFolder = "file_manager.%s has to be a valid, writable folder"
	defaultMaxSize    = shared.GetIntPtr(20 * 1024 * 1024)
)

// Config describes the settings available to GoCrack's filemanager
type Config struct {
	TaskUploadPath string  `yaml:"task_file_path"`
	EngineFilePath string  `yaml:"engine_file_path"`
	TempPath       string  `yaml:"temp_path"`
	TaskMaxSize    *int    `yaml:"task_max_upload_size,omitempty"`
	ImportPath     *string `yaml:"import_path,omitempty"`
}

// Validate the FileManager config by ensuring the directories exist
func (s *Config) Validate() error {
	if s.TaskUploadPath == "" {
		return fmt.Errorf(errNotValidFolder, "task_file_path")
	}

	if _, err := os.Stat(s.TaskUploadPath); os.IsNotExist(err) {
		return fmt.Errorf(errNotValidFolder, "task_file_path")
	}

	if s.EngineFilePath == "" {
		return fmt.Errorf(errNotValidFolder, "engine_file_path")
	}

	if _, err := os.Stat(s.EngineFilePath); os.IsNotExist(err) {
		return fmt.Errorf(errNotValidFolder, "engine_file_path")
	}

	if s.TempPath == "" {
		return fmt.Errorf(errNotValidFolder, "temp_path")
	}

	if _, err := os.Stat(s.TempPath); os.IsNotExist(err) {
		return fmt.Errorf(errNotValidFolder, "temp_path")
	}

	// If optional import path is set, ensure it exists
	if s.ImportPath != nil {
		if _, err := os.Stat(*s.ImportPath); os.IsNotExist(err) {
			return fmt.Errorf(errNotValidFolder, "import_path")
		}
	}

	if s.TaskMaxSize == nil {
		s.TaskMaxSize = defaultMaxSize
	}

	return nil
}
