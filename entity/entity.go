package entity

import (
	"fmt"
	"time"
)

type FileByteInfo struct {
	Filename   string
	Bytes      []byte
	Categories []string
	Tags       []string
	CoverImg   string
	PubDate    string
}
type FileContentInfo struct {
	Filename   string
	Content    string
	Categories []string
	Tags       []string
	CoverImg   string
	PubDate    string
}
type HtmlDTO struct {
	PageTitle string
	Data      interface{}
}
type PostDTO struct {
	ID         int64
	Title      string
	Content    string
	CreatedAt  JsonTime
	UpdatedAt  JsonTime
	Deleted    int16
	PubDate    JsonTime
	CoverImg   string
	Categories []string
	Tags       []string
}
type JsonTime time.Time

func (t JsonTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}
