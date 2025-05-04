#ifndef _RKLLM_INPUT_WRAPPERS_H_
#define _RKLLM_INPUT_WRAPPERS_H_

#include "rkllm.h"

static inline const char* RKLLMInput_GetPrompt(RKLLMInput* in) {
  return in->prompt_input;
}
static inline void RKLLMInput_SetPrompt(RKLLMInput* in, const char* s) {
  in->prompt_input = s;
}

static inline RKLLMEmbedInput RKLLMInput_GetEmbed(RKLLMInput* in) {
  return in->embed_input;
}
static inline void RKLLMInput_SetEmbed(RKLLMInput* in, RKLLMEmbedInput e) {
  in->embed_input = e;
}

static inline RKLLMTokenInput RKLLMInput_GetToken(RKLLMInput* in) {
  return in->token_input;
}
static inline void RKLLMInput_SetToken(RKLLMInput* in, RKLLMTokenInput t) {
  in->token_input = t;
}

static inline RKLLMMultiModelInput RKLLMInput_GetMultimodal(RKLLMInput* in) {
  return in->multimodal_input;
}
static inline void RKLLMInput_SetMultimodal(RKLLMInput* in, RKLLMMultiModelInput m) {
  in->multimodal_input = m;
}

#endif
