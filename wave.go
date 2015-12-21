package govpr

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type wavHeader []byte

func (this wavHeader) headerInit(size uint32) *wavHeader {
	chunkSize := size + 36
	this = make([]byte, 44)
	this[0] = 'R'
	this[1] = 'I'
	this[2] = 'F'
	this[3] = 'F'
	this[4] = byte(chunkSize & 0xff)
	this[5] = byte((chunkSize >> 8) & 0xff)
	this[6] = byte((chunkSize >> 16) & 0xff)
	this[7] = byte((chunkSize >> 24) & 0xff)
	this[8] = 'W'
	this[9] = 'A'
	this[10] = 'V'
	this[11] = 'E'
	this[12] = 'f'
	this[13] = 'm'
	this[14] = 't'
	this[15] = ' '
	this[16] = 16
	this[17] = 0
	this[18] = 0
	this[19] = 0
	this[20] = 1
	this[21] = 0
	this[22] = 1 //channel单通道
	this[23] = 0
	this[24] = 64
	this[25] = 31
	this[26] = 0
	this[27] = 0
	this[28] = 128
	this[29] = 62
	this[30] = 0
	this[31] = 0
	this[32] = 2
	this[33] = 0
	this[34] = 16
	this[35] = 0
	this[36] = 'd'
	this[37] = 'a'
	this[38] = 't'
	this[39] = 'a'
	this[40] = byte(size & 0xff)
	this[41] = byte((size >> 8) & 0xff)
	this[42] = byte((size >> 16) & 0xff)
	this[43] = byte((size >> 24) & 0xff)

	return &this

}

func CopyWavFile(det, src string) bool {
	bufsum := make([]byte, 0)
	rFile, err1 := os.Open(src)
	defer rFile.Close()
	if err1 != nil {
		log.Fatal(src, err1)
		return false

	}

	wFile, err2 := os.Create(det)
	defer wFile.Close()
	if err2 != nil {
		log.Fatal(det, err2)
		return false
	}

	r := bufio.NewReader(rFile)
	w := bufio.NewWriter(wFile)
	i := 0

	for {
		i++
		n, err := r.ReadByte()
		if err != nil {
			break
		}
		bufsum = append(bufsum, n)
	}
	w.Write(bufsum)
	w.Flush()
	return true
}

func waveSave(detFile string, wavData []int16, length uint32) bool {
	length = length * 2
	wFile, err := os.Create(detFile)
	if err != nil {
		fmt.Println(detFile, err)
		return false
	}
	h := new(wavHeader)
	head := h.headerInit(length)
	w := bufio.NewWriter(wFile)
	w.Write(*head)
	lenOfWav16 := len(wavData)
	data := make([]byte, 0)
	for i := 0; i < lenOfWav16; i++ {
		data = append(data, byte(wavData[i]&0xff))
		data = append(data, byte((wavData[i]>>8)&0xff))
	}

	w.Write(data)
	w.Flush()
	fmt.Printf("Save wav file %s successfully,the length of wav data is %d.\n", detFile, length/2)
	return true
}

func waveLoad(srcFile string, wavData *[]int16, length *uint32) bool {
	var ii, iTotalReaded int = 0, 0
	var i int = 0
	var lengthOfData uint32
	var pf *[]int16
	cBuff := make([]byte, 0x4000)
	rFile, err := os.Open(srcFile)
	if err != nil {
		fmt.Println(srcFile, err)
		return false
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
				fmt.Println("Error: That is not a wav file!")
				return false
			}
		}
		if i == 24 {

			if bufsum[22] != 1 || bufsum[23] != 0 {
				fmt.Println("Error: This wavfile's channel is not 1!")
				return false
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
				fmt.Println("Error: Length of Data is 0!")
				return false
			}
			if wavData == nil {
				fmt.Println("Error: Unallocated memory address for wavData!")
				return false
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
	fmt.Printf("Load wav file %s successfully,the length of wav data is %d.\n", srcFile, *length)
	return true
}