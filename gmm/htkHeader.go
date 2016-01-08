package gmm

import (
	"fmt"
	"os"
)

type htkHeader struct {
	nSamples   int64 //frames in the file
	sampPeriod int64 //sample period in 100ns
	sampSize   int16 //vector size
	parmKind   int16 //parameter type

}

func readHTKHeader(file *os.File ) (*htkHeader, error) {

	var buf []byte = make([]byte, 8, 8)
	n, err := file.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != 8 {
		return nil, fmt.Errorf("read buf wrong 1")
	}

	var nSamples int64
	nSamples = int64(buf[0])
	nSamples |= int64(buf[1]) << 8
	nSamples |= int64(buf[2]) << 16
	nSamples |= int64(buf[3]) << 24
	nSamples |= int64(buf[4]) << 32
	nSamples |= int64(buf[5]) << 40
	nSamples |= int64(buf[6]) << 48
	nSamples |= int64(buf[7]) << 56

	//	buf = make([]byte, 8, 8)
	n, err = file.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != 8 {
		return nil, fmt.Errorf("read buf wrong 2")
	}

	var sampPeriod int64
	sampPeriod = int64(buf[0])
	sampPeriod |= int64(buf[1]) << 8
	sampPeriod |= int64(buf[2]) << 16
	sampPeriod |= int64(buf[3]) << 24
	sampPeriod |= int64(buf[4]) << 32
	sampPeriod |= int64(buf[5]) << 40
	sampPeriod |= int64(buf[6]) << 48
	sampPeriod |= int64(buf[7]) << 56

	var buf16 []byte = make([]byte, 4, 4)
	n, err = file.Read(buf16)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, fmt.Errorf("read buf16 wrong 1")
	}

	var sampSize int16
	sampSize = int16(buf[0])
	sampSize |= int16(buf[1]) << 8

	//	buf16 = make([]byte,4,4)
	n, err = file.Read(buf16)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, fmt.Errorf("read buf16 wrong 2")
	}

	var parmKind int16
	parmKind = int16(buf[0])
	parmKind |= int16(buf[1]) << 8
	return &htkHeader{
		nSamples:   nSamples,
		sampPeriod: sampPeriod,
		sampSize:   sampSize,
		parmKind:   parmKind,
	}, nil
}
