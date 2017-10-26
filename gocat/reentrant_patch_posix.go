// +build linux,cgo darwin,cgo

package gocat

/*
#include <pthread.h>
#include "wrapper.h"

extern int patch_event_ctx(hashcat_ctx_t *hashcat_ctx)
{
  int rc = -1;

  pthread_mutex_t pMutex;
  pthread_mutexattr_t pAttr;

  event_ctx_t *event_ctx = hashcat_ctx->event_ctx;

  hc_thread_mutex_delete(event_ctx->mux_event);
  rc = pthread_mutexattr_init(&pAttr);
  if (rc != 0)
    goto finished;

  pthread_mutexattr_settype(&pAttr, PTHREAD_MUTEX_RECURSIVE);

  rc = pthread_mutex_init(&pMutex, &pAttr);
  if (rc != 0)
    goto finished;

  event_ctx->mux_event = pMutex;

finished:
  return rc;
}
*/
import "C"
import "errors"

var errReentrantPatch = errors.New("failed to patch hashcat_ctx->event_ctx mutex")

// patchEventMutex frees and updates hashcat's event mutex with a recursive one
// that allows an event callback to call another event callback without a deadlock condition
// NOTE: this only works on posix systems (linux, darwin)
func patchEventMutex(ctx C.hashcat_ctx_t) (patched bool, err error) {
	if retval := C.patch_event_ctx(&ctx); retval != 0 {
		return false, errReentrantPatch
	}
	return true, nil
}
