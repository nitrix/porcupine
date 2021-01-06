package porcupine

// #include <stdlib.h>
// #include <stdint.h>
import "C"

import (
	_ "embed"
	"golang.org/x/sys/windows"
	"math"
	"unsafe"
)

//go:embed lib/libpv_porcupine.dll
var libraryData []byte

func temporaryLibrary() (string, error) {
	return memoizeIntoFile(libraryData, "libpv_porcupine.dll")
}

type Porcupine struct {
	library *windows.DLL
	processProc *windows.Proc
	frameLengthProc *windows.Proc
	handle  uintptr
	keyword *Keyword
}

func New(keyword *Keyword, sensitivity float32) (*Porcupine, error) {
	modelFilepath, err := temporaryModelFile()
	if err != nil {
		return nil, err
	}

	libraryFilepath, err := temporaryLibrary()
	if err != nil {
		return nil, err
	}

	library, err := windows.LoadDLL(libraryFilepath)
	if err != nil {
		return nil, err
	}

	p := &Porcupine{
		library: library,
		processProc: library.MustFindProc("pv_porcupine_process"),
		frameLengthProc: library.MustFindProc("pv_porcupine_frame_length"),
	}

	initProc := p.library.MustFindProc("pv_porcupine_init")

	cModelFilepath := C.CString(modelFilepath)
	cKeywordFilepath := C.CString(keyword.FilePath)

	defer func() {
		C.free(unsafe.Pointer(cModelFilepath))
		C.free(unsafe.Pointer(cKeywordFilepath))
	}()

	sensitivityBits := math.Float32bits(sensitivity)
	status, _, err := initProc.Call(uintptr(unsafe.Pointer(cModelFilepath)), uintptr(1), uintptr(unsafe.Pointer(&cKeywordFilepath)), uintptr(unsafe.Pointer(&sensitivityBits)), uintptr(unsafe.Pointer(&p.handle)))
	if err := checkStatus(int(status)); err != nil {
		return nil, err
	}

	p.keyword = keyword

	return p, nil
}

func (p *Porcupine) Destroy() {
	deleteProc := p.library.MustFindProc("pv_porcupine_delete")
	_, _, _ = deleteProc.Call(p.handle)
	_ = p.library.Release()
}

func (p Porcupine) Process(data []int16) (string, error) {
	var result C.int32_t

	d := (*C.int16_t)(unsafe.Pointer(&data[0]))
	status, _, _ := p.processProc.Call(p.handle, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(&result)))

	if err := checkStatus(int(status)); err != nil || int(result) == -1 {
		return "", err
	}

	return p.keyword.Label, nil
}

func (p Porcupine) FrameLength() int {
	tmp, _, _ := p.frameLengthProc.Call()
	return int(tmp)
}