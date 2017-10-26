package child

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nightlyone/lockfile"
	"github.com/rs/zerolog/log"
)

func checkFileHash(path string) (string, error) {
	fd, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	sh := sha1.New()

	if _, err := io.Copy(sh, fd); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sh.Sum(nil)), nil
}

func acquireLock(forFilePath string) (*lockfile.Lockfile, error) {
	lckPath := fmt.Sprintf("%s.lck", forFilePath)
	theLock, err := lockfile.New(lckPath)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("lockfile", lckPath).Msg("Acquiring file lock")

	err = theLock.TryLock()
	if err == nil {
		goto AcquiredLock
	}

	for {
		switch err {
		case lockfile.ErrBusy, lockfile.ErrNotExist:
			time.Sleep(5 * time.Second)
			if err = theLock.TryLock(); err == nil {
				// gotem!
				goto AcquiredLock
			}
			log.Debug().Str("lockfile", lckPath).Msg("Waiting for file lock...")
		default:
			// cant lock the file at this point :(
			return nil, err
		}
	}

AcquiredLock:
	log.Debug().Str("lockfile", lckPath).Msg("Acquired file lock")
	return &theLock, nil
}
