package shared

// HashcatAttackMode describes the various supported password cracking attacks available in hashcat
type HashcatAttackMode uint32

const (
	// AttackModeStraight is a dictionary attack with optional mangling rules applied
	AttackModeStraight HashcatAttackMode = 0
	// AttackModeBruteForce is a brute force attack using a list of masks to guess the password(s)
	AttackModeBruteForce HashcatAttackMode = 3
)

// HashcatUserOptions defines the user settable options of a hashcat task
type HashcatUserOptions struct {
	AttackMode       HashcatAttackMode `json:"attack_mode"`
	HashType         int               `json:"hash_type"`
	Masks            *string           `json:"masks,omitempty"`
	DictionaryFile   *string           `json:"dictionary_file,omitempty"`
	ManglingRuleFile *string           `json:"mangling_file,omitempty"`
}

// HModeInfo describes the hashcat mode
type HModeInfo struct {
	Number  int    `json:"mode"`
	Name    string `json:"name"`
	Example string `json:"example,omitempty"`
}
