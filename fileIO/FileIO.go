package fileIO

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"govpr/log"
	"io"
	"os"
)

type FileIO struct {
	Line   string
	file   *os.File
	reader *bufio.Reader
	writer *bufio.Writer
	mode   int // 0:reader, 1:writer
}

func NewFileIO(filename string, mode int) (*FileIO, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	if mode != 0 && mode != 1 {
		return nil, fmt.Errorf("mode only can be 0 or 1")
	}

	fileIO := &FileIO{
		file: file,
		mode: mode,
	}

	if fileIO.mode == 0 {
		fileIO.reader = bufio.NewReader(file)
	} else {
		fileIO.writer = bufio.NewWriter(file)
	}

	return fileIO, nil
}

func NewFileIO2(file *os.File, mode int) (*FileIO, error) {
	if mode != 0 && mode != 1 {
		return nil, fmt.Errorf("mode only can be 0 or 1")
	}

	fileIO := &FileIO{
		file: file,
		mode: mode,
	}

	if fileIO.mode == 0 {
		fileIO.reader = bufio.NewReader(file)
	} else {
		fileIO.writer = bufio.NewWriter(file)
	}

	return fileIO, nil
}

// writeULong
func (f *FileIO) WriteULong(v uint64) error {
	if f.mode != 1 {
		return fmt.Errorf("FileIO mode is reader, can not write")
	}
	var wbuf []byte = make([]byte, 0)
	n := binary.PutUvarint(wbuf, v)
	_, err := f.writer.Write(wbuf[:n])
	return err
}

func (f *FileIO) WriteInt(v int) error {
	var intBuf [4]byte
	if f.mode != 1 {
		return fmt.Errorf("FileIO mode is reader, can not write")
	}
	wbuf := intBuf[:4]
	wbuf[0] = byte(v & 0xff)
	wbuf[1] = byte((v >> 8) & 0xff)
	wbuf[2] = byte((v >> 16) & 0xff)
	wbuf[3] = byte((v >> 24) & 0xff)
	_, err := f.writer.Write(wbuf)
	return err
}

func (f *FileIO) WriteChar(v byte) error {
	if f.mode != 1 {
		return fmt.Errorf("FileIO mode is reader, can not write")
	}

	return f.writer.WriteByte(v)
}

func (f *FileIO) WriteDouble(v float64) error {
	var doubleBuf [8]byte
	if f.mode != 1 {
		return fmt.Errorf("FileIO mode is reader, can not write")
	}
	wBuf := doubleBuf[:8]
	PutFloat64LE(wBuf, v)
	_, err := f.writer.Write(wBuf)
	return err
}

func (f *FileIO) WriteLine(l string) error {
	if f.mode != 1 {
		return fmt.Errorf("FileIO mode is reader, can not write")
	}
	_, err := f.writer.WriteString(fmt.Sprintf("%s\n", l))
	return err
}

func (f *FileIO) ReadULong() (uint64, error) {
	if f.mode != 0 {
		return 0, fmt.Errorf("FileIO mode is write, can not read")
	}

	return binary.ReadUvarint(f.reader)
}

func (f *FileIO) ReadInt() (int, error) {
	if f.mode != 0 {
		return 0, fmt.Errorf("FileIO mode is write, can not read")
	}

	var v uint32
	binary.Read(f.reader, binary.LittleEndian, &v)
	return int(v), nil
}

func (f *FileIO) ReadChar() (byte, error) {
	return f.reader.ReadByte()
}

func (f *FileIO) ReadDouble() (float64, error) {
	var doubleBuf [8]byte
	if f.mode != 0 {
		return .0, fmt.Errorf("FileIO mode is write, can not read")
	}

	wbuf := doubleBuf[:8]
	_, err := io.ReadFull(f.reader, wbuf)
	if err != nil {
		log.Error(err)
		return .0, err
	}

	return GetFloat64LE(wbuf), nil
}

func (f *FileIO) ReadFloat32() (float32, error) {
	var floatBuf [4]byte
	if f.mode != 0 {
		return .0, fmt.Errorf("FileIO mode is write, can not read")
	}

	wbuf := floatBuf[:4]
	_, err := io.ReadFull(f.reader, wbuf)
	if err != nil {
		log.Error(err)
		return .0, err
	}

	return GetFloat32LE(wbuf), nil
}

func (f *FileIO) IsOpen() bool {
	return f.file != nil && (f.reader != nil || f.writer != nil)
}

func (f *FileIO) Rewind() error {
	_, err := f.file.Seek(0, 0)
	return err
}

func (f *FileIO) ReadLine() (string, error) {
	if f.mode != 0 {
		return "", fmt.Errorf("FileIO mode is write, can not read")
	}

	line, _, err := f.reader.ReadLine()
	if err != nil {
		return "", err
	}

	f.Line = string(line)
	return string(line), nil
}

func (f *FileIO) Close() error {
	return f.file.Close()
}
