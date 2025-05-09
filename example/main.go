package main

import (
	"fmt"
	"os"
	"runtime/cgo"
	"strings"
	"unsafe"

	"github.com/tech-arch1tect/rknn-llm-go/generated"
	"github.com/tech-arch1tect/rknn-llm-go/utilities"
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
		Max_new_tokens:     100,
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
		Skip_special_token: true,
		Extend_param: generated.RKLLMExtendParam{
			Enabled_cpus_num:  int8(cpuCount),
			Enabled_cpus_mask: uint32((1 << cpuCount) - 1),
		},
	}

	handles, cleanup, err := utilities.AllocLLMHandles(1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cleanup()

	if rc := generated.Rkllm_init(
		&handles[0],
		&param,
		generated.LLMResultCallback(utilities.DynamicResultCallback),
	); rc != 0 {
		fmt.Fprintf(os.Stderr, "init failed (rc=%d)\n", rc)
		os.Exit(1)
	}

	inputs := []generated.RKLLMInput{{Input_type: generated.RKLLM_INPUT_PROMPT}}
	infer := []generated.RKLLMInferParam{{Mode: generated.RKLLM_INFER_GENERATE}}
	var output strings.Builder

	done1 := make(chan struct{})
	ctx1 := utilities.RunContext{
		Done: done1,
		Callback: func(text string, _ generated.LLMCallState) {
			output.WriteString(text)
			fmt.Print(text)
		},
	}
	handle1 := cgo.NewHandle(ctx1)

	generated.RKLLMInput_SetPrompt(inputs, "Tell me a short joke")
	// or to use tokens
	//inputs = []generated.RKLLMInput{{Input_type: generated.RKLLM_INPUT_TOKENS}}
	//generated.RKLLMInput_SetToken(inputs, generated.RKLLMTokenInput{
	//	Input_ids: []int32{1, 2, 3},
	//	N_tokens:  3,
	//})

	fmt.Println(">>> First inference …")
	if rc := generated.Rkllm_run(
		handles[0],
		inputs,
		infer,
		unsafe.Pointer(handle1),
	); rc != 0 {
		fmt.Fprintf(os.Stderr, "run failed (rc=%d)\n", rc)
		os.Exit(1)
	}
	<-done1
	fmt.Println("\n\n=== Joke complete ===")

	if rc := generated.Rkllm_clear_kv_cache(handles[0], 1); rc != 0 {
		fmt.Fprintf(os.Stderr, "clear_kv_cache failed (rc=%d)\n", rc)
		os.Exit(1)
	}

	output.Reset()
	done2 := make(chan struct{})
	ctx2 := utilities.RunContext{
		Done: done2,
		Callback: func(text string, _ generated.LLMCallState) {
			output.WriteString(text)
			fmt.Print(text)
		},
	}
	handle2 := cgo.NewHandle(ctx2)

	generated.RKLLMInput_SetPrompt(inputs, "Now tell me a short riddle")
	fmt.Println("\n>>> Second inference …")
	if rc := generated.Rkllm_run(
		handles[0],
		inputs,
		infer,
		unsafe.Pointer(handle2),
	); rc != 0 {
		fmt.Fprintf(os.Stderr, "run failed (rc=%d)\n", rc)
		os.Exit(1)
	}
	<-done2
	fmt.Println("\n\n=== Riddle complete ===")

	if rc := generated.Rkllm_destroy(handles[0]); rc != 0 {
		fmt.Fprintf(os.Stderr, "destroy failed (rc=%d)\n", rc)
		os.Exit(1)
	}
}
