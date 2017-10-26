// +build windows

package gocat

// patchEventMutex frees and updates hashcat's event mutex with a recursive one
// that allows an event callback to call another event callback without a deadlock condition.
func patchEventMutex(ctx interface{}) (patched bool, err error) {
	// EnterCriticalSection on windows already allows for reentry into a critical section if called by the same thread
	// see remarks @ https://msdn.microsoft.com/en-us/library/windows/desktop/ms682608(v=vs.85).aspx
	return true, nil
}
