package gocat

// #include "wrapper.h"
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

// ValidationResult is the output from ValidateHashes and includes information about the hash file
type ValidationResult struct {
	Valid           bool
	Errors          []string
	NumHashes       uint32
	NumHashesUnique uint32
	NumSalts        uint32
}

const errStr = "linter: %s failed with rv %d"

// ValidateHashes is a linter that validates hashes before creating and executing a hashcat session
func ValidateHashes(pathToHashes string, hashType uint32) (*ValidationResult, error) {
	var err error
	var hashes *C.hashes_t
	vr := &ValidationResult{
		Valid: true,
	}

	hashpath := C.CString(pathToHashes)
	defer C.free(unsafe.Pointer(hashpath))

	validator := C.gocat_ctx_t{
		ctx:             C.hashcat_ctx_t{},
		gowrapper:       unsafe.Pointer(vr),
		bValidateHashes: true,
	}

	if retval := C.hashcat_init(&validator.ctx, (*[0]byte)(unsafe.Pointer(C.event))); retval != 0 {
		err = fmt.Errorf(errStr, "hashcat_init", retval)
		goto cleanup
	}

	if retval := C.user_options_init(&validator.ctx); retval != 0 {
		err = fmt.Errorf(errStr, "user_options_init", retval)
		goto cleanup
	}

	validator.ctx.user_options.hash_mode = C.u32(hashType)
	validator.ctx.user_options_extra.hc_hash = hashpath

	if retval := C.hashconfig_init(&validator.ctx); retval != 0 {
		err = fmt.Errorf(errStr, "hashconfig_init", retval)
		goto cleanup
	}

	hashes = validator.ctx.hashes
	// Load hashes
	if retval := C.hashes_init_stage1(&validator.ctx); retval != 0 {
		err = fmt.Errorf(errStr, "hashes_init_stage1", retval)
		goto cleanup
	}

	// Removes duplicates
	hashes.hashes_cnt_orig = hashes.hashes_cnt
	if retval := C.hashes_init_stage2(&validator.ctx); retval != 0 {
		err = fmt.Errorf(errStr, "hashes_init_stage2", retval)
		goto cleanup
	}

cleanup:
	if validator.ctx.hashes != nil {
		vr.NumHashes = uint32(validator.ctx.hashes.hashes_cnt_orig)
		vr.NumHashesUnique = uint32(validator.ctx.hashes.digests_cnt)
		vr.NumSalts = uint32(validator.ctx.hashes.salts_cnt)
	}

	if &validator.ctx != nil {
		C.hashcat_destroy(&validator.ctx)
	}

	return vr, err
}

//export validatorCallback
func validatorCallback(id uint32, hcCtx *C.hashcat_ctx_t, results unsafe.Pointer, buf unsafe.Pointer, len C.size_t) {
	var r = (*ValidationResult)(results)

	switch id {
	case C.EVENT_LOG_WARNING:
		ectx := hcCtx.event_ctx
		msg := C.GoStringN(&ectx.msg_buf[0], C.int(ectx.msg_len))
		if strings.Contains(msg, "kernel not found") || strings.Contains(msg, "falling back to") {
			return
		}

		if strings.Contains(msg, "Hashfile") && strings.Contains(msg, "on ") {
			// strip out the filename in the warning as it's unnecessary for our purposes
			onIndex := strings.Index(msg, "on ")
			msg = msg[onIndex+3:]
		}

		r.Valid = false
		r.Errors = append(r.Errors, msg)
	}
}
