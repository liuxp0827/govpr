package feature

import (
	"fmt"
	"github.com/liuxp0827/govpr/constant"
	"github.com/liuxp0827/govpr/gmm"
	"github.com/liuxp0827/govpr/log"
	"github.com/liuxp0827/govpr/param"
	"github.com/liuxp0827/govpr/waveIO"
)

type parameter struct {
	lowCutOff              uint // low cut-off
	highCutOff             uint // high cut-off
	filterBankSize         int  // # num of filter-bank
	frameLength            int  // # frame length
	frameShift             int  // 10 # frame shift
	mfccOrder              int  // 16 # mfcc order
	isStatic               bool // t	# static mfcc
	isDynamic              bool // t	# dynamic mfcc
	isAcce                 bool // f	# acce mfcc
	cmsvn                  bool // t	# cmsvn
	isZeroGlobalMean       bool // t # zero global mean
	isDBNorm               bool // t # decibel normalization
	isDiffPolish           bool // f	# polish differential formula
	isDiffPowerSpectrum    bool // f	# differentail power spectrum
	isPredDiffAmplSpectrum bool // f	# predictive differential amplitude spectrum
	isEnergyNorm           bool
	silFloor               int16
	energyscale            int16
	isFeatWarping          bool
	featWarpWinSize        int16
	isRasta                bool
	rastaCoff              float64
}

func Extract(data []int16, gmm *gmm.GMM) (int, error) {
	var p, para []float32
	var info waveIO.WavInfo
	var cp *param.CParam = param.NewCParam()
	var pm *parameter = new(parameter)
	var err error
	var icol, irow int
	var buflen int = len(data)

	p = make([]float32, buflen, buflen)
	for i := 0; i < buflen; i++ {
		p[i] = float32(data[i])
	}

	pm.lowCutOff = constant.LOW_CUT_OFF
	pm.highCutOff = constant.HIGH_CUT_OFF
	pm.filterBankSize = constant.FILTER_BANK_SIZE
	pm.frameLength = constant.FRAME_LENGTH
	pm.frameShift = constant.FRAME_SHIFTt
	pm.mfccOrder = constant.MFCC_ORDER

	pm.isStatic = constant.BSTATIC
	pm.isDynamic = constant.BDYNAMIC
	pm.isAcce = constant.BACCE

	pm.cmsvn = constant.CMSVN
	pm.isZeroGlobalMean = constant.ZEROGLOBALMEAN
	pm.isDiffPolish = constant.DIFPOL
	pm.isDiffPowerSpectrum = constant.DPSCC
	pm.isPredDiffAmplSpectrum = constant.PDASCC
	pm.isEnergyNorm = constant.ENERGYNORM
	pm.silFloor = constant.SIL_FLOOR

	pm.energyscale = constant.ENERGY_SCALE
	pm.isFeatWarping = constant.FEATWARP
	pm.featWarpWinSize = constant.FEATURE_WARPING_WIN_SIZE
	pm.isDBNorm = constant.DBNORM
	pm.isRasta = constant.RASTA
	pm.rastaCoff = constant.RASTA_COFF

	info.SampleRate = constant.SAMPLERATE
	info.Length = int64(buflen)
	info.BitSPSample = constant.BIT_PER_SAMPLE

	if pm.highCutOff > pm.lowCutOff {
		err = cp.InitFBank2(int(info.SampleRate), pm.frameLength, pm.filterBankSize, int(pm.lowCutOff), int(pm.highCutOff))
	} else {
		err = cp.InitFBank(int(info.SampleRate), pm.frameLength, pm.filterBankSize)
	}

	if err != nil {
		return 0, err
	}

	err = cp.InitMfcc(pm.mfccOrder, float32(pm.frameShift))
	if err != nil {
		return 0, err
	}

	if pm.isStatic {
		cp.GetMfcc().IsStatic = true
	} else {
		cp.GetMfcc().IsStatic = false
	}

	if pm.isDynamic {
		cp.GetMfcc().IsDynamic = true
	} else {
		cp.GetMfcc().IsDynamic = false
	}

	if pm.isAcce {
		cp.GetMfcc().IsAcce = true
	} else {
		cp.GetMfcc().IsAcce = false
	}

	if pm.isZeroGlobalMean {
		cp.GetMfcc().IsZeroGlobalMean = true
	} else {
		cp.GetMfcc().IsZeroGlobalMean = false
	}

	if pm.isDBNorm {
		cp.GetMfcc().IsDBNorm = true
	} else {
		cp.GetMfcc().IsDBNorm = false
	}

	if pm.isDiffPolish {
		cp.GetMfcc().IsPolishDiff = true
	} else {
		cp.GetMfcc().IsPolishDiff = false
	}

	if pm.isDiffPowerSpectrum {
		cp.GetMfcc().IsDiffPowerSpectrum = true
	} else {
		cp.GetMfcc().IsDiffPowerSpectrum = false
	}

	if pm.isPredDiffAmplSpectrum {
		cp.GetMfcc().IsPredDiffAmpSpetrum = true
	} else {
		cp.GetMfcc().IsPredDiffAmpSpetrum = false
	}

	if pm.isEnergyNorm {
		cp.GetMfcc().IsEnergyNorm = true
	} else {
		cp.GetMfcc().IsEnergyNorm = false
	}

	if pm.isEnergyNorm {
		cp.GetMfcc().SilFloor = pm.silFloor
	} else {
		cp.GetMfcc().SilFloor = constant.SIL_FLOOR
	}

	if pm.isEnergyNorm {
		cp.GetMfcc().EnergyScale = pm.energyscale
	} else {
		cp.GetMfcc().EnergyScale = constant.ENERGY_SCALE
	}

	if pm.isFeatWarping {
		cp.GetMfcc().IsFeatWarping = true
	} else {
		cp.GetMfcc().IsFeatWarping = false
	}

	if pm.isFeatWarping {
		cp.GetMfcc().FeatWarpWinSize = pm.featWarpWinSize
	} else {
		cp.GetMfcc().FeatWarpWinSize = constant.FEATURE_WARPING_WIN_SIZE
	}

	if pm.isRasta {
		cp.GetMfcc().IsRasta = true
	} else {
		cp.GetMfcc().IsRasta = false
	}

	cp.GetMfcc().RastaCoff = pm.rastaCoff

	if nil != cp.Wav2Mfcc(p, info, &para, &icol, &irow) && irow < constant.MIN_FRAMES {
		return -2, fmt.Errorf("Feature Extract error -2")
	}

	gmm.VectorSize = icol
	gmm.Frames = irow
	gmm.FeatureData = make([][]float32, gmm.Frames, gmm.Frames)
	for i := 0; i < gmm.Frames; i++ {
		gmm.FeatureData[i] = make([]float32, gmm.VectorSize, gmm.VectorSize)
	}

	for ii := 0; ii < irow; ii++ {
		for jj := 0; jj < icol; jj++ {
			gmm.FeatureData[ii][jj] = para[ii*icol+jj]
		}
	}

	// CMS & CVN
	if pm.cmsvn {
		if err = cp.FeatureNorm(gmm.FeatureData, icol, irow); err != nil {
			log.Error(err)
			return -3, fmt.Errorf("Feature Extract error -3")
		}
	}

	return irow, nil
}
