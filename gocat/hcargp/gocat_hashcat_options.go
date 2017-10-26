package hcargp

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// GetStringPtr returns the pointer of s
func GetStringPtr(s string) *string {
	return &s
}

// GetIntPtr returns the pointer of i
func GetIntPtr(i int) *int {
	return &i
}

// GetBoolPtr returns the pointer of b
func GetBoolPtr(b bool) *bool {
	return &b
}

/*
We skip the following arguments because they are not needed:
- version
- help
- quiet
- status
- status-timer
- machine-readable
- stdout
- show
- left
- benchmark
- speed-only (todo?)
- progress-only (todo?)
- opencl-info
- keyspace
*/

// HashcatSessionOptions represents all the available hashcat options. The values here should always follow the latest version of hashcat
type HashcatSessionOptions struct {
	HashType               *int    `hashcat:"--hash-type,omitempty"`
	AttackMode             *int    `hashcat:"--attack-mode,omitempty"`
	IsHexCharset           *bool   `hashcat:"--hex-charset,omitempty"`
	IsHexSalt              *bool   `hashcat:"--hex-salt,omitempty"`
	IsHexWordlist          *bool   `hashcat:"--hex-wordlist,omitempty"`
	KeepGuessing           *bool   `hashcat:"--keep-guessing,omitempty"`
	Loopback               *bool   `hashcat:"--loopback,omitempty"`
	WeakHashThreshold      *int    `hashcat:"--weak-hash-threshold,omitempty"`
	MarkovHCStat           *string `hashcat:"--markov-hcstat,omitempty"`
	DisableMarkov          *bool   `hashcat:"--markov-disable,omitempty"`
	EnableClassicMarkov    *bool   `hashcat:"--markov-classic,omitempty"`
	MarkovThreshold        *int    `hashcat:"--markov-threshold,omitempty"`
	Force                  *bool   `hashcat:"--force,omitempty"`
	MaxRuntimeSeconds      *int    `hashcat:"--runtime,omitempty"`
	SessionName            *string `hashcat:"--session,omitempty"`
	RestoreSession         *bool   `hashcat:"--restore,omitempty"`
	DisableRestore         *bool   `hashcat:"--restore-disable,omitempty"`
	RestoreFilePath        *string `hashcat:"--restore-file-path,omitempty"`
	OutfilePath            *string `hashcat:"--outfile,omitempty"`
	OutfileFormat          *int    `hashcat:"--outfile-format,omitempty"`
	OutfileDisableAutoHex  *bool   `hashcat:"--outfile-autohex-disable,omitempty"`
	OutfileCheckTimer      *int    `hashcat:"--outfile-check-timer,omitempty"`
	Separator              *string `hashcat:"--separator,omitempty"`
	IgnoreUsername         *bool   `hashcat:"--username,omitempty"`
	RemoveCrackedHash      *bool   `hashcat:"--remove,omitempty"`
	RemoveCrackedHashTimer *int    `hashcat:"--remove-timer,omitempty"`
	PotfileDisable         *bool   `hashcat:"--potfile-disable,omitempty"`
	PotfilePath            *string `hashcat:"--potfile-path,omitempty"`
	DebugMode              *int    `hashcat:"--debug-mode,omitempty"`
	DebugFile              *string `hashcat:"--debug-file,omitempty"`
	InductionDir           *string `hashcat:"--induction-dir,omitempty"`
	LogfileDisable         *bool   `hashcat:"--logfile-disable,omitempty"`
	HccapxMessagePair      *string `hashcat:"--hccapx-message-pair,omitempty"`
	NonceErrorCorrections  *int    `hashcat:"--nonce-error-corrections,omitempty"`
	TrueCryptKeyFiles      *string `hashcat:"--truecrypt-keyfiles,omitempty"`
	VeraCryptKeyFiles      *string `hashcat:"--veracrypt-keyfiles,omitempty"`
	VeraCryptPIM           *int    `hashcat:"--veracrypt-pim,omitempty"`
	SegmentSize            *int    `hashcat:"--segment-size,omitempty"`
	BitmapMin              *int    `hashcat:"--bitmap-min,omitempty"`
	BitmapMax              *int    `hashcat:"--bitmap-max,omitempty"`
	CPUAffinity            *string `hashcat:"--cpu-affinity,omitempty"`
	OpenCLPlatforms        *string `hashcat:"--opencl-platforms,omitempty"`
	OpenCLDevices          *string `hashcat:"--opencl-devices,omitempty"`
	OpenCLDeviceTypes      *string `hashcat:"--opencl-device-types,omitempty"`
	OpenCLVectorWidth      *string `hashcat:"--opencl-vector-width,omitempty"`
	WorkloadProfile        *int    `hashcat:"--workload-profile,omitempty"`
	KernelAccel            *int    `hashcat:"--kernel-accel,omitempty"`
	KernelLoops            *int    `hashcat:"--kernel-loops,omitempty"`
	NVIDIASpinDamp         *int    `hashcat:"--nvidia-spin-damp,omitempty"`
	GPUTempDisable         *bool   `hashcat:"--gpu-temp-disable,omitempty"`
	GPUTempAbort           *int    `hashcat:"--gpu-temp-abort,omitempty"`
	GPUTempRetain          *int    `hashcat:"--gpu-temp-retain,omitempty"`
	PowertuneEnable        *bool   `hashcat:"--powertune-enable,omitempty"`
	ScryptTMTO             *int    `hashcat:"--scrypt-tmto,omitempty"`
	Skip                   *int    `hashcat:"--skip,omitempty"`
	Limit                  *int    `hashcat:"--limit,omitempty"`
	RuleLeft               *string `hashcat:"--rule-left,omitempty"`
	RuleRight              *string `hashcat:"--rule-right,omitempty"`
	RulesFile              *string `hashcat:"--rules-file,omitempty"`
	GenerateRules          *int    `hashcat:"--generate-rules,omitempty"`
	GenerateRulesFuncMin   *int    `hashcat:"--generate-rules-func-min,omitempty"`
	GenerateRulesFuncMax   *int    `hashcat:"--generate-rules-func-max,omitempty"`
	GenerateRulesSeed      *int    `hashcat:"--generate-rules-seed,omitempty"`
	CustomCharset1         *string `hashcat:"--custom-charset1,omitempty"`
	CustomCharset2         *string `hashcat:"--custom-charset2,omitempty"`
	CustomCharset3         *string `hashcat:"--custom-charset3,omitempty"`
	CustomCharset4         *string `hashcat:"--custom-charset4,omitempty"`
	IncrementMask          *bool   `hashcat:"--increment,omitempty"`
	IncrementMaskMin       *int    `hashcat:"--increment-min,omitempty"`
	IncrementMaskMax       *int    `hashcat:"--increment-max,omitempty"`

	// InputFile can be a single hash or multiple hashes via a hashfile or hccapx
	InputFile                    string  `hashcat:","`
	DictionaryMaskDirectoryInput *string `hashcat:",omitempty"`
}

func parseTag(t string) (tag, options string) {
	if idx := strings.Index(t, ","); idx != -1 {
		return t[:idx], t[idx+1:]
	}
	return tag, ""
}

// MarshalArgs returns a list of arguments set by the user to be passed into hashcat's session for execution
func (o HashcatSessionOptions) MarshalArgs() (args []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}

			err = r.(error)
		}
	}()

	v := reflect.ValueOf(o)
	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("hashcat")
		if tag == "" {
			continue
		}

		name, opts := parseTag(tag)
		val := v.Field(i)

		hasOmitEmpty := strings.Contains(opts, "omitempty")
		if (val.Type().Kind() == reflect.Ptr && val.IsNil()) && hasOmitEmpty {
			continue
		}

		if val.Type().Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}

		switch val.Type().Kind() {
		case reflect.Bool:
			if val.Bool() {
				args = append(args, name)
			}
		case reflect.Int:
			// Int's should always have a name...
			if name != "" {
				args = append(args, fmt.Sprintf("%s=%d", name, val.Int()))
			}
		case reflect.String:
			if val.String() == "" {
				continue
			}

			if name != "" {
				args = append(args, fmt.Sprintf("%s=%s", name, val.String()))
			} else {
				args = append(args, val.String())
			}
		default:
			err = fmt.Errorf("unknown type %s", val.Type().Kind())
			return
		}
	}
	return
}
