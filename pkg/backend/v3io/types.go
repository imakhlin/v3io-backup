package v3io

import (
	"os"
	"time"
)

type DataSource interface {
	Connect() error
	Disconnect() error
	ListDir(path string) (*FileInfoIterator, error)
	Scan(path string, modifiedAfterTime time.Time) (*FileInfoIterator, error)
}

type FileInfo struct {
	baseInfo           *os.FileInfo
	extendedAttributes map[string]interface{}
}

type FileInfoIterator interface {
	Next() *FileInfo
	At() *FileInfo
	Error() error
}
