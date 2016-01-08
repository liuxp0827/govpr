package govpr

import (
	"fmt"
	"govpr/constant"
	"govpr/feature"
	"govpr/gmm"
	"govpr/waveIO"
	"math"
)

type VPREngine struct {
	vprType    int // 0: text-dependent; 1: text-independent
	engineType int // type of engine
	level      int // accuracy level
	sexType    int // sex type, default for male
	sampleRate int // 采样率

	verifyBuf []int16
	trainBuf  []int16

	score float64

	workDir     string
	modelPath   string
	modelName   string
	delSilRange int

	_minTrainLen    int64
	_minVerLen      int64
	_minPerTrainLen int64
}

func NewVPREngine(vprType, sampleRate, delSilRange int, workDir, modelPath, modelName string) *VPREngine {
	return &VPREngine{
		vprType:         vprType,
		workDir:         workDir,
		modelPath:       modelPath,
		modelName:       modelName,
		sampleRate:      sampleRate,
		verifyBuf:       make([]int16, 0),
		trainBuf:        make([]int16, 0),
		delSilRange:     delSilRange,
		_minTrainLen:    int64(sampleRate * 2),
		_minVerLen:      int64(float64(sampleRate) * 0.25),
		_minPerTrainLen: int64(float64(sampleRate) * 0.3),
	}
}

/**
 * set engine type (verification, trainer or adaptation)
 *
 * @param vtype
 * 		  [in] 0: verification, 1: training, 2: adaptation, 3: identification, 4: off-line verification
 *
 * @return
 */
func (this *VPREngine) SetEngineType(vtype int) error {
	if vtype < 0 || vtype > 4 {
		return fmt.Errorf("invaild EngineType")
	}
	this.engineType = vtype
	return nil
}

func (this *VPREngine) GetEngineType() int {
	return this.engineType
}

/**
 * set type (0: text-dependent; 1: text-independent)
 *
 * @param vtype
 * 		  [in] 0: text-dependent; 1: text-independent
 *
 * @return
 */
func (this *VPREngine) SetVPRType(vtype int) error {
	if vtype < 0 || vtype > 2 {
		return fmt.Errorf("invaild vprType")
	}
	this.vprType = vtype
	return nil
}

func (this *VPREngine) GetVPRType() int {
	return this.vprType
}

func (this *VPREngine) TrainModel() error {
	if this.trainBuf == nil || int64(len(this.trainBuf)) < this._minTrainLen {
		return LSV_ERR_NO_AVAILABLE_DATA
	}

	var ubm *gmm.GMM = gmm.NewGMM()
	var client *gmm.GMM = gmm.NewGMM()

	if err := ubm.LoadModel(this.workDir + "ubm"); err != nil {
		return NewError(LSV_ERR_MODEL_LOAD_FAILED, err.Error())
	}

	client.DupModel(ubm)
	if _, err := feature.FeatureExtract(this.trainBuf, ubm); err != nil {
		return NewError(LSV_ERR_MEM_INSUFFICIENT, err.Error())
	}

	for k := 0; k < constant.MAXLOP; k++ {
		if ret, err := ubm.EM(ubm.INumMixtures); ret == 0 || err != nil {
			return NewError(LSV_ERR_TRAINING_FAILED, err.Error())
		}

		for i := 0; i < ubm.INumMixtures; i++ {
			for j := 0; j < ubm.IVectorSize; j++ {
				client.DMean[i][j] = (float64(ubm.IFrames)*ubm.DMixtureWeight[i])*
					ubm.DMean[i][j] + constant.REL_FACTOR*client.DMean[i][j]

				client.DMean[i][j] /= (float64(ubm.IFrames)*ubm.DMixtureWeight[i] + constant.REL_FACTOR)
			}
		}
	}

	if err := client.SaveModel(this.modelPath + this.modelName); err != nil {
		return NewError(LSV_ERR_TRAINING_FAILED, err.Error())

	}
	return nil
}

func (this *VPREngine) VerifyModel() error {
	if this.verifyBuf == nil || int64(len(this.verifyBuf)) <= 0 {
		return LSV_ERR_NO_AVAILABLE_DATA
	}

	var buf []int16
	var length int64

	buf = waveIO.DelSilence(this.verifyBuf, this.delSilRange)
	if length < this._minVerLen {
		return LSV_ERR_NEED_MORE_SAMPLE
	}

	length = int64(len(buf))

	var world *gmm.GMM = gmm.NewGMM()
	var client *gmm.GMM = gmm.NewGMM()

	err := world.LoadModel(this.modelPath + this.modelName + ".dat")
	if err != nil {
		return NewError(LSV_ERR_MODEL_LOAD_FAILED, err.Error())
	}

	_, err = feature.FeatureExtract(buf, client)
	if err != nil {
		return NewError(LSV_ERR_MEM_INSUFFICIENT, err.Error())
	}

	err = world.CopyFParam(client)
	if err != nil {
		return NewError(LSV_ERR_MEM_INSUFFICIENT, err.Error())
	}

	var logClient, logWorld float64
	logClient = client.LProb(client.FParam, 0, int64(client.IFrames))
	logWorld = world.LProb(world.FParam, 0, int64(world.IFrames))
	this.score = (logClient - logWorld) / float64(client.IFrames)
	return nil
}

func (this *VPREngine) AddTrainBuffer(buf []int16) error {
	if buf == nil || len(buf) == 0 {
		return LSV_ERR_NO_AVAILABLE_DATA
	}

	this.trainBuf = append(this.trainBuf, buf...)
	return nil
}

func (this *VPREngine) AddVerifyBuffer(buf []int16) error {
	if buf == nil || len(buf) == 0 {
		return LSV_ERR_NO_AVAILABLE_DATA
	}

	this.verifyBuf = buf
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

func getValidVoiceLen(pnSrc []int16) uint32 {
	var nSrcLen, outLength uint32 = uint32(len(pnSrc)), 0
	var pWinBuf [constant.VOC_BLOCK_LEN + 1]int16
	var nWin, nMod, i, k, eng int
	var j, p int = 0, 0
	var old1, old2, old3, curSample int16

	nWin = int(nSrcLen) / constant.VOC_BLOCK_LEN
	nMod = int(nSrcLen) % constant.VOC_BLOCK_LEN

	for i = 0; i < nWin; i++ {
		eng = 0
		for k = 0; k < constant.VOC_BLOCK_LEN; k++ {
			eng += int(math.Abs(float64(pnSrc[constant.VOC_BLOCK_LEN*i+k])))
		}

		if eng > constant.MIN_VOC_ENG*constant.VOC_BLOCK_LEN {
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
			outLength += uint32(j)
		}
	}

	eng = 0
	for i = 0; i < nMod; i++ {
		eng += int(math.Abs(float64(pnSrc[constant.VOC_BLOCK_LEN*nWin+i])))
	}

	if eng > constant.MIN_VOC_ENG*nMod {
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

		outLength += uint32(j)
	}
	return outLength
}
