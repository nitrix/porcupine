package porcupine

/*
#cgo LDFLAGS: -ldl
#include "include/pv_porcupine.h"
#include <dlfcn.h>
#include <string.h>
#include <stdlib.h>
#include <stdint.h>

typedef int the_pv_porcupine_init(const char *model_path, int32_t num_keywords, const char * const * keyword_paths, const float *sensitivities, pv_porcupine_t **object);
typedef void the_pv_porcupine_delete(pv_porcupine_t *object);
typedef int the_pv_porcupine_process(pv_porcupine_t *object, const int16_t *pcm, int32_t *keyword_index);
typedef int32_t the_pv_porcupine_frame_length(void);

int my_pv_porcupine_init(void *f, const char *model_path, int32_t num_keywords, const char * const * keyword_paths, const float *sensitivities, pv_porcupine_t **object) {
    return ((the_pv_porcupine_init*) f)(model_path, num_keywords, keyword_paths, sensitivities, object);
}

void my_pv_porcupine_delete(void *f, pv_porcupine_t *object) {
    ((the_pv_porcupine_delete*) f)(object);
}

int my_pv_porcupine_process(void *f, pv_porcupine_t *object, const int16_t *pcm, int32_t *keyword_index) {
    return ((the_pv_porcupine_process*) f)(object, pcm, keyword_index);
}

int32_t my_pv_porcupine_frame_length(void *f) {
    return ((the_pv_porcupine_frame_length*) f)();
}
*/
import "C"
import _ "embed"
import "unsafe"

//go:embed lib/libpv_porcupine.so
var libraryData []byte

func temporaryLibrary() (string, error) {
	return memoizeIntoFile(libraryData, "libpv_porcupine.so")
}

type Porcupine struct {
	library unsafe.Pointer
	initProc unsafe.Pointer
	processProc unsafe.Pointer
	frameLengthProc unsafe.Pointer
	deleteProc unsafe.Pointer
	handle  *C.struct_pv_porcupine
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

	library := C.dlopen(C.CString(libraryFilepath), C.RTLD_LAZY)

	cPorcupineInitStr := C.CString("pv_porcupine_init")
	cPorcupineProcessStr := C.CString("pv_porcupine_process")
	cPorcupineFrameLengthStr := C.CString("pv_porcupine_frame_length")
	cPorcupineDeleteStr := C.CString("pv_porcupine_delete")

	p := &Porcupine{
		library: library,
		initProc: C.dlsym(library, cPorcupineInitStr),
		processProc: C.dlsym(library, cPorcupineProcessStr),
		frameLengthProc: C.dlsym(library, cPorcupineFrameLengthStr),
		deleteProc: C.dlsym(library, cPorcupineDeleteStr),
	}

	cModelFilepath := C.CString(modelFilepath)
	cKeywordFilepath := C.CString(keyword.FilePath)

	defer func() {
		C.free(unsafe.Pointer(cModelFilepath))
		C.free(unsafe.Pointer(cKeywordFilepath))
		C.free(unsafe.Pointer(cPorcupineInitStr))
		C.free(unsafe.Pointer(cPorcupineProcessStr))
		C.free(unsafe.Pointer(cPorcupineFrameLengthStr))
		C.free(unsafe.Pointer(cPorcupineDeleteStr))
	}()

	cSensitivity := C.float(sensitivity)
	status := C.my_pv_porcupine_init(p.initProc, cModelFilepath, C.int32_t(1), &cKeywordFilepath, &cSensitivity, &p.handle)
	if err := checkStatus(int(status)); err != nil {
		return nil, err
	}

	p.keyword = keyword

	return p, nil
}

func (p *Porcupine) Destroy() {
	C.my_pv_porcupine_delete(p.deleteProc, p.handle)
	C.dlclose(p.library)
}

func (p Porcupine) Process(data []int16) (string, error) {
	var result C.int32_t

	d := (*C.int16_t)(unsafe.Pointer(&data[0]))
	status := C.my_pv_porcupine_process(p.processProc, p.handle, d, &result)

	if err := checkStatus(int(status)); err != nil || int(result) == -1 {
		return "", err
	}

	return p.keyword.Label, nil
}

func (p Porcupine) FrameLength() int {
	tmp := C.my_pv_porcupine_frame_length(p.frameLengthProc)
	return int(tmp)
}