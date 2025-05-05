package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/tech-arch1tect/rknn-llm-go/generated"
)

func main() {
	mp := os.Getenv("MODEL_PATH")
	if mp == "" {
		fmt.Fprintln(os.Stderr, "Please set MODEL_PATH")
		os.Exit(1)
	}
	fi, err := os.Stat(mp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot stat model: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using %q (%d bytes)\n\n", mp, fi.Size())

	def := generated.Rkllm_createDefaultParam()
	def.Deref()

	const cpuCount = 8
	param := generated.RKLLMParam{
		Model_path:         append([]byte(mp), 0),
		Is_async:           true,
		Max_context_len:    def.Max_context_len,
		Max_new_tokens:     def.Max_new_tokens,
		Top_k:              def.Top_k,
		N_keep:             def.N_keep,
		Top_p:              def.Top_p,
		Temperature:        def.Temperature,
		Repeat_penalty:     def.Repeat_penalty,
		Frequency_penalty:  def.Frequency_penalty,
		Presence_penalty:   def.Presence_penalty,
		Mirostat:           def.Mirostat,
		Mirostat_tau:       def.Mirostat_tau,
		Mirostat_eta:       def.Mirostat_eta,
		Skip_special_token: def.Skip_special_token,
		Extend_param: generated.RKLLMExtendParam{
			Enabled_cpus_num:  int8(cpuCount),
			Enabled_cpus_mask: uint32((1 << cpuCount) - 1),
		},
	}
	params := []generated.RKLLMParam{param}

	var h generated.LLMHandle
	cArr := C.calloc(1, C.size_t(unsafe.Sizeof(h)))
	if cArr == nil {
		fmt.Fprintln(os.Stderr, "calloc failed")
		os.Exit(1)
	}
	defer C.free(cArr)

	*(**generated.LLMHandle)(cArr) = &h

	handles := unsafe.Slice((**generated.LLMHandle)(cArr), 1)

	var output strings.Builder
	done := make(chan struct{})

	cb := generated.LLMResultCallback(func(res *generated.RKLLMResult, _ unsafe.Pointer, state generated.LLMCallState) {
		if res == nil {
			close(done)
			return
		}
		res.Deref()

		p := (*C.char)(unsafe.Pointer(res.Text))
		s := C.GoString(p)
		output.WriteString(s)
		fmt.Print(s)
	})

	if rc := generated.Rkllm_init(handles, params, cb); rc != 0 {
		fmt.Fprintf(os.Stderr, "init failed (rc=%d)\n", rc)
		os.Exit(1)
	}

	prompt := "Tell me a short joke"
	inputs := []generated.RKLLMInput{{Input_type: generated.RKLLM_INPUT_PROMPT}}
	generated.RKLLMInput_SetPrompt(inputs, prompt)
	// or to use tokens
	//inputs = []generated.RKLLMInput{{Input_type: generated.RKLLM_INPUT_TOKENS}}
	//generated.RKLLMInput_SetToken(inputs, generated.RKLLMTokenInput{
	//	Input_ids: []int32{1, 2, 3},
	//	N_tokens:  3,
	//})

	infer := []generated.RKLLMInferParam{{Mode: generated.RKLLM_INFER_GENERATE}}

	fmt.Println(">>> Running inference …")
	if rc := generated.Rkllm_run(handles[0], inputs, infer, nil); rc != 0 {
		fmt.Fprintf(os.Stderr, "run_async failed (rc=%d)\n", rc)
		os.Exit(1)
	}

	<-done
	fmt.Println("\n\n=== Full output ===")
	fmt.Println(output.String())

	if rc := generated.Rkllm_destroy(handles[0]); rc != 0 {
		fmt.Fprintf(os.Stderr, "destroy failed (rc=%d)\n", rc)
		os.Exit(1)
	}
}
