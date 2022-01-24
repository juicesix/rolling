package rolling

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type RollingFile struct {
	mu sync.Mutex

	closed    bool
	exit      chan struct{}
	syncFlush chan struct{}

	file       *os.File
	current    *bytes.Buffer
	fullBuffer chan *bytes.Buffer

	basePath string
	filePath string
	fileFrag string

	rollMutex sync.RWMutex
	rolling   RollingFormat
}

type RollingFormat string

const (
	MonthlyRolling  RollingFormat = "200601"
	DailyRolling                  = "20060102"
	HourlyRolling                 = "2006010215"
	MinutelyRolling               = "200601021504"
	SecondlyRolling               = "20060102150405"
)

func (r *RollingFile) SetRolling(fmt RollingFormat) {
	r.rollMutex.Lock()
	r.rolling = fmt
	r.rollMutex.Unlock()
}

func (r *RollingFile) roll() error {
	r.rollMutex.RLock()
	roll := r.rolling
	r.rollMutex.RUnlock()
	suffix := time.Now().Format(string(roll))
	if r.file != nil {
		if suffix == r.fileFrag {
			return nil
		}
		r.file.Close()
		r.file = nil
	}
	r.fileFrag = suffix
	dir, filename := filepath.Split(r.basePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	if r.fileFrag == "" {
		r.filePath = filepath.Join(dir, filename+".log")
	} else {
		r.filePath = filepath.Join(dir, filename+"-"+r.fileFrag+".log")
	}
	f, err := os.OpenFile(r.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	r.file = f
	return nil
}
