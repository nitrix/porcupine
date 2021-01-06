package porcupine

// #include <stdlib.h>
// #include "include/pv_porcupine.h"
import "C"

import (
	"crypto/sha1"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed model/porcupine_params.pv
var modelData []byte

var (
	ErrOutOfMemory     = errors.New("porcupine: out of memory")
	ErrIOError         = errors.New("porcupine: IO error")
	ErrInvalidArgument = errors.New("porcupine: invalid argument")
	ErrUnknownStatus   = errors.New("unknown status code")
)

type Keyword struct {
	Label       string
	FilePath    string
	Sensitivity float32
}

func checkStatus(status int) error {
	switch status {
	case C.PV_STATUS_SUCCESS:
		return nil
	case C.PV_STATUS_OUT_OF_MEMORY:
		return ErrOutOfMemory
	case C.PV_STATUS_INVALID_ARGUMENT:
		return ErrInvalidArgument
	case C.PV_STATUS_IO_ERROR:
		return ErrIOError
	default:
		return ErrUnknownStatus
	}
}

func temporaryModelFile() (string, error) {
	return memoizeIntoFile(modelData, "porcupine_params.pv")
}

func memoizeIntoFile(source []byte, name string) (string, error) {
	hasher := sha1.New()
	hasher.Write(source)
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	temporaryFolder := filepath.Join(os.TempDir(), hash)
	temporaryPath := filepath.Join(temporaryFolder, name)

	err := os.Mkdir(temporaryFolder, 0755)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	if _, err = os.Stat(temporaryPath); os.IsNotExist(err) {
		err = os.WriteFile(temporaryPath, source, 0600)
		if err != nil {
			return "", err
		}
	}

	return temporaryPath, nil
}