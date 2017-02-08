package file

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"github.com/liuxp0827/govpr/log"
	"io"
	"os"
)

type VPRFile struct {
	file       *os.File
	readwriter *bufio.ReadWriter
}

func NewVPRFile(filename string) (*VPRFile, error) {

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}

	greader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}

	gwriter := gzip.NewWriter(file)
	vprFile := &VPRFile{
		file:       file,
		readwriter: bufio.NewReadWriter(bufio.NewReader(greader), bufio.NewWriter(gwriter)),
	}

	//vprFile := &VPRFile{
	//	file:       file,
	//	readwriter: bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file)),
	//}

	return vprFile, nil
}

func (f *VPRFile) PutInt(v int) (int, error) {
	var intBuf [4]byte

	data := intBuf[:4]
	data[0] = byte(v & 0xff)
	data[1] = byte((v >> 8) & 0xff)
	data[2] = byte((v >> 16) & 0xff)
	data[3] = byte((v >> 24) & 0xff)
	return f.readwriter.Write(data)
}

func (f *VPRFile) PutByte(v byte) error {
	return f.readwriter.WriteByte(v)
}

func (f *VPRFile) PutFloat64(v float64) (int, error) {
	var float64Buf [8]byte
	data := float64Buf[:8]
	PutFloat64LE(data, v)
	return f.readwriter.Write(data)
}

func (f *VPRFile) GetInt() (int, error) {

	var v uint32
	binary.Read(f.readwriter, binary.LittleEndian, &v)
	return int(v), nil
}

func (f *VPRFile) GetByte() (byte, error) {
	return f.readwriter.ReadByte()
}

func (f *VPRFile) GetFloat64() (float64, error) {
	var float64Buf [8]byte

	data := float64Buf[:8]
	_, err := io.ReadFull(f.readwriter, data)
	if err != nil {
		log.Error(err)
		return .0, err
	}

	return GetFloat64LE(data), nil
}

func (f *VPRFile) GetFloat32() (float32, error) {
	var floatBuf [4]byte

	data := floatBuf[:4]
	_, err := io.ReadFull(f.readwriter, data)
	if err != nil {
		log.Error(err)
		return .0, err
	}

	return GetFloat32LE(data), nil
}

func (f *VPRFile) Close() error {
	err := f.readwriter.Flush()
	if err != nil {
		return err
	}
	return f.file.Close()
}
