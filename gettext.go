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
		st := strs[i]
		tr := trans[i]

		if st.length == 0 || tr.length == 0 {
			continue
		}

		key := string(readSection(f, st.offset, st.length))
		val := string(readSection(f, tr.offset, tr.length))

		dict[key] = val
	}
	return dict, nil
}

func parseHeader(f io.ReadSeeker, st state) header {
	var h header
	// skip file format revision
	if _, err := f.Seek(4, io.SeekCurrent); err != nil {
		panic(err)
	}

	var buf uint32
	if err := binary.Read(f, st.ByteOrder, &buf); err != nil {
		panic(err)
	}
	h.n = int64(buf)

	if err := binary.Read(f, st.ByteOrder, &buf); err != nil {
		panic(err)
	}
	h.o = int64(buf)

	if err := binary.Read(f, st.ByteOrder, &buf); err != nil {
		panic(err)
	}
	h.t = int64(buf)
	return h
}

func parseDescriptor(f io.ReaderAt, st state, frm, nos int64) []nthString {
	nthStrings := make([]nthString, nos)
	for i := 0; i < int(nos); i++ {
		sec := readSection(f, frm+int64(i*8), 8)
		length := int64(st.ByteOrder.Uint32(sec[:4]))
		offset := int64(st.ByteOrder.Uint32(sec[4:]))
		nthStrings[i] = nthString{length, offset}
	}
	return nthStrings
}

func readSection(ra io.ReaderAt, offset, length int64) []byte {
	buf := make([]byte, length)
	if _, err := ra.ReadAt(buf, offset); err != nil {
		panic(err)
	}
	return buf
}
