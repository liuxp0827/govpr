package fileIO

import (
	"bufio"
	"encoding/binary"
	"fmt"
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
	if f.mode != 1 {
		return fmt.Errorf("FileIO mode is reader, can not write")
	}
	wbuf := make([]byte, 4, 4)
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
	if f.mode != 1 {
		return fmt.Errorf("FileIO mode is reader, can not write")
	}
	wBuf := make([]byte, 8, 8)
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
	wbuf := make([]byte, 4, 4)
	n, err := f.reader.Read(wbuf)
	if err != nil {
		return 0, err
	}
	if n != 4 {
		return 0, fmt.Errorf("FileIO read Int wrong")
	}

	var v int
	v = int(wbuf[0])
	v |= int(wbuf[1]) << 8
	v |= int(wbuf[2]) << 16
	v |= int(wbuf[3]) << 24
	return v, nil
}

func (f *FileIO) ReadChar() (byte, error) {
	return f.reader.ReadByte()
}

func (f *FileIO) ReadDouble() (float64, error) {
	if f.mode != 0 {
		return .0, fmt.Errorf("FileIO mode is write, can not read")
	}
	wbuf := make([]byte, 8, 8)
	n, err := f.reader.Read(wbuf)
	if err != nil {
		return .0, err
	}
	if n != 8 {
		return .0, fmt.Errorf("FileIO read Double wrong")
	}

	return GetFloat64LE(wbuf), nil
}

func (f *FileIO) ReadFloat32() (float32, error) {
	if f.mode != 0 {
		return .0, fmt.Errorf("FileIO mode is write, can not read")
	}
	wbuf := make([]byte, 4, 4)
	n, err := f.reader.Read(wbuf)
	if err != nil {
		return .0, err
	}
	if n != 4 {
		return .0, fmt.Errorf("FileIO read float32 wrong")
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
