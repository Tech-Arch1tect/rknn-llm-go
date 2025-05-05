package utilities

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/tech-arch1tect/rknn-llm-go/generated"
)

func AllocLLMHandles(n int) ([]*generated.LLMHandle, func(), error) {
	if n <= 0 {
		return nil, nil, fmt.Errorf("n must be ≥ 1, got %d", n)
	}

	ptrSize := unsafe.Sizeof((*generated.LLMHandle)(nil))
	cArr := C.calloc(C.size_t(n), C.size_t(ptrSize))
	if cArr == nil {
		return nil, nil, fmt.Errorf("C.calloc failed")
	}

	slice := unsafe.Slice((**generated.LLMHandle)(cArr), n)

	for i := range slice {
		slice[i] = new(generated.LLMHandle)
	}

	cleanup := func() { C.free(cArr) }
	return slice, cleanup, nil
}

func NewCallbackHelper(done chan struct{}, callback func(string, generated.LLMCallState)) generated.LLMResultCallback {
	return generated.LLMResultCallback(func(
		res *generated.RKLLMResult,
		_ unsafe.Pointer,
		state generated.LLMCallState,
	) {
		if res == nil {
			close(done)
			return
		}

		res.Deref()

		p := (*C.char)(unsafe.Pointer(res.Text))
		s := C.GoString(p)

		callback(s, state)
	})
}
