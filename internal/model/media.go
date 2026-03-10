package model

import (
	"database/sql/driver"
	"errors"

	"github.com/bytedance/sonic"
	"github.com/dustin/go-humanize"
)

type MediaItem struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Hash        string `json:"hash"`
}

func (m *MediaItem) SizeHumanized() string {
	return humanize.Bytes(uint64(m.Size))
}

func (m *MediaItem) Value() (driver.Value, error) {
	if m == nil || *m == (MediaItem{}) {
		return nil, nil
	}
	b, err := sonic.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (m *MediaItem) Scan(value interface{}) error {
	if value == nil {
		*m = MediaItem{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("unsupported type for MediaItem")
	}

	return sonic.Unmarshal(b, m)
}
