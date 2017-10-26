package gocat

/*
#cgo CFLAGS: -I/usr/local/include/hashcat -std=c99 -Wall -O0 -g
#cgo linux CFLAGS: -D_GNU_SOURCE
#cgo LDFLAGS: -L/usr/local/lib -lhashcat

#include "wrapper.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unsafe"

	"github.com/fireeye/gocrack/gocat/hcargp"
)

var (
	// CompileTime should be set using -ldflags "-X" during compilation. This value is passed into hashcat_session_init
	CompileTime = time.Now().UTC().Unix()
	// ErrUnableToStopAtCheckpoint is raised whenever hashcat is unable to stop at next checkpoint. This is caused when
	// the device status != STATUS_RUNNING or --restore-disable is set
	ErrUnableToStopAtCheckpoint = errors.New("gocat: unable to stop at next checkpoint")
	// ErrUnableToStop is raised whenever we were unable to stop a hashcat session. At this time, hashcat's hashcat_session_quit
	// always returns success so you'll most likely never see this.
	ErrUnableToStop = errors.New("gocat: unable to stop task")
)

const (
	// sessionCracked indicates all hashes were cracked
	sessionCracked int = iota
	// sessionExhausted indicates all possible permutations were reached in the session
	sessionExhausted
	// sessionQuit indicates the session was quit by the user
	sessionQuit
	// sessionAborted indicates the session was aborted by the user
	sessionAborted
	// sessionAbortedCheckpoint indicates the session was checkpointed by the user (usually due to an abort signal or temperature limit)
	sessionAbortedCheckpoint
	// sessionAbortedRuntime indicates the session was stopped by hashcat (usually due to a runtime limit)
	sessionAbortedRuntime
)

// EventCallback defines the callback that hashcat/gocat calls
type EventCallback func(Hashcat unsafe.Pointer, Payload interface{})

// Options defines all the configuration options for gocat/hashcat
type Options struct {
	// SharedPath should be set to the location where OpenCL kernels and .hcstat/hctune files live
	SharedPath string
	// ExecutablePath should be set to the location of the binary and not the binary itself.
	// Hashcat does some weird logic and will ignore the shared path if this is incorrectly set.
	// If ExecutablePath is not set, we'll attempt to calculate it using os.Args[0]
	ExecutablePath string
	// PatchEventContext if true will patch hashcat's event_ctx with a reentrant lock which will allow
	// you to call several hashcat APIs (which fire another callback) from within an event callback.
	// This is supported on macOS, Linux, and Windows.
	PatchEventContext bool
}

// ErrNoSharedPath is raised whenever Options.SharedPath is not set
var ErrNoSharedPath = errors.New("shared path must be set")

func (o *Options) validate() (err error) {
	if o.SharedPath == "" {
		return ErrNoSharedPath
	}

	if o.ExecutablePath == "" {
		if o.ExecutablePath, _ = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			return err
		}
	} else {
		if _, err = os.Stat(o.ExecutablePath); err != nil {
			return err
		}
	}

	return nil
}

// Hashcat is the interface which interfaces with libhashcat to provide password cracking capabilities.
type Hashcat struct {
	// unexported fields below
	wrapper        C.gocat_ctx_t
	cb             EventCallback
	opts           Options
	isEventPatched bool
	l              sync.Mutex

	// these must be free'd
	executablePath *C.char
	sharedPath     *C.char
}

// New creates a context that can be reused to crack passwords.
func New(opts Options, cb EventCallback) (hc *Hashcat, err error) {
	if err = opts.validate(); err != nil {
		return nil, err
	}

	hc = &Hashcat{
		executablePath: C.CString(opts.ExecutablePath),
		sharedPath:     C.CString(opts.SharedPath),
		cb:             cb,
		opts:           opts,
	}

	hc.wrapper = C.gocat_ctx_t{
		ctx:             C.hashcat_ctx_t{},
		gowrapper:       unsafe.Pointer(hc),
		bValidateHashes: false,
	}

	if retval := C.hashcat_init(&hc.wrapper.ctx, (*[0]byte)(unsafe.Pointer(C.event))); retval != 0 {
		return
	}

	return
}

// EventCallbackIsReentrant returns a boolean indicating if hashcat_ctx.event_ctx has been patched to allow an event to fire another event
func (hc *Hashcat) EventCallbackIsReentrant() bool {
	return hc.isEventPatched
}

// RunJob starts a hashcat session and blocks until it has been finished.
func (hc *Hashcat) RunJob(args ...string) (err error) {
	hc.l.Lock()
	defer hc.l.Unlock()

	// initialize the default options in hashcat_ctx->user_options
	if retval := C.user_options_init(&hc.wrapper.ctx); retval != 0 {
		return
	}

	argc, argv := convertArgsToC(append([]string{hc.opts.ExecutablePath}, args...)...)
	defer C.freeargv(argc, argv)

	if retval := C.user_options_getopt(&hc.wrapper.ctx, argc, argv); retval != 0 {
		return getErrorFromCtx(hc.wrapper.ctx)
	}

	if retval := C.user_options_sanity(&hc.wrapper.ctx); retval != 0 {
		return getErrorFromCtx(hc.wrapper.ctx)
	}

	if retval := C.hashcat_session_init(&hc.wrapper.ctx, hc.executablePath, hc.sharedPath, argc, argv, C.int(CompileTime)); retval != 0 {
		return getErrorFromCtx(hc.wrapper.ctx)
	}
	defer C.hashcat_session_destroy(&hc.wrapper.ctx)

	if hc.opts.PatchEventContext {
		isPatchSuccessful, err := patchEventMutex(hc.wrapper.ctx)
		if err != nil {
			return err
		}
		hc.isEventPatched = isPatchSuccessful
	}

	rc := C.hashcat_session_execute(&hc.wrapper.ctx)
	switch int(rc) {
	case sessionCracked, sessionExhausted, sessionQuit, sessionAborted,
		sessionAbortedCheckpoint, sessionAbortedRuntime:
		err = nil
	default:
		return getErrorFromCtx(hc.wrapper.ctx)
	}

	return
}

// RunJobWithOptions is a convenience function to take a HashcatSessionOptions struct and craft the necessary argvs to use
// for the hashcat session.
// This is NOT goroutine safe. If you are needing to run multiple jobs, create a context for each one.
func (hc *Hashcat) RunJobWithOptions(opts hcargp.HashcatSessionOptions) error {
	args, err := opts.MarshalArgs()
	if err != nil {
		return err
	}
	return hc.RunJob(args...)
}

// StopAtCheckpoint instructs the running hashcat session to stop at the next available checkpoint
func (hc *Hashcat) StopAtCheckpoint() error {
	if retval := C.hashcat_session_checkpoint(&hc.wrapper.ctx); retval != 0 {
		return ErrUnableToStopAtCheckpoint
	}
	return nil
}

// AbortRunningTask instructs hashcat to abruptly stop the running session
func (hc *Hashcat) AbortRunningTask() {
	C.hashcat_session_quit(&hc.wrapper.ctx)
}

func getErrorFromCtx(ctx C.hashcat_ctx_t) error {
	msg := C.hashcat_get_log(&ctx)
	return fmt.Errorf("gocat: %s", C.GoString(msg))
}

//export callback
func callback(id uint32, hcCtx *C.hashcat_ctx_t, wrapper unsafe.Pointer, buf unsafe.Pointer, len C.size_t) {
	ctx := (*Hashcat)(wrapper)

	var payload interface{}
	var err error

	switch id {
	case C.EVENT_LOG_ERROR:
		payload = logMessageCbFromEvent(hcCtx, InfoMessage)
	case C.EVENT_LOG_INFO:
		payload = logMessageCbFromEvent(hcCtx, InfoMessage)
	case C.EVENT_LOG_WARNING:
		payload = logMessageCbFromEvent(hcCtx, WarnMessage)
	case C.EVENT_BITMAP_INIT_PRE:
		payload = logHashcatAction(id, "Generating bitmap tables")
	case C.EVENT_BITMAP_INIT_POST:
		payload = logHashcatAction(id, "Generated bitmap tables")
	case C.EVENT_CALCULATED_WORDS_BASE:
		if hcCtx.user_options.keyspace {
			payload = logHashcatAction(id, fmt.Sprintf("Calculated Words Base: %d", hcCtx.status_ctx.words_base))
		}
	case C.EVENT_WEAK_HASH_PRE:
		payload = logHashcatAction(id, "Checking for weak hashes")
	case C.EVENT_WEAK_HASH_POST:
		payload = logHashcatAction(id, "Checked for weak hashes")
	case C.EVENT_HASHLIST_SORT_SALT_PRE:
		payload = logHashcatAction(id, "Sorting salts...")
	case C.EVENT_HASHLIST_SORT_SALT_POST:
		payload = logHashcatAction(id, "Sorted salts...")
	case C.EVENT_OPENCL_SESSION_PRE:
		payload = logHashcatAction(id, "Initializing device kernels and memory")
	case C.EVENT_OPENCL_SESSION_POST:
		payload = logHashcatAction(id, "Initialized device kernels and memory")
	case C.EVENT_AUTOTUNE_STARTING:
		payload = logHashcatAction(id, "Starting Autotune threads")
	case C.EVENT_AUTOTUNE_FINISHED:
		payload = logHashcatAction(id, "Autotune threads have started..")
	case C.EVENT_OUTERLOOP_MAINSCREEN:
		hashes := ctx.wrapper.ctx.hashes
		payload = TaskInformationPayload{
			NumHashes:       uint32(hashes.hashes_cnt_orig),
			NumHashesUnique: uint32(hashes.digests_cnt),
			NumSalts:        uint32(hashes.salts_cnt),
		}
	case C.EVENT_MONITOR_PERFORMANCE_HINT:
		payload = logHashcatAction(id, "Device performance might be suffering due to a less than optimal configuration")
	case C.EVENT_POTFILE_REMOVE_PARSE_PRE:
		payload = logHashcatAction(id, "Comparing hashes with potfile entries...")
	case C.EVENT_POTFILE_REMOVE_PARSE_POST:
		payload = logHashcatAction(id, "Compared hashes with potfile entries")
	case C.EVENT_POTFILE_ALL_CRACKED:
		payload = logHashcatAction(id, "All hashes exist in potfile")
		if ctx.isEventPatched {
			C.potfile_handle_show(&ctx.wrapper.ctx)
		}
		payload = FinalStatusPayload{
			Status:           nil,
			EndedAt:          time.Now().UTC(),
			AllHashesCracked: true,
		}
	case C.EVENT_SET_KERNEL_POWER_FINAL:
		payload = logHashcatAction(id, "Approaching final keyspace, workload adjusted")
	case C.EVENT_POTFILE_NUM_CRACKED:
		ctxHashes := hcCtx.hashes
		if ctxHashes.digests_done > 0 {
			payload = logHashcatAction(id, fmt.Sprintf("Removed %d hash(s) found in potfile", ctxHashes.digests_done))
		}
	case C.EVENT_CRACKER_HASH_CRACKED, C.EVENT_POTFILE_HASH_SHOW:
		// Grab the separator for this session out of user options
		sepr := C.GoString(&hcCtx.user_options.separator)
		msg := C.GoString((*C.char)(buf))
		if payload, err = getCrackedPassword(id, msg, sepr); err != nil {
			payload = logMessageWithError(id, err)
		}
	case C.EVENT_OUTERLOOP_FINISHED:
		payload = FinalStatusPayload{
			Status:  ctx.GetStatus(),
			EndedAt: time.Now().UTC(),
		}
	case C.EVENT_WEAK_HASH_ALL_CRACKED:
		payload = FinalStatusPayload{
			Status:           nil,
			EndedAt:          time.Now().UTC(),
			AllHashesCracked: true,
		}
	}

	// Events we're ignoring:
	// EVENT_CRACKER_STARTING
	// EVENT_OUTERLOOP_MAINSCREEN

	ctx.cb(unsafe.Pointer(hcCtx), payload)
}

// Free releases all allocations. Call this when you're done with hashcat or exiting the application
func (hc *Hashcat) Free() {
	C.hashcat_destroy(&hc.wrapper.ctx)
	C.free(unsafe.Pointer(hc.executablePath))
	C.free(unsafe.Pointer(hc.sharedPath))
}
