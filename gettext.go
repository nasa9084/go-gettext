package gettext

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	magicNumber = "950412de"
	rebmunCigam = "de120495"
)

type state struct {
	ByteOrder binary.ByteOrder
}

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

type header struct {
	n int64
	o int64
	t int64
}

type nthString struct {
	length int64
	offset int64
}

// New returns new locale object
func New(loc string, opts ...func(*locale)) Locale {
	l := &locale{
		locale: loc,
		dict:   map[string]string{},
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Path option for locale files' dir
func Path(path string) func(loc *locale) {
	return func(loc *locale) {
		loc.path = path
	}
}

func (loc *locale) Load() error {
	moFile, err := os.Open(filepath.Join(loc.path, loc.locale+".mo"))
	if err != nil {
		return errors.Wrap(err, `opening mo file`)
	}
	loc.dict, err = parse(moFile)
	if err != nil {
		return errors.Wrap(err, `parsing mo file`)
	}
	return nil
}

func (loc *locale) SetPath(path string) {
	loc.path = path
}

// Get translation
func (loc *locale) Get(s string) string {
	v, ok := loc.dict[s]
	if !ok {
		return s
	}
	return v
}

func parse(f *os.File) (map[string]string, error) {
	st := state{}
	mn := make([]byte, 4)
	if _, err := f.Read(mn); err != nil {
		panic(err)
	}
	st.ByteOrder = getEndian(mn)

	h := parseHeader(f, st)
	strs := parseDescriptor(f, st, h.o, h.n)
	trans := parseDescriptor(f, st, h.t, h.n)

	dict := map[string]string{}
	for i := 0; i < len(strs); i++ {
		if strs[i].length == 0 || trans[i].length == 0 {
			continue
		}

		key := readSectionToString(f, strs[i].offset, strs[i].length)
		val := readSectionToString(f, trans[i].offset, trans[i].length)

		dict[key] = val
	}
	return dict, nil
}

func parseHeader(rs io.ReadSeeker, st state) header {
	buf := make([]byte, 28)
	if _, err := rs.Read(buf); err != nil {
		panic(err)
	}
	return header{
		n: int64(st.ByteOrder.Uint32(buf[4:8])),
		o: int64(st.ByteOrder.Uint32(buf[8:12])),
		t: int64(st.ByteOrder.Uint32(buf[12:16])),
	}
}

func parseDescriptor(f io.ReaderAt, st state, frm, nos int64) []nthString {
	nthStrings := make([]nthString, nos)
	for i := 0; i < int(nos); i++ {
		sec := readSection(f, frm+int64(i*8), 8)
		length := bytesToInt64(sec[:4], st)
		offset := bytesToInt64(sec[4:], st)
		nthStrings[i] = nthString{length, offset}
	}
	return nthStrings
}

func bytesToInt64(b []byte, st state) int64 {
	return int64(st.ByteOrder.Uint32(b))
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
