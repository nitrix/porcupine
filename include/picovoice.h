#ifndef PICOVOICE_H
#define PICOVOICE_H

#define PV_API __attribute__((visibility ("default")))

PV_API int pv_sample_rate(void);

typedef enum {
    PV_STATUS_SUCCESS = 0,
    PV_STATUS_OUT_OF_MEMORY,
    PV_STATUS_IO_ERROR,
    PV_STATUS_INVALID_ARGUMENT,
    PV_STATUS_STOP_ITERATION,
    PV_STATUS_KEY_ERROR,
    PV_STATUS_INVALID_STATE,
} pv_status_t;

PV_API const char *pv_status_to_string(pv_status_t status);

#endif
