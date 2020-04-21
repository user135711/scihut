package lib

import (
	"io"
	"sync"

	"github.com/anacrolix/torrent"
)

type ReaderAt struct {
	r torrent.Reader
	m sync.Mutex
}

func NewReaderAt(r torrent.Reader) *ReaderAt {
	ra := new(ReaderAt)
	ra.r = r
	return ra
}

func (ra *ReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	ra.m.Lock()
	defer ra.m.Unlock()
	if _, err = ra.r.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}
	return ra.r.Read(p)
}
