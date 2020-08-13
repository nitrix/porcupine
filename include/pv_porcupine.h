#ifndef PV_PORCUPINE_H
#define PV_PORCUPINE_H

#include <stdbool.h>
#include <stdint.h>

#include "picovoice.h"

typedef struct pv_porcupine pv_porcupine_t;

PV_API pv_status_t pv_porcupine_init(const char *model_path, int32_t num_keywords, const char * const * keyword_paths, const float *sensitivities, pv_porcupine_t **object);
PV_API void pv_porcupine_delete(pv_porcupine_t *object);
PV_API pv_status_t pv_porcupine_process(pv_porcupine_t *object, const int16_t *pcm, int32_t *keyword_index);
PV_API const char *pv_porcupine_version(void);
PV_API int32_t pv_porcupine_frame_length(void);

#endif
