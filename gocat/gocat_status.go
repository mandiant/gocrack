package gocat

// #include "wrapper.h"
import "C"
import (
	"fmt"
	"unsafe"
)

// DeviceStatus contains information about the OpenCL device that is cracking
type DeviceStatus struct {
	DeviceID  int
	HashesSec string
	ExecDev   float64
}

// Status contains data about the current cracking session
type Status struct {
	Session               string
	Status                string
	HashType              string
	HashTarget            string
	TimeStarted           string
	TimeEstimated         string
	TimeEstimatedRelative string
	DeviceStatus          []DeviceStatus
	TotalSpeed            string
	ProgressMode          int
	Candidates            map[int]string // map[DeviceID]string
	Progress              string
	Rejected              string
	Recovered             string
	RestorePoint          string
	GuessMode             int
	GuessMask             string `json:",omitempty"`
	GuessQueue            string `json:",omitempty"`
	GuessBase             string `json:",omitempty"`
	GuessMod              string `json:",omitempty"`
	GuessCharset          string `json:",omitempty"`
}

var szStatusStruct = unsafe.Sizeof(C.hashcat_status_t{})

// GetStatus returns the status of the cracking.
// This is an implementation of https://github.com/hashcat/hashcat/blob/master/src/terminal.c#L709
func (hc *Hashcat) GetStatus() *Status {
	ptr := C.malloc(C.size_t(szStatusStruct))
	hcStatus := (*C.hashcat_status_t)(unsafe.Pointer(ptr))
	defer C.free(unsafe.Pointer(hcStatus))

	if retval := C.hashcat_get_status(&hc.wrapper.ctx, hcStatus); retval != 0 {
		return nil
	}
	defer C.status_status_destroy(&hc.wrapper.ctx, hcStatus)

	stats := &Status{
		Session:               C.GoString(hcStatus.session),
		Status:                C.GoString(hcStatus.status_string),
		HashType:              C.GoString(hcStatus.hash_type),
		HashTarget:            C.GoString(hcStatus.hash_target),
		TimeStarted:           C.GoString(hcStatus.time_started_absolute),
		TimeEstimated:         C.GoString(hcStatus.time_estimated_absolute),
		TimeEstimatedRelative: C.GoString(hcStatus.time_estimated_relative),
		// Instead of using hcStatus.device_info_cnt here, we'll let go manage this
		// to avoid having empty DeviceStatus's
		DeviceStatus: make([]DeviceStatus, 0),
		TotalSpeed:   C.GoString(hcStatus.speed_sec_all),
		ProgressMode: int(hcStatus.progress_mode),
		Candidates:   make(map[int]string),
		GuessMode:    int(hcStatus.guess_mode),
	}

	switch stats.ProgressMode {
	case C.PROGRESS_MODE_KEYSPACE_KNOWN:
		stats.Progress = fmt.Sprintf("%d/%d (%.02f%%)", hcStatus.progress_cur_relative_skip,
			hcStatus.progress_end_relative_skip,
			hcStatus.progress_finished_percent)
		stats.Rejected = fmt.Sprintf("%d/%d (%.02f%%)", hcStatus.progress_rejected,
			hcStatus.progress_cur_relative_skip,
			hcStatus.progress_rejected_percent)
		stats.RestorePoint = fmt.Sprintf("%d/%d (%.02f%%)", hcStatus.restore_point,
			hcStatus.restore_total,
			hcStatus.restore_percent)
	case C.PROGRESS_MODE_KEYSPACE_UNKNOWN:
		stats.Progress = fmt.Sprintf("%d", hcStatus.progress_cur_relative_skip)
		stats.Rejected = fmt.Sprintf("%d", hcStatus.progress_rejected)
		stats.RestorePoint = fmt.Sprintf("%d", hcStatus.restore_point)
	}

	switch stats.GuessMode {
	case C.GUESS_MODE_STRAIGHT_FILE:
		stats.GuessBase = C.GoString(hcStatus.guess_base)
	case C.GUESS_MODE_STRAIGHT_FILE_RULES_FILE:
		stats.GuessBase = C.GoString(hcStatus.guess_base)
		stats.GuessMod = C.GoString(hcStatus.guess_mod)
	case C.GUESS_MODE_STRAIGHT_FILE_RULES_GEN:
		stats.GuessBase = C.GoString(hcStatus.guess_base)
		stats.GuessMod = "Rules (Generated)"
	case C.GUESS_MODE_STRAIGHT_STDIN:
		stats.GuessBase = "Pipe"
	case C.GUESS_MODE_STRAIGHT_STDIN_RULES_FILE:
		stats.GuessBase = "Pipe"
		stats.GuessMod = C.GoString(hcStatus.guess_mod)
	case C.GUESS_MODE_STRAIGHT_STDIN_RULES_GEN:
		stats.GuessBase = "Pipe"
		stats.GuessMod = "Rules (Generated)"
	case C.GUESS_MODE_COMBINATOR_BASE_LEFT:
		stats.GuessBase = fmt.Sprintf("File (%s), Left Side", C.GoString(hcStatus.guess_base))
		stats.GuessMod = fmt.Sprintf("File (%s), Right Side", C.GoString(hcStatus.guess_mod))
	case C.GUESS_MODE_COMBINATOR_BASE_RIGHT:
		stats.GuessBase = fmt.Sprintf("File (%s), Right Side", C.GoString(hcStatus.guess_base))
		stats.GuessMod = fmt.Sprintf("File (%s), Left Side", C.GoString(hcStatus.guess_mod))
	case C.GUESS_MODE_MASK:
		stats.GuessMask = fmt.Sprintf("%s [%d]", C.GoString(hcStatus.guess_base), int(hcStatus.guess_mask_length))
	case C.GUESS_MODE_MASK_CS:
		stats.GuessMask = fmt.Sprintf("%s [%d]", C.GoString(hcStatus.guess_base), int(hcStatus.guess_mask_length))
		stats.GuessCharset = C.GoString(hcStatus.guess_charset)
	case C.GUESS_MODE_HYBRID1_CS:
		stats.GuessCharset = C.GoString(hcStatus.guess_charset)
		fallthrough // grab GuessBase/GuessMod from below as it's the same
	case C.GUESS_MODE_HYBRID1:
		stats.GuessBase = fmt.Sprintf("File (%s), Left Side", C.GoString(hcStatus.guess_base))
		stats.GuessMod = fmt.Sprintf("Mask (%s) [%d], Right Side", C.GoString(hcStatus.guess_base), int(hcStatus.guess_mask_length))
	}

	switch stats.GuessMode {
	case C.GUESS_MODE_STRAIGHT_FILE, C.GUESS_MODE_STRAIGHT_FILE_RULES_FILE,
		C.GUESS_MODE_STRAIGHT_FILE_RULES_GEN, C.GUESS_MODE_MASK:
		stats.GuessQueue = fmt.Sprintf("%d/%d (%.02f%%)", hcStatus.guess_base_offset,
			hcStatus.guess_base_count, hcStatus.guess_base_percent)
	case C.GUESS_MODE_HYBRID1, C.GUESS_MODE_HYBRID2:
		stats.GuessQueue = fmt.Sprintf("%d/%d (%.02f%%)", hcStatus.guess_base_offset,
			hcStatus.guess_base_count, hcStatus.guess_base_percent)
	}

	stats.Recovered = fmt.Sprintf("%d/%d (%.2f%%) Digests, %d/%d (%.2f%%) Salts",
		hcStatus.digests_done,
		hcStatus.digests_cnt,
		hcStatus.digests_percent,
		hcStatus.salts_done,
		hcStatus.salts_cnt,
		hcStatus.salts_percent,
	)

	for i := 0; i < int(hcStatus.device_info_cnt); i++ {
		deviceInfo := hcStatus.device_info_buf[i]
		if deviceInfo.skipped_dev {
			continue
		}

		stats.DeviceStatus = append(stats.DeviceStatus, DeviceStatus{
			DeviceID:  i + 1,
			HashesSec: C.GoString(deviceInfo.speed_sec_dev),
			ExecDev:   float64(deviceInfo.exec_msec_dev),
		})

		if deviceInfo.guess_candidates_dev != nil {
			stats.Candidates[i+1] = C.GoString(deviceInfo.guess_candidates_dev)
		}
	}

	return stats
}
