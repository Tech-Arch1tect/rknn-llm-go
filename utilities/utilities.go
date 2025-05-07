package utilities

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"

	"github.com/tech-arch1tect/rknn-llm-go/generated"
)

func AllocLLMHandles(n int) ([]*generated.LLMHandle, func(), error) {
	if n <= 0 {
		return nil, nil, fmt.Errorf("n must be ≥ 1, got %d", n)
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
	return slice, func() { C.free(cArr) }, nil
}

type RunContext struct {
	Done     chan struct{}
	Callback func(text string, state generated.LLMCallState)
}

func DynamicResultCallback(
	res *generated.RKLLMResult,
	userdata unsafe.Pointer,
	state generated.LLMCallState,
) {
	h := cgo.Handle(uintptr(userdata))
	ctx := h.Value().(RunContext)

	if res == nil {
		close(ctx.Done)
		h.Delete()
		return
	}

	res.Deref()
	p := (*C.char)(unsafe.Pointer(res.Text))
	text := C.GoString(p)
	ctx.Callback(text, state)
}
