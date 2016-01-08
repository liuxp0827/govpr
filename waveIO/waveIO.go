package waveIO

import (
	"bufio"
	"fmt"
	"govpr/constant"
	"io"
	"math"
	"os"
)

type WaveChunk struct {
	riff   []byte // RIFF file identification (4 bytes)
	length uint32 // length field (4 bytes)
	wave   []byte // WAVE chunk identification (4 bytes)
}

func (w WaveChunk) Bytes() []byte {
	var buf []byte = make([]byte, 0)
	buf = append(buf, w.riff...)
	buf = append(buf, byte(w.length&0xff))
	buf = append(buf, byte((w.length>>8)&0xff))
	buf = append(buf, byte((w.length>>16)&0xff))
	buf = append(buf, byte((w.length>>24)&0xff))
	buf = append(buf, w.wave...)
	return buf
}

type FmtChunk struct {
	fmt       []byte // format sub-chunk identification  (4 bytes)
	flength   uint32 // length of format sub-chunk (4 byte integer)
	format    int16  // format specifier (2 byte integer)
	chans     int16  // number of channels (2 byte integer)
	sampsRate uint32 // sample rate in Hz (4 byte integer)
	bpsec     uint32 // bytes per second (4 byte integer)
	bpsample  int16  // bytes per sample (2 byte integer)
	bpchan    int16  // bits per channel (2 byte integer)
}

func (f FmtChunk) Bytes() []byte {
	var buf []byte = make([]byte, 0)
	// fmt
	buf = append(buf, f.fmt...)
	// flength
	buf = append(buf, byte(f.flength&0xff))
	buf = append(buf, byte((f.flength>>8)&0xff))
	buf = append(buf, byte((f.flength>>16)&0xff))
	buf = append(buf, byte((f.flength>>24)&0xff))
	// format
	buf = append(buf, byte(f.format&0xff))
	buf = append(buf, byte((f.format>>8)&0xff))
	// chans
	buf = append(buf, byte(f.chans&0xff))
	buf = append(buf, byte((f.chans>>8)&0xff))
	// sampsRate
	buf = append(buf, byte(f.sampsRate&0xff))
	buf = append(buf, byte((f.sampsRate>>8)&0xff))
	buf = append(buf, byte((f.sampsRate>>16)&0xff))
	buf = append(buf, byte((f.sampsRate>>24)&0xff))
	// bpsec
	buf = append(buf, byte(f.bpsec&0xff))
	buf = append(buf, byte((f.bpsec>>8)&0xff))
	buf = append(buf, byte((f.bpsec>>16)&0xff))
	buf = append(buf, byte((f.bpsec>>24)&0xff))
	// bpsample
	buf = append(buf, byte(f.bpsample&0xff))
	buf = append(buf, byte((f.bpsample>>8)&0xff))
	// bpchan
	buf = append(buf, byte(f.bpchan&0xff))
	buf = append(buf, byte((f.bpchan>>8)&0xff))
	return buf
}

type DataChunk struct {
	data    []byte // data sub-chunk identification  (4 bytes)
	dlength uint32 // length of data sub-chunk (4 byte integer)
}

func (d DataChunk) Bytes() []byte {
	var buf []byte = make([]byte, 0)
	// data
	buf = append(buf, d.data...)
	// dlength
	buf = append(buf, byte(d.dlength&0xff))
	buf = append(buf, byte((d.dlength>>8)&0xff))
	buf = append(buf, byte((d.dlength>>16)&0xff))
	buf = append(buf, byte((d.dlength>>24)&0xff))

	return buf
}

type WavInfo struct {
	Length      int64 // number of samples in the data chunk
	SampleRate  int64 // sample rate
	BitSPSample int64 // bits per sample
}

type WaveIO struct {
	waveChunk WaveChunk // 12 Bytes
	fmtChunk  FmtChunk  // 24 Bytes
	dataChunk DataChunk // 8 Bytes
}

func (w *WaveIO) Bytes() []byte {
	var buf []byte = make([]byte, 0)
	// waveChunk
	buf = append(buf, w.waveChunk.Bytes()...)
	// fmtChunk
	buf = append(buf, w.fmtChunk.Bytes()...)
	// dataChunk
	buf = append(buf, w.dataChunk.Bytes()...)

	return buf
}

func WaveSave(detFile string, wavData []int16, length uint32) error {
	waveIO := new(WaveIO)

	wFile, err := os.Create(detFile)
	if err != nil {
		return err
	}

	// waveChunk
	waveIO.waveChunk.riff = []byte("RIFF")
	waveIO.waveChunk.wave = []byte("WAVE")
	waveIO.waveChunk.length = length*2 + 36

	// fmtChunk
	waveIO.fmtChunk.bpchan = 16
	waveIO.fmtChunk.bpsample = 2
	waveIO.fmtChunk.bpsec = 32000
	waveIO.fmtChunk.chans = 1
	waveIO.fmtChunk.flength = 16
	waveIO.fmtChunk.fmt = []byte{'f', 'm', 't', ' '}
	waveIO.fmtChunk.format = 1
	waveIO.fmtChunk.sampsRate = 16000

	// dataChunk
	waveIO.dataChunk.data = []byte("data")
	waveIO.dataChunk.dlength = length * 2

	w := bufio.NewWriter(wFile)
	_, err = w.Write(waveIO.Bytes())
	if err != nil {
		return err
	}

	lenOfWav16 := len(wavData)
	data := make([]byte, 0)
	for i := 0; i < lenOfWav16; i++ {
		data = append(data, byte(wavData[i]&0xff))
		data = append(data, byte((wavData[i]>>8)&0xff))
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}

	w.Flush()
	return nil
}

func WaveLoad(srcFile string, wavData *[]float32, wavinfo *WavInfo) error {
	var ii, iTotalReaded int = 0, 0
	var i, bpsample, sampleRate int64
	var lengthOfData uint32

	var pf *[]float32

	cBuff := make([]byte, 0x4000)
	rFile, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer rFile.Close()
	r := bufio.NewReader(rFile)
	bufsum := make([]byte, 0)
	for {
		i++
		buf, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		bufsum = append(bufsum, buf)

		if i == 4 {
			if bufsum[0] != 'R' || bufsum[1] != 'I' || bufsum[2] != 'F' || bufsum[3] != 'F' {
				return fmt.Errorf("invalid wave haeder")
			}
		}
		if i == 24 {
			if bufsum[22] != 1 || bufsum[23] != 0 {
				return fmt.Errorf("this wave channel is not 1")
			}
		}

		if i == 28 {
			sampleRate = int64(bufsum[24])
			sampleRate |= int64(bufsum[25]) << 8
			sampleRate |= int64(bufsum[26]) << 16
			sampleRate |= int64(bufsum[27]) << 24
			wavinfo.SampleRate = sampleRate
		}

		if i == 34 {
			bpsample = int64(bufsum[32])
			bpsample |= int64(bufsum[33]) << 8
			if bpsample != 1 && bpsample != 2 {
				return fmt.Errorf("Can only process 8 or 16 quantization bit")
			}
			wavinfo.BitSPSample = bpsample * 8
		}

		if i == 44 {
			lengthOfData = uint32(bufsum[40])
			lengthOfData |= uint32(bufsum[41]) << 8
			lengthOfData |= uint32(bufsum[42]) << 16
			lengthOfData |= uint32(bufsum[43]) << 24
			if lengthOfData > 0 {
				wavinfo.Length = int64(lengthOfData) / bpsample
			} else {
				return fmt.Errorf("length of wave data is 0")
			}
			if wavData == nil {
				return fmt.Errorf("Unallocated memory address for wave data")
			}
			pf = wavData
			break
		}

	}
	for {
		iBytesReaded, err := r.Read(cBuff)
		if err != nil || iTotalReaded >= int(lengthOfData) {
			break
		}

		iTotalReaded += iBytesReaded

		if iTotalReaded >= int(lengthOfData) {
			iBytesReaded = iBytesReaded - (iTotalReaded - int(lengthOfData))
		}

		for ii = 0; ii < iBytesReaded; ii += 2 { //byte--->short
			cBuff16 := int16(cBuff[ii])
			cBuff16 |= int16(cBuff[ii+1]) << 8
			*pf = append(*pf, float32(cBuff16))
		}
	}

	return nil
}

func WaveLoad2(srcFile string, wavData *[]int16, length *uint32) error {
	var ii, iTotalReaded int = 0, 0
	var i int = 0
	var lengthOfData uint32
	var pf *[]int16
	cBuff := make([]byte, 0x4000)
	rFile, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer rFile.Close()
	r := bufio.NewReader(rFile)
	bufsum := make([]byte, 0)
	for {
		i++
		buf, err := r.ReadByte()
		if err != nil {
			break
		}
		bufsum = append(bufsum, buf)
		if i == 4 {
			if bufsum[0] != 'R' || bufsum[1] != 'I' || bufsum[2] != 'F' || bufsum[3] != 'F' {
				return fmt.Errorf("invalid wave haeder")
			}
		}
		if i == 24 {

			if bufsum[22] != 1 || bufsum[23] != 0 {
				return fmt.Errorf("this wave channel is not 1")
			}
		}
		if i == 44 {
			lengthOfData = uint32(bufsum[40])
			lengthOfData |= uint32(bufsum[41]) << 8
			lengthOfData |= uint32(bufsum[42]) << 16
			lengthOfData |= uint32(bufsum[43]) << 24
			if lengthOfData > 0 {
				*length = lengthOfData / 2
			} else {
				return fmt.Errorf("length of wave data is 0")
			}
			if wavData == nil {
				return fmt.Errorf("Unallocated memory address for wave data")
			}
			pf = wavData
			break
		}

	}
	for {
		iBytesReaded, err := r.Read(cBuff)
		if err != nil || iTotalReaded >= int(lengthOfData) {
			break
		}

		iTotalReaded += iBytesReaded

		if iTotalReaded >= int(lengthOfData) {
			iBytesReaded = iBytesReaded - (iTotalReaded - int(lengthOfData))
		}

		for ii = 0; ii < iBytesReaded; ii += 2 { //byte--->short
			cBuff16 := int16(cBuff[ii])
			cBuff16 |= int16(cBuff[ii+1]) << 8
			*pf = append(*pf, cBuff16)
		}
	}

	return nil
}

func DelSilence(pnSrc []int16, K int) []int16 {
	var max_sample_value int = -(constant.SHRT_MAX)
	var nSrcLen, outLength int64 = int64(len(pnSrc)), 0

	for i := int64(0); i < nSrcLen; i++ {
		if int(math.Abs(float64(pnSrc[i]))) > max_sample_value {
			max_sample_value = int(math.Abs(float64(pnSrc[i])))
		}
	}

	MIN_VOC_ENG := constant.MIN_VOC_ENG
	if max_sample_value < MIN_VOC_ENG {
		MIN_VOC_ENG = max_sample_value / 2
	}

	if K > 50 {
		MIN_VOC_ENG = MIN_VOC_ENG + (max_sample_value-MIN_VOC_ENG)*(K-50)/50
	} else if K < 50 {
		MIN_VOC_ENG = MIN_VOC_ENG - (MIN_VOC_ENG)*(50-K)/50
	}

	var j, p int = 0, 0
	var old1, old2, old3, curSample int16
	var pnTarget []int16
	var pCur *[]int16
	var pWinBuf [constant.VOC_BLOCK_LEN + 1]int16
	var nWin, nMod, i, k, eng int

	pnTarget = make([]int16, 0)
	pCur = &pnTarget

	nWin = int(nSrcLen) / constant.VOC_BLOCK_LEN
	nMod = int(nSrcLen) % constant.VOC_BLOCK_LEN

	for i = 0; i < nWin; i++ {
		eng = 0
		for k = 0; k < constant.VOC_BLOCK_LEN; k++ {
			eng += int(math.Abs(float64(pnSrc[constant.VOC_BLOCK_LEN*i+k])))
		}

		if eng > MIN_VOC_ENG*constant.VOC_BLOCK_LEN {
			j, p = 0, 0
			old1, old2, old3 = 0, 0, 0
			for k = 0; k < constant.VOC_BLOCK_LEN; k++ {
				curSample = pnSrc[constant.VOC_BLOCK_LEN*i+k]
				if curSample == old1 && old1 == old2 && old2 == old3 {
					if p >= 0 {
						j = p
					}
				} else {
					pWinBuf[j] = curSample
					j++
					p = j - 3
				}
				old3 = old2
				old2 = old1
				old1 = curSample
			}
			for _, v := range pWinBuf[:j] {
				*pCur = append(*pCur, v)
			}
			outLength += int64(j)
		}
	}
	////////////////////////////////////////////////////////////////////////////
	eng = 0
	for i = 0; i < nMod; i++ {
		eng += int(math.Abs(float64(pnSrc[constant.VOC_BLOCK_LEN*nWin+i])))
	}

	if eng > MIN_VOC_ENG*nMod {
		j, p = 0, 0
		old1, old2, old3 = 0, 0, 0
		for i = 0; i < nMod; i++ {
			curSample = pnSrc[constant.VOC_BLOCK_LEN*nWin+i]
			if curSample == old1 && old1 == old2 && old2 == old3 {
				if p >= 0 {
					j = p
				}
			} else {
				pWinBuf[j] = curSample
				j++
				p = j - 3
			}
			old3 = old2
			old2 = old1
			old1 = curSample
		}
		for _, v := range pWinBuf[:j] {
			*pCur = append(*pCur, v)
		}
		outLength += int64(j)
	}

	return pnTarget
}
