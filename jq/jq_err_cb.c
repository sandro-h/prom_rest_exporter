#include <stdio.h>

#include <jv.h>
#include <jq.h>

#include "_cgo_export.h"

static inline void callGoErrorHandler(void *data, jv it) {
	// goJqErrorHandler is an exported go function in jq.go
	goJqErrorHandler((GoUint64)data, it);
}

void install_jq_error_cb(jq_state *jq, GoUint64 id) {
	jq_set_error_cb(jq, callGoErrorHandler, (void*)id);
}