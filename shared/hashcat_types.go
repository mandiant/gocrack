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
	Number   int    `json:"mode"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Example  string `json:"example,omitempty"`
}

// SupportedHashcatModes is a list of all the supported hashcat cracking SupportedHashcatModes
// This is manually generated and is up-to-date as of v3.40
var SupportedHashcatModes = []HModeInfo{
	{
		Number:   900,
		Name:     "MD4",
		Category: "Raw Hash",
	},
	{
		Number:   0,
		Name:     "MD5",
		Category: "Raw Hash",
		Example:  "8743b52063cd84097a65d1633f5c74f5",
	},
	{
		Number:   5100,
		Name:     "Half MD5",
		Category: "Raw Hash",
	},
	{
		Number:   100,
		Name:     "SHA1",
		Category: "Raw Hash",
		Example:  "b89eaac7e61417341b710b727768294d0e6a277b",
	},
	{
		Number:   1300,
		Name:     "SHA-224",
		Category: "Raw Hash",
	},
	{
		Number:   1400,
		Name:     "SHA-256",
		Category: "Raw Hash",
	},
	{
		Number:   10800,
		Name:     "SHA-384",
		Category: "Raw Hash",
	},
	{
		Number:   1700,
		Name:     "SHA-512",
		Category: "Raw Hash",
	},
	{
		Number:   17300,
		Name:     "SHA3-224",
		Category: "Raw Hash",
	},
	{
		Number:   17400,
		Name:     "SHA3-256",
		Category: "Raw Hash",
	},
	{
		Number:   17500,
		Name:     "SHA3-384",
		Category: "Raw Hash",
	},
	{
		Number:   17600,
		Name:     "SHA3-512",
		Category: "Raw Hash",
	},
	{
		Number:   17700,
		Name:     "Keccak-224",
		Category: "Raw Hash",
	},
	{
		Number:   17800,
		Name:     "Keccak-256",
		Category: "Raw Hash",
	},
	{
		Number:   17900,
		Name:     "Keccak-384",
		Category: "Raw Hash",
	},
	{
		Number:   18000,
		Name:     "Keccak-512",
		Category: "Raw Hash",
	},
	{
		Number:   600,
		Name:     "BLAKE2b-512",
		Category: "Raw Hash",
	},
	{
		Number:   10100,
		Name:     "SipHash",
		Category: "Raw Hash",
	},
	{
		Number:   6000,
		Name:     "RipeMD160",
		Category: "Raw Hash",
	},
	{
		Number:   6100,
		Name:     "Whirlpool",
		Category: "Raw Hash",
	},
	{
		Number:   6900,
		Name:     "GOST R 34.11-94",
		Category: "Raw Hash",
	},
	{
		Number:   11700,
		Name:     "GOST R 34.11-2012 (Streebog) 256-bit, big-endian",
		Category: "Raw Hash",
	},
	{
		Number:   11800,
		Name:     "GOST R 34.11-2012 (Streebog) 512-bit, big-endian",
		Category: "Raw Hash",
	},
	{
		Number:   10,
		Name:     "md5($pass.$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
		Example:  "01dfae6e5d4d90d9892622325959afbe:7050461",
	},
	{
		Number:   20,
		Name:     "md5($salt.$pass)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   30,
		Name:     "md5(unicode($pass).$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   40,
		Name:     "md5($salt.unicode($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   3800,
		Name:     "md5($salt.$pass.$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   3710,
		Name:     "md5($salt.md5($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   4010,
		Name:     "md5($salt.md5($salt.$pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   4110,
		Name:     "md5($salt.md5($pass.$salt))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   2600,
		Name:     "md5(md5($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   3910,
		Name:     "md5(md5($pass).md5($salt))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   4300,
		Name:     "md5(strtoupper(md5($pass)))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   4400,
		Name:     "md5(sha1($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   110,
		Name:     "sha1($pass.$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   120,
		Name:     "sha1($salt.$pass)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   130,
		Name:     "sha1(unicode($pass).$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   140,
		Name:     "sha1($salt.unicode($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   4500,
		Name:     "sha1(sha1($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   4520,
		Name:     "sha1($salt.sha1($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   4700,
		Name:     "sha1(md5($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   4900,
		Name:     "sha1($salt.$pass.$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   14400,
		Name:     "sha1(CX)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   1410,
		Name:     "sha256($pass.$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   1420,
		Name:     "sha256($salt.$pass)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   1430,
		Name:     "sha256(unicode($pass).$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   1440,
		Name:     "sha256($salt.unicode($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   1710,
		Name:     "sha512($pass.$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   1720,
		Name:     "sha512($salt.$pass)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   1730,
		Name:     "sha512(unicode($pass).$salt)",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   1740,
		Name:     "sha512($salt.unicode($pass))",
		Category: "Raw Hash, Salted and / or Iterated",
	},
	{
		Number:   50,
		Name:     "HMAC-MD5 (key = $pass)",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   60,
		Name:     "HMAC-MD5 (key = $salt)",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   150,
		Name:     "HMAC-SHA1 (key = $pass)",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   160,
		Name:     "HMAC-SHA1 (key = $salt)",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   1450,
		Name:     "HMAC-SHA256 (key = $pass)",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   1460,
		Name:     "HMAC-SHA256 (key = $salt)",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   1750,
		Name:     "HMAC-SHA512 (key = $pass)",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   1760,
		Name:     "HMAC-SHA512 (key = $salt)",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   11750,
		Name:     "HMAC-Streebog-256 (key = $pass), big-endian",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   11760,
		Name:     "HMAC-Streebog-256 (key = $salt), big-endian",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   11850,
		Name:     "HMAC-Streebog-512 (key = $pass), big-endian",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   11860,
		Name:     "HMAC-Streebog-512 (key = $salt), big-endian",
		Category: "Raw Hash, Authenticated",
	},
	{
		Number:   14000,
		Name:     "DES (PT = $salt, key = $pass)",
		Category: "Raw Cipher, Known-Plaintext attack",
	},
	{
		Number:   14100,
		Name:     "3DES (PT = $salt, key = $pass)",
		Category: "Raw Cipher, Known-Plaintext attack",
	},
	{
		Number:   14900,
		Name:     "Skip32 (PT = $salt, key = $pass)",
		Category: "Raw Cipher, Known-Plaintext attack",
	},
	{
		Number:   400,
		Name:     "phpass",
		Category: "Generic KDF",
	},
	{
		Number:   8900,
		Name:     "scrypt",
		Category: "Generic KDF",
	},
	{
		Number:   11900,
		Name:     "PBKDF2-HMAC-MD5",
		Category: "Generic KDF",
	},
	{
		Number:   12000,
		Name:     "PBKDF2-HMAC-SHA1",
		Category: "Generic KDF",
	},
	{
		Number:   10900,
		Name:     "PBKDF2-HMAC-SHA256",
		Category: "Generic KDF",
	},
	{
		Number:   12100,
		Name:     "PBKDF2-HMAC-SHA512",
		Category: "Generic KDF",
	},
	{
		Number:   23,
		Name:     "Skype",
		Category: "Network protocols",
	},
	{
		Number:   2500,
		Name:     "WPA-EAPOL-PBKDF2",
		Category: "Network protocols",
	},
	{
		Number:   2501,
		Name:     "WPA-EAPOL-PMK",
		Category: "Network protocols",
	},
	{
		Number:   16800,
		Name:     "WPA-PMKID-PBKDF2",
		Category: "Network protocols",
	},
	{
		Number:   16801,
		Name:     "WPA-PMKID-PMK",
		Category: "Network protocols",
	},
	{
		Number:   4800,
		Name:     "iSCSI CHAP authentication, MD5(Chap)",
		Category: "Network protocols",
	},
	{
		Number:   5300,
		Name:     "IKE-PSK MD5",
		Category: "Network protocols",
	},
	{
		Number:   5400,
		Name:     "IKE-PSK SHA1",
		Category: "Network protocols",
	},
	{
		Number:   5500,
		Name:     "NetNTLMv1 / NetNTLMv1 + ESS",
		Category: "Network protocols",
		Example:  "u4-netntlm::kNS:338d08f8e26de93300000000000000000000000000000000:9526fb8c23a90751cdd619b6cea564742e1e4bf33006ba41:cb8086049ec4736c",
	},
	{
		Number:   5600,
		Name:     "NetNTLMv2",
		Category: "Network protocols",
		Example:  "admin::N46iSNekpT:08ca45b7d7ea58ee:88dcbe4446168966a153a0064958dac6:5c7830315c7830310000000000000b45c67103d07d7b95acd12ffa11230e0000000052920b85f78d013c31cdb3b92f5d765c783030",
	},
	{
		Number:   7300,
		Name:     "IPMI2 RAKP HMAC-SHA1",
		Category: "Network protocols",
	},
	{
		Number:   7500,
		Name:     "Kerberos 5 AS-REQ Pre-Auth etype 23",
		Category: "Network protocols",
	},
	{
		Number:   8300,
		Name:     "DNSSEC (NSEC3)",
		Category: "Network protocols",
	},
	{
		Number:   10200,
		Name:     "Cram MD5",
		Category: "Network protocols",
	},
	{
		Number:   11100,
		Name:     "PostgreSQL CRAM (MD5)",
		Category: "Network protocols",
	},
	{
		Number:   11200,
		Name:     "MySQL CRAM (SHA1)",
		Category: "Network protocols",
	},
	{
		Number:   11400,
		Name:     "SIP digest authentication (MD5)",
		Category: "Network protocols",
	},
	{
		Number:   13100,
		Name:     "Kerberos 5 TGS-REP etype 23",
		Category: "Network protocols",
		Example:  "$krb5tgs$23$*user$realm$test/spn*$63386d22d359fe42230300d56852c9eb$891ad31d09ab89c6b3b8c5e5de6c06a7f49fd559d7a9a3c32576c8fedf705376cea582ab5938f7fc8bc741acf05c5990741b36ef4311fe3562a41b70a4ec6ecba849905f2385bb3799d92499909658c7287c49160276bca0006c350b0db4fd387adc27c01e9e9ad0c20ed53a7e6356dee2452e35eca2a6a1d1432796fc5c19d068978df74d3d0baf35c77de12456bf1144b6a750d11f55805f5a16ece2975246e2d026dce997fba34ac8757312e9e4e6272de35e20d52fb668c5ed",
	},
	{
		Number:   16100,
		Name:     "TACACS+",
		Category: "Network protocols",
	},
	{
		Number:   16500,
		Name:     "JWT (JSON Web Token)",
		Category: "Network protocols",
	},
	{
		Number:   18200,
		Name:     "Kerberos 5 AS-REP etype 23",
		Category: "Network protocols",
	},
	{
		Number:   121,
		Name:     "SMF (Simple Machines Forum)",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   400,
		Name:     "phpBB3",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   2611,
		Name:     "vBulletin < v3.8.5",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   2711,
		Name:     "vBulletin > v3.8.5",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   2811,
		Name:     "MyBB",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   2811,
		Name:     "IPB (Invison Power Board)",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   8400,
		Name:     "WBB3 (Woltlab Burning Board)",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   11,
		Name:     "Joomla < 2.5.18",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   400,
		Name:     "Joomla > 2.5.18",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   400,
		Name:     "Wordpress",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   2612,
		Name:     "PHPS",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   7900,
		Name:     "Drupal7",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   21,
		Name:     "osCommerce",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   21,
		Name:     "xt:Commerce",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   11000,
		Name:     "PrestaShop",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   124,
		Name:     "Django (SHA-1)",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   10000,
		Name:     "Django (PBKDF2-SHA256)",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   16000,
		Name:     "Tripcode",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   3711,
		Name:     "Mediawiki B type",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   13900,
		Name:     "OpenCart",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   4521,
		Name:     "Redmine",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   4522,
		Name:     "PunBB",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   12001,
		Name:     "Atlassian (PBKDF2-HMAC-SHA1)",
		Category: "Forums, CMS, E-Commerce, Frameworks",
	},
	{
		Number:   12,
		Name:     "PostgreSQL",
		Category: "Database Server",
	},
	{
		Number:   131,
		Name:     "MSSQL(2000)",
		Category: "Database Server",
	},
	{
		Number:   132,
		Name:     "MSSQL(2005)",
		Category: "Database Server",
	},
	{
		Number:   1731,
		Name:     "MSSQL(2012)",
		Category: "Database Server",
	},
	{
		Number:   1731,
		Name:     "MSSQL(2014)",
		Category: "Database Server",
	},
	{
		Number:   200,
		Name:     "MySQL323",
		Category: "Database Server",
	},
	{
		Number:   300,
		Name:     "MySQL4.1/MySQL5",
		Category: "Database Server",
	},
	{
		Number:   3100,
		Name:     "Oracle H: Type (Oracle 7+)",
		Category: "Database Server",
	},
	{
		Number:   112,
		Name:     "Oracle S: Type (Oracle 11+)",
		Category: "Database Server",
	},
	{
		Number:   12300,
		Name:     "Oracle T: Type (Oracle 12+)",
		Category: "Database Server",
	},
	{
		Number:   8000,
		Name:     "Sybase ASE",
		Category: "Database Server",
	},
	{
		Number:   141,
		Name:     "EPiServer 6.x < v4",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   1441,
		Name:     "EPiServer 6.x > v4",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   1600,
		Name:     "Apache $apr1$",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   12600,
		Name:     "ColdFusion 10+",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   1421,
		Name:     "hMailServer",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   101,
		Name:     "nsldap, SHA-1(Base64), Netscape LDAP SHA",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   111,
		Name:     "nsldaps, SSHA-1(Base64), Netscape LDAP SSHA",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   1411,
		Name:     "SSHA-256(Base64), LDAP {SSHA256}",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   1711,
		Name:     "SSHA-512(Base64), LDAP {SSHA512}",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   16400,
		Name:     "CRAM-MD5 Dovecot",
		Category: "HTTP, SMTP, LDAP Server",
	},
	{
		Number:   15000,
		Name:     "FileZilla Server >= 0.9.55",
		Category: "FTP Server",
	},
	{
		Number:   11500,
		Name:     "CRC32",
		Category: "Checksums",
	},
	{
		Number:   3000,
		Name:     "LM",
		Category: "Operating-Systems",
		Example:  "299bd128c1101fd6",
	},
	{
		Number:   1000,
		Name:     "NTLM",
		Category: "Operating-Systems",
		Example:  "b4b9b02e6f09a9bd760f388b67351e2b",
	},
	{
		Number:   1100,
		Name:     "Domain Cached Credentials (DCC), MS Cache",
		Category: "Operating-Systems",
	},
	{
		Number:   2100,
		Name:     "Domain Cached Credentials 2 (DCC2), MS Cache 2",
		Category: "Operating-Systems",
	},
	{
		Number:   12800,
		Name:     "MS-AzureSync PBKDF2-HMAC-SHA256",
		Category: "Operating-Systems",
	},
	{
		Number:   1500,
		Name:     "descrypt, DES(Unix), Traditional DES",
		Category: "Operating-Systems",
	},
	{
		Number:   12400,
		Name:     "BSDiCrypt, Extended DES",
		Category: "Operating-Systems",
	},
	{
		Number:   500,
		Name:     "md5crypt $1$, MD5(Unix)",
		Category: "Operating-Systems",
	},
	{
		Number:   3200,
		Name:     "bcrypt $2*$, Blowfish(Unix)",
		Category: "Operating-Systems",
		Example:  "$2a$05$LhayLxezLhK1LhWvKxCyLOj0j1u.Kj0jZ0pEmm134uzrQlFvQJLF6",
	},
	{
		Number:   7400,
		Name:     "sha256crypt $5$, SHA256(Unix)",
		Category: "Operating-Systems",
	},
	{
		Number:   1800,
		Name:     "sha512crypt $6$, SHA512(Unix)",
		Category: "Operating-Systems",
	},
	{
		Number:   122,
		Name:     "macOS v10.4, MacOS v10.5, MacOS v10.6",
		Category: "Operating-Systems",
	},
	{
		Number:   1722,
		Name:     "macOS v10.7",
		Category: "Operating-Systems",
	},
	{
		Number:   7100,
		Name:     "macOS v10.8+ (PBKDF2-SHA512)",
		Category: "Operating-Systems",
	},
	{
		Number:   6300,
		Name:     "AIX {smd5}",
		Category: "Operating-Systems",
	},
	{
		Number:   6700,
		Name:     "AIX {ssha1}",
		Category: "Operating-Systems",
	},
	{
		Number:   6400,
		Name:     "AIX {ssha256}",
		Category: "Operating-Systems",
	},
	{
		Number:   6500,
		Name:     "AIX {ssha512}",
		Category: "Operating-Systems",
	},
	{
		Number:   2400,
		Name:     "Cisco-PIX",
		Category: "Operating-Systems",
	},
	{
		Number:   2410,
		Name:     "Cisco-ASA",
		Category: "Operating-Systems",
	},
	{
		Number:   500,
		Name:     "Cisco-IOS $1$",
		Category: "Operating-Systems",
	},
	{
		Number:   5700,
		Name:     "Cisco-IOS $4$",
		Category: "Operating-Systems",
	},
	{
		Number:   9200,
		Name:     "Cisco-IOS $8$",
		Category: "Operating-Systems",
	},
	{
		Number:   9300,
		Name:     "Cisco-IOS $9$",
		Category: "Operating-Systems",
	},
	{
		Number:   22,
		Name:     "Juniper Netscreen/SSG (ScreenOS)",
		Category: "Operating-Systems",
	},
	{
		Number:   501,
		Name:     "Juniper IVE",
		Category: "Operating-Systems",
	},
	{
		Number:   7000,
		Name:     "Fortigate (FortiOS)",
		Category: "Operating-Systems",
	},
	{
		Number:   5800,
		Name:     "Android PIN",
		Category: "Operating-Systems",
	},
	{
		Number:   13800,
		Name:     "Windows 8+ phone PIN/Password",
		Category: "Operating-Systems",
	},
	{
		Number:   8100,
		Name:     "Citrix Netscaler",
		Category: "Operating-Systems",
	},
	{
		Number:   8500,
		Name:     "RACF",
		Category: "Operating-Systems",
	},
	{
		Number:   7200,
		Name:     "GRUB 2",
		Category: "Operating-Systems",
	},
	{
		Number:   9900,
		Name:     "Radmin2",
		Category: "Operating-Systems",
	},
	{
		Number:   125,
		Name:     "ArubaOS",
		Category: "Operating-Systems",
	},
	{
		Number:   7700,
		Name:     "SAP CODVN B (BCODE)",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   7701,
		Name:     "SAP CODVN B (BCODE) via RFC_READ_TABLE",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   7800,
		Name:     "SAP CODVN F/G (PASSCODE)",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   7801,
		Name:     "SAP CODVN F/G (PASSCODE) via RFC_READ_TABLE",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   10300,
		Name:     "SAP CODVN H (PWDSALTEDHASH) iSSHA-1",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   8600,
		Name:     "Lotus Notes/Domino 5",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   8700,
		Name:     "Lotus Notes/Domino 6",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   9100,
		Name:     "Lotus Notes/Domino 8",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   133,
		Name:     "PeopleSoft",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   13500,
		Name:     "PeopleSoft Token",
		Category: "Enterprise Application Software (EAS)",
	},
	{
		Number:   11600,
		Name:     "7-Zip",
		Category: "Archives",
	},
	{
		Number:   12500,
		Name:     "RAR3-hp",
		Category: "Archives",
	},
	{
		Number:   13000,
		Name:     "RAR5",
		Category: "Archives",
	},
	{
		Number:   13200,
		Name:     "AxCrypt",
		Category: "Archives",
	},
	{
		Number:   13300,
		Name:     "AxCrypt in memory SHA1",
		Category: "Archives",
	},
	{
		Number:   13600,
		Name:     "WinZip",
		Category: "Archives",
	},
	{
		Number:   14700,
		Name:     "iTunes Backup < 10.0",
		Category: "Backup",
	},
	{
		Number:   14800,
		Name:     "iTunes Backup >= 10.0",
		Category: "Backup",
	},
	{
		Number:   8800,
		Name:     "Android FDE < v4.3",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   12900,
		Name:     "Android FDE (Samsung DEK)",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   12200,
		Name:     "eCryptfs",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   14600,
		Name:     "LUKS",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   16700,
		Name:     "FileVault 2",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   18300,
		Name:     "Apple File System (APFS)",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   9700,
		Name:     "MS Office <= 2003 $0",
		Category: "$1, MD5 + RC4               | Documents",
	},
	{
		Number:   9710,
		Name:     "MS Office <= 2003 $0",
		Category: "$1, MD5 + RC4, collider #1  | Documents",
	},
	{
		Number:   9720,
		Name:     "MS Office <= 2003 $0",
		Category: "$1, MD5 + RC4, collider #2  | Documents",
	},
	{
		Number:   9800,
		Name:     "MS Office <= 2003 $3",
		Category: "$4, SHA1 + RC4              | Documents",
	},
	{
		Number:   9810,
		Name:     "MS Office <= 2003 $3",
		Category: "$4, SHA1 + RC4, collider #1 | Documents",
	},
	{
		Number:   9820,
		Name:     "MS Office <= 2003 $3",
		Category: "$4, SHA1 + RC4, collider #2 | Documents",
	},
	{
		Number:   9400,
		Name:     "MS Office 2007",
		Category: "Documents",
	},
	{
		Number:   9500,
		Name:     "MS Office 2010",
		Category: "Documents",
	},
	{
		Number:   9600,
		Name:     "MS Office 2013",
		Category: "Documents",
	},
	{
		Number:   10400,
		Name:     "PDF 1.1 - 1.3 (Acrobat 2 - 4)",
		Category: "Documents",
	},
	{
		Number:   10410,
		Name:     "PDF 1.1 - 1.3 (Acrobat 2 - 4), collider #1",
		Category: "Documents",
	},
	{
		Number:   10420,
		Name:     "PDF 1.1 - 1.3 (Acrobat 2 - 4), collider #2",
		Category: "Documents",
	},
	{
		Number:   10500,
		Name:     "PDF 1.4 - 1.6 (Acrobat 5 - 8)",
		Category: "Documents",
	},
	{
		Number:   10600,
		Name:     "PDF 1.7 Level 3 (Acrobat 9)",
		Category: "Documents",
	},
	{
		Number:   10700,
		Name:     "PDF 1.7 Level 8 (Acrobat 10 - 11)",
		Category: "Documents",
	},
	{
		Number:   16200,
		Name:     "Apple Secure Notes",
		Category: "Documents",
	},
	{
		Number:   9000,
		Name:     "Password Safe v2",
		Category: "Password Managers",
	},
	{
		Number:   5200,
		Name:     "Password Safe v3",
		Category: "Password Managers",
	},
	{
		Number:   6800,
		Name:     "Lastpass + Lastpass sniffed",
		Category: "Password Managers",
	},
	{
		Number:   6600,
		Name:     "1Password, agilekeychain",
		Category: "Password Managers",
	},
	{
		Number:   8200,
		Name:     "1Password, cloudkeychain",
		Category: "Password Managers",
	},
	{
		Number:   11300,
		Name:     "Bitcoin/Litecoin wallet.dat",
		Category: "Password Managers",
	},
	{
		Number:   12700,
		Name:     "Blockchain, My Wallet",
		Category: "Password Managers",
	},
	{
		Number:   15200,
		Name:     "Blockchain, My Wallet V2",
		Category: "Password Managers",
	},
	{
		Number:   16600,
		Name:     "Electrum Wallet (Salt-Type 1-3)",
		Category: "Password Managers",
	},
	{
		Number:   13400,
		Name:     "Keepass 1 (AES/Twofish) and Keepass 2 (AES)",
		Category: "Password Managers",
	},
	{
		Number:   16300,
		Name:     "Ethereum Pre-Sale Wallet, PBKDF2-HMAC-SHA256",
		Category: "Password Managers",
	},
	{
		Number:   16900,
		Name:     "Ansible Vault",
		Category: "Password Managers",
	},
	{
		Number:   18100,
		Name:     "TOTP (HMAC-SHA1)",
		Category: "One-Time Passwords",
	},
	{
		Number:   99999,
		Name:     "Plaintext",
		Category: "Plaintext",
	},
	{
		Number:   6211,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 512 bit pure AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6211,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 512 bit pure Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6211,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 512 bit pure Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6212,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 1024 bit pure AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6212,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 1024 bit pure Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6212,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 1024 bit pure Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6212,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 1024 bit cascaded AES-Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6212,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 1024 bit cascaded Serpent-AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6212,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 1024 bit cascaded Twofish-Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6213,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 XTS 1536 bit all",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6221,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 512 bit pure AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6221,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 512 bit pure Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6221,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 512 bit pure Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6222,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 1024 bit pure AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6222,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 1024 bit pure Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6222,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 1024 bit pure Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6222,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 1024 bit cascaded AES-Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6222,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 1024 bit cascaded Serpent-AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6222,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 1024 bit cascaded Twofish-Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6223,
		Name:     "TrueCrypt PBKDF2-HMAC-SHA512 XTS 1536 bit all",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6231,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 512 bit pure AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6231,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 512 bit pure Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6231,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 512 bit pure Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6232,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 1024 bit pure AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6232,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 1024 bit pure Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6232,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 1024 bit pure Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6232,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 1024 bit cascaded AES-Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6232,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 1024 bit cascaded Serpent-AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6232,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 1024 bit cascaded Twofish-Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6233,
		Name:     "TrueCrypt PBKDF2-HMAC-Whirlpool XTS 1536 bit all",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6241,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 512 bit pure AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6241,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 512 bit pure Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6241,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 512 bit pure Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6242,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 1024 bit pure AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6242,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 1024 bit pure Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6242,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 1024 bit pure Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6242,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 1024 bit cascaded AES-Twofish",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6242,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 1024 bit cascaded Serpent-AES",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6242,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 1024 bit cascaded Twofish-Serpent",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   6243,
		Name:     "TrueCrypt PBKDF2-HMAC-RipeMD160 + boot-mode XTS 1536 bit all",
		Category: "Full-Disk encryptions (FDE)",
	},
	{
		Number:   15300,
		Name:     "DPAPI masterkey file v1",
		Category: "Operating Systems",
	},
	{
		Number:   15900,
		Name:     "DPAPI masterkey file v2",
		Category: "Operating Systems",
	},
}

// LookupHashcatHashType takes a hash type integer and returns info about it if it exists
func LookupHashcatHashType(hashtype int) *HModeInfo {
	for _, h := range SupportedHashcatModes {
		if h.Number == hashtype {
			return &h
		}
	}
	return nil
}
