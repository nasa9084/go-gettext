package gettext

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

const (
	magicNumber = "950412de"
	rebmunCigam = "de120495"
)

const (
	defaultPath = "locale"
)

func getEndian(mn []byte) binary.ByteOrder {
	switch hex.EncodeToString(mn) {
	case magicNumber:
		return binary.BigEndian
	case rebmunCigam:
		return binary.LittleEndian
	default:
		panic(nil)
	}
}

// implementation of Locale
type locale struct {
	locale string
	path   string
	dict   map[string]string
}

// New returns new locale object
func New(loc string, opts ...func(*locale)) Locale {
	l := &locale{
		locale: loc,
		path:   defaultPath,
		dict:   map[string]string{},
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Path option for locale files' dir
func Path(path string) func(*locale) {
	return func(loc *locale) {
		loc.path = path
	}
}

func (loc *locale) Load() error {
	moFile, err := os.Open(filepath.Join(loc.path, loc.locale+".mo"))
	if err != nil {
		return err
	}
	m := newMo(moFile)
	loc.dict, err = m.parse()
	if err != nil {
		return err
	}
	return nil
}

// Get translation
func (loc *locale) Get(s string) string {
	v, ok := loc.dict[s]
	if !ok {
		return s
	}
	return v
}

func bytesToInt64(b []byte, bo binary.ByteOrder) int64 {
	return int64(bo.Uint32(b))
}

func readSection(ra io.ReaderAt, offset, length int64) []byte {
	buf := make([]byte, length)
	if _, err := ra.ReadAt(buf, offset); err != nil {
		panic(err)
	}
	return buf
}

func readSectionToString(ra io.ReaderAt, offset, length int64) string {
	return string(readSection(ra, offset, length))
}
