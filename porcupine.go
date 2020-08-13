package porcupine

// #cgo LDFLAGS: -L../../others/lib -lpv_porcupine
// #include <stdlib.h>
// #include "../../others/include/pv_porcupine.h"
import "C"

import (
	"errors"
	"unsafe"
)

var (
	ErrOutOfMemory     = errors.New("porcupine: out of memory")
	ErrIOError         = errors.New("porcupine: IO error")
	ErrInvalidArgument = errors.New("porcupine: invalid argument")
	ErrUnknownStatus   = errors.New("unknown status code")
)

type Porcupine struct {
	handle  *C.struct_pv_porcupine
	keyword *Keyword
}

type Keyword struct {
	Label       string
	FilePath    string
	Sensitivity float32
}


func New(modelFilepath string, keyword *Keyword) (Porcupine, error) {
	var handle *C.struct_pv_porcupine

	cModelFilepath := C.CString(modelFilepath)
	cKeywordCount := C.int32_t(1)
	cKeywordFilepath := C.CString(keyword.FilePath)
	cSensitivity := C.float(keyword.Sensitivity)

	p := Porcupine{}

	defer func() {
		C.free(unsafe.Pointer(cModelFilepath))
		C.free(unsafe.Pointer(cKeywordFilepath))
	}()

	status := C.pv_porcupine_init(cModelFilepath, cKeywordCount, &cKeywordFilepath, &cSensitivity, &handle)
	if err := checkStatus(status); err != nil {
		return p, err
	}

	p.handle = handle
	p.keyword = keyword

	return p, nil
}

func (p *Porcupine) Destroy() {
	C.pv_porcupine_delete(p.handle)
	p.handle = nil
}

func (p Porcupine) Process(data []int16) (string, error) {
	var result C.int32_t

	status := C.pv_porcupine_process(p.handle, (*C.int16_t)(unsafe.Pointer(&data[0])), &result)

	if err := checkStatus(status); err != nil || int(result) == -1 {
		return "", err
	}

	return p.keyword.Label, nil
}

func FrameLength() int {
	tmp := C.pv_porcupine_frame_length()
	return int(tmp)
}


func checkStatus(status C.pv_status_t) error {
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
