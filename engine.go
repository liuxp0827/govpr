package govpr

import (
	"github.com/liuxp0827/govpr/constant"
	"github.com/liuxp0827/govpr/feature"
	"github.com/liuxp0827/govpr/gmm"
	"github.com/liuxp0827/govpr/log"
	"math"
	"os"
	"path"
)

type VPREngine struct {
	trainBuf  []int16
	verifyBuf []int16

	score float64

	ubmFile       string
	userModelFile string
	delSilRange   int

	_minTrainLen int64
	_minVerLen   int64
}

func NewVPREngine(sampleRate, delSilRange int, ubmFile, userModelFile string) *VPREngine {
	return &VPREngine{
		ubmFile:       ubmFile,
		userModelFile: userModelFile,
		verifyBuf:     make([]int16, 0),
		trainBuf:      make([]int16, 0),
		delSilRange:   delSilRange,
		_minTrainLen:  int64(sampleRate * 2),
		_minVerLen:    int64(float64(sampleRate) * 0.25),
	}
}

func (this *VPREngine) TrainModel() error {
	if this.trainBuf == nil || int64(len(this.trainBuf)) < this._minTrainLen {
		return LSV_ERR_NO_AVAILABLE_DATA
	}

	var ubm *gmm.GMM = gmm.NewGMM()
	var client *gmm.GMM = gmm.NewGMM()

	if err := ubm.LoadModel(this.ubmFile); err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MODEL_LOAD_FAILED, err.Error())
	}

	client.DupModel(ubm)
	if _, err := feature.Extract(this.trainBuf, ubm); err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MEM_INSUFFICIENT, err.Error())
	}

	for k := 0; k < constant.MAXLOP; k++ {
		if ret, err := ubm.EM(ubm.Mixtures); ret == 0 || err != nil {
			log.Error(err)
			return NewError(LSV_ERR_TRAINING_FAILED, err.Error())
		}

		for i := 0; i < ubm.Mixtures; i++ {
			for j := 0; j < ubm.VectorSize; j++ {
				client.Mean[i][j] = (float64(ubm.Frames)*ubm.MixtureWeight[i])*
					ubm.Mean[i][j] + constant.REL_FACTOR*client.Mean[i][j]

				client.Mean[i][j] /= (float64(ubm.Frames)*ubm.MixtureWeight[i] + constant.REL_FACTOR)
			}
		}
	}

	userModelPath := path.Dir(this.userModelFile)
	err := os.MkdirAll(userModelPath, 0755)
	if err != nil {
		log.Error(err)
		return NewError(LSV_ERR_TRAINING_FAILED, err.Error())
	}

	if err = client.SaveModel(this.userModelFile); err != nil {
		log.Error(err)
		return NewError(LSV_ERR_TRAINING_FAILED, err.Error())
	}
	return nil
}

func (this *VPREngine) VerifyModel() error {
	if this.verifyBuf == nil || len(this.verifyBuf) <= 0 {
		return LSV_ERR_NO_AVAILABLE_DATA
	}

	var buf []int16 = this.verifyBuf
	var length int64

	//buf = waveIO.DelSilence(this.verifyBuf, this.delSilRange)

	length = int64(len(buf))
	if length < this._minVerLen {
		return LSV_ERR_NEED_MORE_SAMPLE
	}

	var world *gmm.GMM = gmm.NewGMM()
	var client *gmm.GMM = gmm.NewGMM()

	err := world.LoadModel(this.ubmFile)
	if err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MODEL_LOAD_FAILED, err.Error())
	}

	err = client.LoadModel(this.userModelFile)
	if err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MODEL_LOAD_FAILED, err.Error())
	}

	_, err = feature.Extract(buf, client)
	if err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MEM_INSUFFICIENT, err.Error())
	}

	err = world.CopyFeatureData(client)
	if err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MEM_INSUFFICIENT, err.Error())
	}

	var logClient, logWorld float64
	logClient = client.LProb(client.FeatureData, 0, int64(client.Frames))
	logWorld = world.LProb(world.FeatureData, 0, int64(world.Frames))
	this.score = (logClient - logWorld) / float64(client.Frames)
	return nil
}

func (this *VPREngine) AddTrainBuffer(buf []byte) error {
	if buf == nil || len(buf) == 0 {
		return LSV_ERR_NO_AVAILABLE_DATA
	}
	sBuff := make([]int16, 0)
	length := len(buf)
	for ii := 0; ii < length; ii += 2 {
		cBuff16 := int16(buf[ii])
		cBuff16 |= int16(buf[ii+1]) << 8
		sBuff = append(sBuff, cBuff16)
	}

	this.trainBuf = append(this.trainBuf, sBuff...)
	return nil
}

func (this *VPREngine) AddVerifyBuffer(buf []byte) error {
	if buf == nil || len(buf) == 0 {
		return LSV_ERR_NO_AVAILABLE_DATA
	}

	sBuff := make([]int16, 0)
	length := len(buf)
	for ii := 0; ii < length; ii += 2 {
		cBuff16 := int16(buf[ii])
		cBuff16 |= int16(buf[ii+1]) << 8
		sBuff = append(sBuff, cBuff16)
	}

	this.verifyBuf = sBuff
	return nil
}

func (this *VPREngine) ClearTrainBuffer() {
	this.trainBuf = this.trainBuf[:0]
}

func (this *VPREngine) ClearVerifyBuffer() {
	this.verifyBuf = this.verifyBuf[:0]
}

func (this *VPREngine) ClearAllBuffer() {
	this.ClearTrainBuffer()
	this.ClearVerifyBuffer()
}

func (this *VPREngine) GetScore() float64 {
	return this.score
}
