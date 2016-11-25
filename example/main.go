package main

import (
	"github.com/liuxp0827/govpr"
	"github.com/liuxp0827/govpr/log"
	"github.com/liuxp0827/govpr/waveIO"
	"io/ioutil"
)

type engine struct {
	vprEngine *govpr.VPREngine
}

func NewEngine(sampleRate, delSilRange int, ubmFile, userModelFile string) (*engine, error) {
	vprEngine, err := govpr.NewVPREngine(sampleRate, delSilRange, true, ubmFile, userModelFile)
	if err != nil {
		return nil, err
	}
	return &engine{vprEngine: vprEngine}, nil
}

func (this *engine) DestroyEngine() {
	this.vprEngine = nil
}

func (this *engine) TrainSpeech(buffers [][]byte) error {

	var err error
	count := len(buffers)
	for i := 0; i < count; i++ {
		err = this.vprEngine.AddTrainBuffer(buffers[i])
		if err != nil {
			log.Error(err)
			return err
		}
	}

	defer this.vprEngine.ClearTrainBuffer()
	defer this.vprEngine.ClearAllBuffer()

	err = this.vprEngine.TrainModel()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (this *engine) RecSpeech(buffer []byte) (float64, error) {

	err := this.vprEngine.AddVerifyBuffer(buffer)
	defer this.vprEngine.ClearVerifyBuffer()
	if err != nil {
		log.Error(err)
		return -1.0, err
	}

	err = this.vprEngine.VerifyModel()
	if err != nil {
		log.Error(err)
		return -1.0, err
	}

	return this.vprEngine.GetScore(), nil
}

func main() {
	log.SetLevel(log.LevelDebug)

	vprEngine, err := NewEngine(16000, 50, "../ubm/ubm", "model/test.dat")
	if err != nil {
		log.Fatal(err)
	}

	trainlist := []string{
		"wav/train/01_32468975.wav",
		"wav/train/02_58769423.wav",
		"wav/train/03_59682734.wav",
		"wav/train/04_64958273.wav",
		"wav/train/05_65432978.wav",
	}

	trainBuffer := make([][]byte, 0)

	for _, file := range trainlist {
		buf, err := loadWaveData(file)
		if err != nil {
			log.Error(err)
			return
		}
		trainBuffer = append(trainBuffer, buf)
	}

	err = vprEngine.TrainSpeech(trainBuffer)
	if err != nil {
		log.Fatal(err)
	}

	var threshold float64 = 1.0

	selfverifyBuffer, err := waveIO.WaveLoad("wav/verify/self_34986527.wav")
	if err != nil {
		log.Fatal(err)
	}

	self_score, err := vprEngine.RecSpeech(selfverifyBuffer)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("self score %f, pass? %v", self_score, self_score >= threshold)

	otherverifyBuffer, err := waveIO.WaveLoad("wav/verify/other_38974652.wav")
	if err != nil {
		log.Fatal(err)
	}

	other_score, err := vprEngine.RecSpeech(otherverifyBuffer)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("other score %f, pass? %v", other_score, other_score >= threshold)
}

func loadWaveData(file string) ([]byte, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	// remove .wav header info 44 bits
	data = data[44:]
	return data, nil
}
