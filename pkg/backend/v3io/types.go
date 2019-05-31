package v3io

import (
	"encoding/xml"
	"os"
	"time"
)

type DataSource interface {
	Connect() error
	Disconnect() error
	ListDir(paths []string) (*FileInfoIterator, error)
	Scan(paths []string, modifiedAfterTime time.Time) (*FileInfoIterator, error)
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

type ListBucketResult struct {
	XMLName     xml.Name `xml:"ListBucketResult"`
	Name        string   `xml:"Name"`
	Prefix      string   `xml:"Prefix"`
	Marker      string   `xml:"Marker"`
	Delimiter   string   `xml:"Delimiter"`
	NextMarker  string   `xml:"NextMarker"`
	MaxKeys     string   `xml:"MaxKeys"`
	IsTruncated string   `xml:"IsTruncated"`
	Contents    []Contents
}

type Contents struct {
	Key          string `xml:"Key"`
	Size         int32  `xml:"Size"`
	LastModified string `xml:"LastModified"`
}
