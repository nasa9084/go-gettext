package gettext

import (
	"encoding/binary"
	"os"
)

const headerLength = 28

type pos struct {
	length int64
	offset int64
}

type header struct {
	n int64
	o int64
	t int64
}

type mo struct {
	rawFile       *os.File
	byteOrder     binary.ByteOrder
	header        header
	stringPoss    []pos
	translatePoss []pos
}

func (m *mo) parse() (map[string]string, error) {
	mn := make([]byte, 4)
	if _, err := m.rawFile.Read(mn); err != nil {
		return nil, err
	}
	m.byteOrder = getEndian(mn)
	if err := m.parseHeader(); err != nil {
		return nil, err
	}
	if err := m.parseStrings(); err != nil {
		return nil, err
	}
	if err := m.parseTranslates(); err != nil {
		return nil, err
	}
	return m.genDict()
}

func (m *mo) parseHeader() error {
	buf := make([]byte, headerLength)
	if _, err := m.rawFile.Read(buf); err != nil {
		return err
	}
	m.header = header{
		n: int64(m.byteOrder.Uint32(buf[4:8])),
		o: int64(m.byteOrder.Uint32(buf[8:12])),
		t: int64(m.byteOrder.Uint32(buf[12:16])),
	}
	return nil
}

func (m *mo) parseStrings() error {
	poss, err := m.parseDescriptor(m.header.o)
	if err != nil {
		return err
	}
	m.stringPoss = poss
	return nil
}

func (m *mo) parseTranslates() error {
	poss, err := m.parseDescriptor(m.header.t)
	if err != nil {
		return err
	}
	m.translatePoss = poss
	return nil
}

func (m *mo) parseDescriptor(frm int64) ([]pos, error) {
	poss := make([]pos, m.header.n)
	for i := 0; i < int(m.header.n); i++ {
		sec := readSection(m.rawFile, frm+int64(i*8), 8)
		poss[i] = pos{
			length: bytesToInt64(sec[:4], m.byteOrder),
			offset: bytesToInt64(sec[4:], m.byteOrder),
		}
	}
	return poss, nil
}

func (m *mo) genDict() (map[string]string, error) {
	dict := map[string]string{}
	for i := 0; i < len(m.stringPoss); i++ {
		if m.stringPoss[i].length == 0 || m.translatePoss[i].length == 0 {
			continue
		}

		key := readSectionToString(m.rawFile, m.stringPoss[i].offset, m.stringPoss[i].length)
		val := readSectionToString(m.rawFile, m.translatePoss[i].offset, m.translatePoss[i].length)

		dict[key] = val
	}
	return dict, nil
}
