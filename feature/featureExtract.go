package feature

import (
	"fmt"
	"govpr/constant"
	"govpr/gmm"
	"govpr/param"
	"govpr/waveIO"
)

//-------------------------------------------------------------------------------
//            the implementation for extracting feature parameter
//buffer	  : data buffer for storing feature parameters(must allocate memory before)
//cFileName   : the file path of utterance to be processed
//return: frame number if succeed, else return 0
func FeatureExtract(pBuf []int16, gmm *gmm.GMM) (int, error) {
	var p, para []float32
	var info waveIO.WavInfo
	var cparam *param.CParam = param.NewCParam()
	var paramconf *ParamConf = new(ParamConf)
	var err error
	var icol, irow int
	var buflen int64 = int64(len(pBuf))

	p = make([]float32, buflen, buflen)
	for i := int64(0); i < buflen; i++ {
		p[i] = float32(pBuf[i])
	}

	paramconf.bHTK = constant.PARAMCONF_BHTK
	paramconf.lcut = constant.PARAMCONF_LCUT
	paramconf.hcut = constant.PARAMCONF_HCUT
	paramconf.nfb = constant.PARAMCONF_NFB
	paramconf.nflen = constant.PARAMCONF_NFLEN
	paramconf.nfsft = constant.PARAMCONF_NFSFT
	paramconf.nmfcc = constant.PARAMCONF_NMFCC

	paramconf.bs0 = constant.PARAMCONF_BS0
	paramconf.bd0 = constant.PARAMCONF_BD0
	paramconf.ba0 = constant.PARAMCONF_BA0
	paramconf.bs = constant.PARAMCONF_BS
	paramconf.bd = constant.PARAMCONF_BD
	paramconf.ba = constant.PARAMCONF_BA

	paramconf.cmsvn = constant.PARAMCONF_CMSVN
	paramconf.ZeroGlobalMean = constant.PARAMCONF_ZEROGLOBALMEAN
	paramconf.bDiffpolish = constant.PARAMCONF_DIFPOL
	paramconf.bdpscc = constant.PARAMCONF_DPSCC
	paramconf.bpdascc = constant.PARAMCONF_PDASCC
	paramconf.bEnergyNorm = constant.PARAMCONF_ENERGYNORM
	paramconf.silFloor = constant.PARAMCONF_SILFLOOR

	paramconf.energyscale = constant.PARAMCONF_ENERGYSCALE
	paramconf.bFeatWarping = constant.PARAMCONF_FEATWARP
	paramconf.featWarpWinSize = constant.PARAMCONF_FEATWARP_WIN
	paramconf.bdBNorm = constant.PARAMCONF_DBNORM
	paramconf.bRasta = constant.PARAMCONF_RASTA
	paramconf.rastaCoff = constant.PARAMCONF_RASTA_COFF

	if paramconf.bHTK {
		return 0, fmt.Errorf("paramconf does not support htk format")
	}

	/************************************************************************/
	/*
		WaveLoad完成三项工作
		1.得到语音缓冲
		2.得到采样率
		3.得到采样点数
	*/
	/************************************************************************/

	info.SampleRate = constant.SAMPLERATE
	info.Length = buflen
	info.BitSPSample = constant.BIT_PER_SAMPLE

	if paramconf.hcut > paramconf.lcut {
		err = cparam.InitFBank2(int(info.SampleRate), paramconf.nflen, paramconf.nfb, int(paramconf.lcut), int(paramconf.hcut))
	} else {
		err = cparam.InitFBank(int(info.SampleRate), paramconf.nflen, paramconf.nfb)
	}

	if err != nil {
		return 0, err
	}

	err = cparam.InitMFCC(paramconf.nmfcc, float32(paramconf.nfsft))
	if err != nil {
		return 0, err
	}

	if paramconf.bs0 {
		cparam.MfccInfo.B0 = true
	} else {
		cparam.MfccInfo.B0 = false
	}

	if paramconf.bs {
		cparam.MfccInfo.BStatic = true
	} else {
		cparam.MfccInfo.BStatic = false
	}

	if paramconf.bd0 {
		cparam.MfccInfo.BD0 = true
	} else {
		cparam.MfccInfo.BD0 = false
	}

	if paramconf.bd {
		cparam.MfccInfo.BDelta = true
	} else {
		cparam.MfccInfo.BDelta = false
	}

	if paramconf.ba0 {
		cparam.MfccInfo.BA0 = true
	} else {
		cparam.MfccInfo.BA0 = false
	}

	if paramconf.ba {
		cparam.MfccInfo.BAcce = true
	} else {
		cparam.MfccInfo.BAcce = false
	}

	if paramconf.ZeroGlobalMean {
		cparam.MfccInfo.ZeroGlobalMean = true
	} else {
		cparam.MfccInfo.ZeroGlobalMean = false
	}

	if paramconf.bdBNorm {
		cparam.MfccInfo.BdBNorm = true
	} else {
		cparam.MfccInfo.BdBNorm = false
	}

	if paramconf.bDiffpolish {
		cparam.MfccInfo.BPolishDiff = true
	} else {
		cparam.MfccInfo.BPolishDiff = false
	}

	if paramconf.bdpscc {
		cparam.MfccInfo.BDPSCC = true
	} else {
		cparam.MfccInfo.BDPSCC = false
	}

	if paramconf.bpdascc {
		cparam.MfccInfo.BPDASCC = true
	} else {
		cparam.MfccInfo.BPDASCC = false
	}

	if paramconf.bEnergyNorm {
		cparam.MfccInfo.BEnergyNorm = true
	} else {
		cparam.MfccInfo.BEnergyNorm = false
	}

	if paramconf.bEnergyNorm {
		cparam.MfccInfo.SilFloor = paramconf.silFloor
	} else {
		cparam.MfccInfo.SilFloor = constant.PARAMCONF_SILFLOOR
	}

	if paramconf.bEnergyNorm {
		cparam.MfccInfo.Energyscale = paramconf.energyscale
	} else {
		cparam.MfccInfo.Energyscale = constant.PARAMCONF_ENERGYSCALE
	}

	if paramconf.bFeatWarping {
		cparam.MfccInfo.BFeatWarping = true
	} else {
		cparam.MfccInfo.BFeatWarping = false
	}

	if paramconf.bFeatWarping {
		cparam.MfccInfo.FeatWarpWinSize = paramconf.featWarpWinSize
	} else {
		cparam.MfccInfo.FeatWarpWinSize = constant.PARAMCONF_FEATWARP_WIN
	}

	if paramconf.bRasta {
		cparam.MfccInfo.BRasta = true
	} else {
		cparam.MfccInfo.BRasta = false
	}

	cparam.MfccInfo.RastaCoff = paramconf.rastaCoff

	if nil == cparam.WAV2MFCC(p, info, para, &icol, &irow) && irow < constant.MIN_FRAMES {
		return -2, fmt.Errorf("FeatureExtract error -2")
	}

	cparam.UnInitMFCC()
	cparam.UnInitFBank()

	if gmm.BPLoaded {
		gmm.CleanUpPar()
	}

	gmm.IVectorSize = icol
	gmm.IFrames = irow
	gmm.FParam = make([][]float32, gmm.IFrames, gmm.IFrames)
	for i := 0; i < gmm.IFrames; i++ {
		gmm.FParam[i] = make([]float32, gmm.IVectorSize, gmm.IVectorSize)
	}

	//数据缓冲区申请内存
	//irow是帧数,icol是特征维数
	for ii := 0; ii < irow; ii++ {
		for jj := 0; jj < icol; jj++ {
			gmm.FParam[ii][jj] = para[ii*icol+jj]
		}
	}

	// CMS & CVN
	if paramconf.cmsvn {
		if cparam.FeatureNorm(gmm.FParam, icol, irow) != nil {
			return -3, fmt.Errorf("FeatureExtract error -3")
		}
	}

	gmm.BPLoaded = true

	return irow, nil
}

func FeatureExtract2(cFileName string, gmm *gmm.GMM) (int, error) {
	var p, para []float32
	var info waveIO.WavInfo
	var cparam *param.CParam = param.NewCParam()
	var paramconf *ParamConf = new(ParamConf)
	var err error
	var icol, irow int

	paramconf.bHTK = constant.PARAMCONF_BHTK
	paramconf.lcut = constant.PARAMCONF_LCUT
	paramconf.hcut = constant.PARAMCONF_HCUT
	paramconf.nfb = constant.PARAMCONF_NFB
	paramconf.nflen = constant.PARAMCONF_NFLEN
	paramconf.nfsft = constant.PARAMCONF_NFSFT
	paramconf.nmfcc = constant.PARAMCONF_NMFCC

	paramconf.bs0 = constant.PARAMCONF_BS0
	paramconf.bd0 = constant.PARAMCONF_BD0
	paramconf.ba0 = constant.PARAMCONF_BA0
	paramconf.bs = constant.PARAMCONF_BS
	paramconf.bd = constant.PARAMCONF_BD
	paramconf.ba = constant.PARAMCONF_BA

	paramconf.cmsvn = constant.PARAMCONF_CMSVN
	paramconf.ZeroGlobalMean = constant.PARAMCONF_ZEROGLOBALMEAN
	paramconf.bDiffpolish = constant.PARAMCONF_DIFPOL
	paramconf.bdpscc = constant.PARAMCONF_DPSCC
	paramconf.bpdascc = constant.PARAMCONF_PDASCC
	paramconf.bEnergyNorm = constant.PARAMCONF_ENERGYNORM
	paramconf.silFloor = constant.PARAMCONF_SILFLOOR

	paramconf.energyscale = constant.PARAMCONF_ENERGYSCALE
	paramconf.bFeatWarping = constant.PARAMCONF_FEATWARP
	paramconf.featWarpWinSize = constant.PARAMCONF_FEATWARP_WIN
	paramconf.bdBNorm = constant.PARAMCONF_DBNORM
	paramconf.bRasta = constant.PARAMCONF_RASTA
	paramconf.rastaCoff = constant.PARAMCONF_RASTA_COFF

	if paramconf.bHTK {
		return 0, fmt.Errorf("paramconf does not support htk format")
	}

	/************************************************************************/
	/*
		WaveLoad完成三项工作
		1.得到语音缓冲
		2.得到采样率
		3.得到采样点数
	*/
	/************************************************************************/

	if err = waveIO.WaveLoad(cFileName, &p, &info); err != nil {
		return -1, err
	}

	if paramconf.hcut > paramconf.lcut {
		err = cparam.InitFBank2(int(info.SampleRate), paramconf.nflen, paramconf.nfb, int(paramconf.lcut), int(paramconf.hcut))
	} else {
		err = cparam.InitFBank(int(info.SampleRate), paramconf.nflen, paramconf.nfb)
	}

	if err != nil {
		return 0, err
	}

	err = cparam.InitMFCC(paramconf.nmfcc, float32(paramconf.nfsft))
	if err != nil {
		return 0, err
	}

	if paramconf.bs0 {
		cparam.MfccInfo.B0 = true
	} else {
		cparam.MfccInfo.B0 = false
	}

	if paramconf.bs {
		cparam.MfccInfo.BStatic = true
	} else {
		cparam.MfccInfo.BStatic = false
	}

	if paramconf.bd0 {
		cparam.MfccInfo.BD0 = true
	} else {
		cparam.MfccInfo.BD0 = false
	}

	if paramconf.bd {
		cparam.MfccInfo.BDelta = true
	} else {
		cparam.MfccInfo.BDelta = false
	}

	if paramconf.ba0 {
		cparam.MfccInfo.BA0 = true
	} else {
		cparam.MfccInfo.BA0 = false
	}

	if paramconf.ba {
		cparam.MfccInfo.BAcce = true
	} else {
		cparam.MfccInfo.BAcce = false
	}

	if paramconf.ZeroGlobalMean {
		cparam.MfccInfo.ZeroGlobalMean = true
	} else {
		cparam.MfccInfo.ZeroGlobalMean = false
	}

	if paramconf.bdBNorm {
		cparam.MfccInfo.BdBNorm = true
	} else {
		cparam.MfccInfo.BdBNorm = false
	}

	if paramconf.bDiffpolish {
		cparam.MfccInfo.BPolishDiff = true
	} else {
		cparam.MfccInfo.BPolishDiff = false
	}

	if paramconf.bdpscc {
		cparam.MfccInfo.BDPSCC = true
	} else {
		cparam.MfccInfo.BDPSCC = false
	}

	if paramconf.bpdascc {
		cparam.MfccInfo.BPDASCC = true
	} else {
		cparam.MfccInfo.BPDASCC = false
	}

	if paramconf.bEnergyNorm {
		cparam.MfccInfo.BEnergyNorm = true
	} else {
		cparam.MfccInfo.BEnergyNorm = false
	}

	if paramconf.bEnergyNorm {
		cparam.MfccInfo.SilFloor = paramconf.silFloor
	} else {
		cparam.MfccInfo.SilFloor = constant.PARAMCONF_SILFLOOR
	}

	if paramconf.bEnergyNorm {
		cparam.MfccInfo.Energyscale = paramconf.energyscale
	} else {
		cparam.MfccInfo.Energyscale = constant.PARAMCONF_ENERGYSCALE
	}

	if paramconf.bFeatWarping {
		cparam.MfccInfo.BFeatWarping = true
	} else {
		cparam.MfccInfo.BFeatWarping = false
	}

	if paramconf.bFeatWarping {
		cparam.MfccInfo.FeatWarpWinSize = paramconf.featWarpWinSize
	} else {
		cparam.MfccInfo.FeatWarpWinSize = constant.PARAMCONF_FEATWARP_WIN
	}

	if paramconf.bRasta {
		cparam.MfccInfo.BRasta = true
	} else {
		cparam.MfccInfo.BRasta = false
	}

	cparam.MfccInfo.RastaCoff = paramconf.rastaCoff

	if nil == cparam.WAV2MFCC(p, info, para, &icol, &irow) && irow < constant.MIN_FRAMES {
		return -2, fmt.Errorf("FeatureExtract error -2")
	}

	cparam.UnInitMFCC()
	cparam.UnInitFBank()

	if gmm.BPLoaded {
		gmm.CleanUpPar()
	}

	gmm.IVectorSize = icol
	gmm.IFrames = irow
	gmm.FParam = make([][]float32, gmm.IFrames, gmm.IFrames)
	for i := 0; i < gmm.IFrames; i++ {
		gmm.FParam[i] = make([]float32, gmm.IVectorSize, gmm.IVectorSize)
	}

	//数据缓冲区申请内存
	//irow是帧数,icol是特征维数
	for ii := 0; ii < irow; ii++ {
		for jj := 0; jj < icol; jj++ {
			gmm.FParam[ii][jj] = para[ii*icol+jj]
		}
	}

	// CMS & CVN
	if paramconf.cmsvn {
		if cparam.FeatureNorm(gmm.FParam, icol, irow) != nil {
			return -3, fmt.Errorf("FeatureExtract error -3")
		}
	}

	gmm.BPLoaded = true

	return irow, nil
}
