package feature

import (
	"fmt"
	"github.com/liuxp0827/govpr/constant"
	"github.com/liuxp0827/govpr/gmm"
	"github.com/liuxp0827/govpr/log"
	"github.com/liuxp0827/govpr/param"
	"github.com/liuxp0827/govpr/waveIO"
)

func Extract(data []int16, gmm *gmm.GMM) (int, error) {
	var p, para []float32
	var info waveIO.WavInfo
	var cp *param.CParam = param.NewCParam()
	var param *Param = new(Param)
	var err error
	var icol, irow int
	var buflen int = len(data)

	p = make([]float32, buflen, buflen)
	for i := 0; i < buflen; i++ {
		p[i] = float32(data[i])
	}

	param.lowCutOff = constant.LOW_CUT_OFF
	param.highCutOff = constant.HIGH_CUT_OFF
	param.filterBankSize = constant.FILTER_BANK_SIZE
	param.frameLength = constant.FRAME_LENGTH
	param.frameShift = constant.FRAME_SHIFTt
	param.mfccOrder = constant.MFCC_ORDER

	param.isStatic = constant.BSTATIC
	param.isDynamic = constant.BDYNAMIC
	param.isAcce = constant.BACCE

	param.cmsvn = constant.CMSVN
	param.isZeroGlobalMean = constant.ZEROGLOBALMEAN
	param.isDiffPolish = constant.DIFPOL
	param.isDiffPowerSpectrum = constant.DPSCC
	param.isPredDiffAmplSpectrum = constant.PDASCC
	param.isEnergyNorm = constant.ENERGYNORM
	param.silFloor = constant.SIL_FLOOR

	param.energyscale = constant.ENERGY_SCALE
	param.isFeatWarping = constant.FEATWARP
	param.featWarpWinSize = constant.FEATURE_WARPING_WIN_SIZE
	param.isDBNorm = constant.DBNORM
	param.isRasta = constant.RASTA
	param.rastaCoff = constant.RASTA_COFF

	info.SampleRate = constant.SAMPLERATE
	info.Length = int64(buflen)
	info.BitSPSample = constant.BIT_PER_SAMPLE

	if param.highCutOff > param.lowCutOff {
		err = cp.InitFBank2(int(info.SampleRate), param.frameLength, param.filterBankSize, int(param.lowCutOff), int(param.highCutOff))
	} else {
		err = cp.InitFBank(int(info.SampleRate), param.frameLength, param.filterBankSize)
	}

	if err != nil {
		return 0, err
	}

	err = cp.InitMfcc(param.mfccOrder, float32(param.frameShift))
	if err != nil {
		return 0, err
	}

	if param.isStatic {
		cp.GetMfcc().IsStatic = true
	} else {
		cp.GetMfcc().IsStatic = false
	}

	if param.isDynamic {
		cp.GetMfcc().IsDynamic = true
	} else {
		cp.GetMfcc().IsDynamic = false
	}

	if param.isAcce {
		cp.GetMfcc().IsAcce = true
	} else {
		cp.GetMfcc().IsAcce = false
	}

	if param.isZeroGlobalMean {
		cp.GetMfcc().IsZeroGlobalMean = true
	} else {
		cp.GetMfcc().IsZeroGlobalMean = false
	}

	if param.isDBNorm {
		cp.GetMfcc().IsDBNorm = true
	} else {
		cp.GetMfcc().IsDBNorm = false
	}

	if param.isDiffPolish {
		cp.GetMfcc().IsPolishDiff = true
	} else {
		cp.GetMfcc().IsPolishDiff = false
	}

	if param.isDiffPowerSpectrum {
		cp.GetMfcc().IsDiffPowerSpectrum = true
	} else {
		cp.GetMfcc().IsDiffPowerSpectrum = false
	}

	if param.isPredDiffAmplSpectrum {
		cp.GetMfcc().IsPredDiffAmpSpetrum = true
	} else {
		cp.GetMfcc().IsPredDiffAmpSpetrum = false
	}

	if param.isEnergyNorm {
		cp.GetMfcc().IsEnergyNorm = true
	} else {
		cp.GetMfcc().IsEnergyNorm = false
	}

	if param.isEnergyNorm {
		cp.GetMfcc().SilFloor = param.silFloor
	} else {
		cp.GetMfcc().SilFloor = constant.SIL_FLOOR
	}

	if param.isEnergyNorm {
		cp.GetMfcc().EnergyScale = param.energyscale
	} else {
		cp.GetMfcc().EnergyScale = constant.ENERGY_SCALE
	}

	if param.isFeatWarping {
		cp.GetMfcc().IsFeatWarping = true
	} else {
		cp.GetMfcc().IsFeatWarping = false
	}

	if param.isFeatWarping {
		cp.GetMfcc().FeatWarpWinSize = param.featWarpWinSize
	} else {
		cp.GetMfcc().FeatWarpWinSize = constant.FEATURE_WARPING_WIN_SIZE
	}

	if param.isRasta {
		cp.GetMfcc().IsRasta = true
	} else {
		cp.GetMfcc().IsRasta = false
	}

	cp.GetMfcc().RastaCoff = param.rastaCoff

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
	if param.cmsvn {
		if err = cp.FeatureNorm(gmm.FeatureData, icol, irow); err != nil {
			log.Error(err)
			return -3, fmt.Errorf("Feature Extract error -3")
		}
	}

	return irow, nil
}
