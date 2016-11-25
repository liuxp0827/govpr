package govpr

import (
	"fmt"
	"github.com/liuxp0827/govpr/constant"
	"github.com/liuxp0827/govpr/feature"
	"github.com/liuxp0827/govpr/gmm"
	"github.com/liuxp0827/govpr/log"
	"github.com/liuxp0827/govpr/waveIO"
	"os"
	"path"
)

type VPREngine struct {
	trainBuf  []int16
	verifyBuf []int16

	score float64

	ubmFile       string
	userModelFile string

	deleteSil   bool
	delSilRange int

	ubm *gmm.GMM

	_minTrainLen int64
	_minVerLen   int64
}

func NewVPREngine(sampleRate, delSilRange int, deleteSil bool, ubmFile, userModelFile string) (*VPREngine, error) {
	engine := VPREngine{
		ubmFile:       ubmFile,
		userModelFile: userModelFile,
		verifyBuf:     make([]int16, 0),
		trainBuf:      make([]int16, 0),
		deleteSil:     deleteSil,
		delSilRange:   delSilRange,
		ubm:           gmm.NewGMM(),
		_minTrainLen:  int64(sampleRate * 2),
		_minVerLen:    int64(float64(sampleRate) * 0.25),
	}

	err := engine.init()
	if err != nil {
		return nil, err
	}

	return &engine, nil
}

func (this *VPREngine) init() error {
	if this.ubm == nil {
		return fmt.Errorf("ubm model is nil")
	}

	if err := this.ubm.LoadModel(this.ubmFile); err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MODEL_LOAD_FAILED, err.Error())
	}
	return nil
}

func (this *VPREngine) TrainModel() error {
	if this.trainBuf == nil || int64(len(this.trainBuf)) < this._minTrainLen {
		return LSV_ERR_NO_AVAILABLE_DATA
	}

	tmpubm := gmm.NewGMM()
	tmpubm.Copy(this.ubm)

	client := gmm.NewGMM()
	client.DupModel(this.ubm)
	if err := feature.Extract(this.trainBuf, tmpubm); err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MEM_INSUFFICIENT, err.Error())
	}

	for k := 0; k < constant.MAXLOP; k++ {
		if ret, err := tmpubm.EM(tmpubm.Mixtures); ret == 0 || err != nil {
			log.Error(err)
			return NewError(LSV_ERR_TRAINING_FAILED, err.Error())
		}

		for i := 0; i < tmpubm.Mixtures; i++ {
			for j := 0; j < tmpubm.VectorSize; j++ {
				client.Mean[i][j] = (float64(tmpubm.Frames)*tmpubm.MixtureWeight[i])*
					tmpubm.Mean[i][j] + constant.REL_FACTOR*client.Mean[i][j]

				client.Mean[i][j] /= (float64(tmpubm.Frames)*tmpubm.MixtureWeight[i] + constant.REL_FACTOR)
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

	var client *gmm.GMM = gmm.NewGMM()
	err := client.LoadModel(this.userModelFile)
	if err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MODEL_LOAD_FAILED, err.Error())
	}

	tmpubm := gmm.NewGMM()
	tmpubm.Copy(this.ubm)

	err = feature.Extract(buf, client)
	if err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MEM_INSUFFICIENT, err.Error())
	}

	err = tmpubm.CopyFeatureData(client)
	if err != nil {
		log.Error(err)
		return NewError(LSV_ERR_MEM_INSUFFICIENT, err.Error())
	}

	var logClient, logWorld float64
	logClient = client.LProb(client.FeatureData, 0, int64(client.Frames))
	logWorld = tmpubm.LProb(tmpubm.FeatureData, 0, int64(tmpubm.Frames))
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

	if this.deleteSil {
		sBuff = waveIO.DelSilence(sBuff, this.delSilRange)
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

	if this.deleteSil {
		sBuff = waveIO.DelSilence(sBuff, this.delSilRange)
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
