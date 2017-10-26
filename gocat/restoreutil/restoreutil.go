package restoreutil

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"strings"
)

// StructSize is the size of the hashcat restore structure as defined here
// https://hashcat.net/wiki/doku.php?id=restore
const StructSize = 283

// RestoreData defines hashcat's .restore file
type RestoreData struct {
	// Version is the hashcat version used to create the file
	Version uint32
	// WorkingDirectory is the current working directory that hashcat was in at the time of this restore point.
	// Hashcat will change to this directory.
	WorkingDirectory string
	// DictionaryPosition is the current poisition within the dictionary
	DictionaryPosition uint32
	// MasksPosition is the current position within the list of masks
	MasksPosition uint32
	// WordsPosition is the position within the dictionary/mask
	WordsPosition uint64
	// ArgCount is the number of command line arguments passed into hashcat
	ArgCount uint32

	ArgvPointer uint64
	// Args contains the command line arguments
	Args []string
}

func (s *RestoreData) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &s.Version); err != nil {
		return err
	}

	bwd := make([]byte, 256)
	for i, r := range s.WorkingDirectory {
		bwd[i] = byte(r)
	}

	if _, err := w.Write(bwd); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, &s.DictionaryPosition); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, &s.MasksPosition); err != nil {
		return err
	}

	// padding
	if _, err := w.Write(make([]byte, 0x4)); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, &s.WordsPosition); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, &s.ArgCount); err != nil {
		return err
	}

	// padding
	if _, err := w.Write(make([]byte, 0x4)); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, &s.ArgvPointer); err != nil {
		return err
	}

	for _, arg := range s.Args {
		barg := []byte(arg)
		if !bytes.HasSuffix(barg, []byte("\n")) {
			barg = append(barg, 10)
		}
		if _, err := w.Write(barg); err != nil {
			return err
		}
	}

	return nil
}

func restoreParser(f io.ReadSeeker, rd *RestoreData) error {
	if err := binary.Read(f, binary.LittleEndian, &rd.Version); err != nil {
		return err
	}

	buf := make([]byte, 256)
	if err := binary.Read(f, binary.LittleEndian, &buf); err != nil {
		return err
	}
	rd.WorkingDirectory = string(bytes.Trim(buf, "\x00"))

	if err := binary.Read(f, binary.LittleEndian, &rd.DictionaryPosition); err != nil {
		return err
	}

	if err := binary.Read(f, binary.LittleEndian, &rd.MasksPosition); err != nil {
		return err
	}

	// there's 4 bytes of padding here in the struct
	if _, err := f.Seek(0x4, os.SEEK_CUR); err != nil {
		return err
	}

	if err := binary.Read(f, binary.LittleEndian, &rd.WordsPosition); err != nil {
		return err
	}

	if err := binary.Read(f, binary.LittleEndian, &rd.ArgCount); err != nil {
		return err
	}

	// more padding...
	if _, err := f.Seek(0x4, os.SEEK_CUR); err != nil {
		return err
	}

	if err := binary.Read(f, binary.LittleEndian, &rd.ArgvPointer); err != nil {
		return err
	}

	rdr := bufio.NewReader(f)
	arg, err := rdr.ReadString('\n')
	for err != io.EOF {
		rd.Args = append(rd.Args, strings.TrimSpace(arg))
		arg, err = rdr.ReadString('\n')
	}
	return nil
}

// ReadRestoreFile reads the restore file from fp
func ReadRestoreFile(fp string) (rd RestoreData, err error) {
	f, err := os.Open(fp)
	if err != nil {
		return rd, err
	}
	defer f.Close()

	err = restoreParser(f, &rd)
	return
}

// ReadRestoreBytes reads the restore file from the bytes passed in
func ReadRestoreBytes(b []byte) (rd RestoreData, err error) {
	rdr := bytes.NewReader(b)
	err = restoreParser(rdr, &rd)
	return
}
