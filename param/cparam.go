package param

import (
	"fmt"
	gomath"github.com/liuxp0827/govpr/math"
	"github.com/liuxp0827/govpr/constant"
	"github.com/liuxp0827/govpr/waveIO"
	"math"
)

type CParam struct {
	filterBank       *FilterBank
	mfcc             *Mfcc
	cepLifterWinSize []float32
	hammingWinSize   []float64 // vector of the hamming window
	warpWinLength    int       // warping window size
	warpTable        []float32 // warping probability table
}

func NewCParam() *CParam {
	return &CParam{
		warpWinLength: 300,
	}
}

func (cp *CParam) GetMfcc() *Mfcc {
	return cp.mfcc
}

// Initialize the filter bank info struct. User should
//  call this function before calling wav2MFCC().
// - Arguments -
//   sampleRate : Sample rate (samples per second)
//   frameRate : Frame length in ms
//   filterBankSize : Number of filter banks
//   lowCutFreq : Low cut-off frequency (in Hz) default = -2
//   highCutFreq : High cut-off frequency (in Hz)default = -1
func (cp *CParam) InitFBank(sampleRate, frameRate int, filterBankSize int) error {
	return cp.InitFBank2(sampleRate, frameRate, filterBankSize, -2, -1)
}

func (cp *CParam) InitFBank2(sampleRate, frameRate int, filterBankSize, lowCutFreq, highCutFreq int) error {
	var melLowCutFreq, melHighCutFreq, melStep, curFreq, fa, fb, fc float32

	if lowCutFreq >= highCutFreq || highCutFreq > (sampleRate>>1) || lowCutFreq > (sampleRate>>1) {
		return fmt.Errorf("Low and High cut off frequencies set incorrectly")
	}

	// check number of filter bancks
	if cp.mfcc != nil {
		if cp.mfcc.mfccOrder+1 > cp.filterBank.filterBankSize {
			return fmt.Errorf("param order nb greater than filter bank nb")
		}
	}

	cp.filterBank = new(FilterBank)

	// given by arguments
	cp.filterBank.sampleRate = sampleRate
	cp.filterBank.frameSize = int(float32(sampleRate) * float32(frameRate) * 1e-3)
	cp.filterBank.filterBankSize = filterBankSize

	// calculated from arguments
	cp.filterBank.fttSize = 2
	for cp.filterBank.frameSize > cp.filterBank.fttSize {
		cp.filterBank.fttSize <<= 1
	}

	fttIndex := cp.filterBank.fttSize >> 1
	cp.filterBank.fttResolution = float32(sampleRate) / float32(cp.filterBank.fttSize)

	// the low and high cut-off indices
	if lowCutFreq < 0 {
		cp.filterBank.start = 0
		lowCutFreq = 0
	} else {
		cp.filterBank.start = int(lowCutFreq / int(cp.filterBank.fttResolution))
	}

	if highCutFreq < 0 {
		highCutFreq = sampleRate >> 1
		cp.filterBank.end = fttIndex
	} else {
		cp.filterBank.end = int(float32(highCutFreq) / cp.filterBank.fttResolution)
		if cp.filterBank.end > fttIndex {
			cp.filterBank.end = fttIndex
		}
	}

	cp.filterBank.centerFreqs = nil
	cp.filterBank.lowerFilterBanksIndex = nil
	cp.filterBank.lowerFilterBanksWeight = nil
	cp.filterBank.fftRealValue = nil
	cp.filterBank.fftComplexValue = nil

	// the center frequencies
	cp.filterBank.centerFreqs = make([]float32, filterBankSize+1, filterBankSize+1)

	melLowCutFreq = cp.mel(float32(lowCutFreq))
	melHighCutFreq = cp.mel(float32(highCutFreq))
	melStep = (melHighCutFreq - melLowCutFreq) / float32(filterBankSize+1)
	cp.filterBank.centerFreqs[0] = float32(lowCutFreq) // the zero index is the low cut-off

	for i := 1; i <= filterBankSize; i++ {
		cp.filterBank.centerFreqs[i] = melLowCutFreq + melStep*float32(i)
		cp.filterBank.centerFreqs[i] = cp.freq(cp.filterBank.centerFreqs[i])
	}

	// lower channel indices
	cp.filterBank.lowerFilterBanksIndex = make([]int16, fttIndex, fttIndex)

	for i, ichan := 0, 0; i < fttIndex; i++ {
		curFreq = float32(i) * cp.filterBank.fttResolution
		if i < cp.filterBank.start || i > cp.filterBank.end {
			cp.filterBank.lowerFilterBanksIndex[i] = -1
		} else {
			for ichan <= filterBankSize && cp.filterBank.centerFreqs[ichan] <= curFreq {
				ichan++
			}
			cp.filterBank.lowerFilterBanksIndex[i] = int16(ichan - 1)
		}
	}

	// lower channel weights
	cp.filterBank.lowerFilterBanksWeight = make([]float32, fttIndex, fttIndex)
	for i := 0; i < fttIndex; i++ {
		curFreq = float32(i) * cp.filterBank.fttResolution
		if cp.filterBank.lowerFilterBanksIndex[i] == -1 {
			cp.filterBank.lowerFilterBanksWeight[i] = 0.0
		} else {
			if int(cp.filterBank.lowerFilterBanksIndex[i]) < filterBankSize {
				fc = cp.filterBank.centerFreqs[cp.filterBank.lowerFilterBanksIndex[i]+1]
			} else {
				fc = float32(highCutFreq)
			}

			fa = 1 / (cp.filterBank.centerFreqs[cp.filterBank.lowerFilterBanksIndex[i]] - fc)
			fb = -fa * fc
			cp.filterBank.lowerFilterBanksWeight[i] = fa*curFreq + fb
		}
	}

	// alloc memory for data buffer
	cp.filterBank.fftRealValue = make([]float64, cp.filterBank.fttSize, cp.filterBank.fttSize)
	cp.filterBank.fftComplexValue = make([]float64, cp.filterBank.fttSize, cp.filterBank.fttSize)

	// the defaults
	cp.filterBank.isLogFBChannels = true
	cp.filterBank.isUsePower = false
	cp.filterBank.isPreEmphasize = true
	cp.filterBank.isUseHamming = true
	return nil
}


// Initialize the mfcc info struct. User should call
//  this function before calling wav2MFCC().
// - Arguments -
//      iOrder : MFCC order (except the 0th)
//  fFrameRate : MFCC frame rate in ms

func (cp *CParam) InitMfcc(iOrder int, fFrmRate float32) error {

	cp.mfcc = new(Mfcc)

	if cp.filterBank != nil {
		if iOrder+1 > cp.filterBank.filterBankSize {
			return fmt.Errorf("param order nb greater than filter bank nb")
		}
	}

	cp.mfcc.mfccOrder = iOrder   // mfcc order, except the 0th
	cp.mfcc.FrameRate = fFrmRate // frame rate in ms

	// the defaults
	cp.mfcc.dynamicWinSize = 2
	cp.mfcc.isFilter = false
	cp.mfcc.IsStatic = true
	cp.mfcc.IsDynamic = true
	cp.mfcc.IsAcce = false
	cp.mfcc.IsLiftCepstral = true
	cp.mfcc.CepstralLifter = 22

	return nil
}

// wave -> mfcc.
// - Arguments -
//       data : buffer for wave data sequence
//      fParam : buffer for storing the converted parameters,
//               memory is alloced within the function, so the
//               CALLER is RESPONSIBLE to free the memory.
//        col : width of the param vector
//        row : length of the param vector sequence

func (cp *CParam) Wav2Mfcc(data []float32, wavinfo waveIO.WavInfo, fParam *[]float32, col, row *int) error {

	if cp.mfcc.IsZeroGlobalMean {
		cp.IsZeroGlobalMean(data, wavinfo.Length)
	}

	if cp.mfcc.IsDBNorm {
		cp.dBNorm(data, wavinfo.Length)
	}

	if cp.filterBank == nil || cp.mfcc == nil {
		return fmt.Errorf("Filter bank info and MFCC info not initialized")
	}

	var width, iFrameRate int
	var fttIndex int = cp.filterBank.fttSize >> 1
	var melfloor float32 = float32(1.0)
	var fstatic []float32
	var err error

	// calculate number of rows (frames)
	iFrameRate = int(1e-3 * float32(cp.mfcc.FrameRate) * float32(cp.filterBank.sampleRate))

	if iFrameRate > cp.filterBank.frameSize {
		return fmt.Errorf("Sample point equal to zero")
	}

	*row = int((wavinfo.Length - int64(cp.filterBank.frameSize-iFrameRate)) / int64(iFrameRate))

	// buffer for raw static params (include the 0th coef)
	width = cp.mfcc.mfccOrder + 1

	fstatic = make([]float32, (*row)*width, (*row)*width)

	// buffer for filter banks

	for i := 0; i < *row; i++ {
		cp.filterBank.fftRealValue = make([]float64, cp.filterBank.fttSize, cp.filterBank.fttSize)
		cp.filterBank.fftComplexValue = make([]float64, cp.filterBank.fttSize, cp.filterBank.fttSize)

		for j := 0; j < cp.filterBank.frameSize; j++ {
			cp.filterBank.fftRealValue[j] = float64(data[i*iFrameRate+j])
		}

		// Do pre-emphasis
		if cp.filterBank.isPreEmphasize {
			cp.preEmphasise(cp.filterBank.fftRealValue, cp.filterBank.frameSize)
		}

		// Do hamming
		if cp.filterBank.isUseHamming {
			cp.doHamming(cp.filterBank.fftRealValue, cp.filterBank.frameSize)
		}

		// take fft
		err = gomath.FFT(cp.filterBank.fftRealValue, cp.filterBank.fftComplexValue, cp.filterBank.fttSize)
		if err != nil {
			return err
		}

		var filterBank []float64 = make([]float64, cp.filterBank.filterBankSize)
		for j := 0; j < fttIndex; j++ {
			cp.filterBank.fftRealValue[j] =
				cp.filterBank.fftRealValue[j]*cp.filterBank.fftRealValue[j] +
					cp.filterBank.fftComplexValue[j]*cp.filterBank.fftComplexValue[j]
		}

		//	Differential Power Spectrum
		if cp.filterBank.isUsePower && cp.mfcc.IsDiffPowerSpectrum {
			cp.DPSCC(fttIndex)
		}

		// use power or amp
		if !cp.filterBank.isUsePower {
			for j := 0; j < fttIndex; j++ {
				cp.filterBank.fftRealValue[j] = math.Sqrt(cp.filterBank.fftRealValue[j])
			}
		}

		// Predictive Differential Amplitude Spectrum
		if !cp.filterBank.isUsePower && cp.mfcc.IsPredDiffAmpSpetrum {
			cp.PDASCC(fttIndex)
		}

		// accumulate filter banks
		for j := 0; j < fttIndex; j++ {
			if cp.filterBank.lowerFilterBanksIndex[j] < 0 {
				continue
			}

			// accumulate the lower bank
			if cp.filterBank.lowerFilterBanksIndex[j] != 0 {
				filterBank[cp.filterBank.lowerFilterBanksIndex[j]-1] +=
					cp.filterBank.fftRealValue[j] * float64(cp.filterBank.lowerFilterBanksWeight[j])
			}
			// accumulate the upper bank
			if int(cp.filterBank.lowerFilterBanksIndex[j]) < cp.filterBank.filterBankSize {
				filterBank[cp.filterBank.lowerFilterBanksIndex[j]] +=
					cp.filterBank.fftRealValue[j] * float64(1-cp.filterBank.lowerFilterBanksWeight[j])
			}
		}

		// take logs
		if cp.filterBank.isLogFBChannels {
			for j := 0; j < cp.filterBank.filterBankSize; j++ {
				if filterBank[j] >= float64(melfloor) {
					filterBank[j] = math.Log(filterBank[j])
				} else {
					filterBank[j] = math.Log(float64(melfloor))
				}
			}
		}

		// take dct
		if !cp.mfcc.isFilter {
			err = gomath.DCT(filterBank, &width)
			if err != nil {
				return err
			}

			// Liftering
			if cp.mfcc.IsLiftCepstral {
				err = cp.liftCepstral(filterBank)
				if err != nil {
					return err
				}
			}
		}

		// copy data
		for j := 0; j < width; j++ {
			fstatic[i*width+j] = float32(filterBank[j])
		}
	}

	if cp.mfcc.IsFeatWarping {
		err = cp.warping(fstatic, width, row, width, int(cp.mfcc.FeatWarpWinSize))
		if err != nil {
			return err
		}
	}

	if cp.mfcc.IsRasta {
		err := cp.rastaFiltering(fstatic, width, row, width)
		if err != nil {
			return err
		}
	}

	if cp.mfcc.IsEnergyNorm {
		cp.energyNorm(fstatic, width, *row)
	}

	*fParam, err = cp.static2Full(fstatic, &width, row)

	*col = width

	return nil
}

// Calculate the requested parameters from the raw static coefficients.
//  This function convert raw static coef. to conjunct coef. (0th, static,
//  delta, acce or any conbinations of them) specified in pmfccinfo.
//  In calculating deltas, some heading and tailing frames may be discarded
//  in this function. The caller is responsible to to free the memory of
//  the returned dbuffer.
// - Arguments -
//     fStatic : buffer for raw static coef. (include the 0th)
//        col : width of the raw static coef., and is changed
//               by this function to be the width of the conjunct
//               params.
//        row : frames of the raw static coef. and is recalculated
//               in this function to be the actual number of frames
//                of the conjunct params.
func (cp *CParam) static2Full(fstatic []float32, col, row *int) ([]float32, error) {
	var width int = *col
	var iSOff, iDOff, ipt int
	var fdelta, facce, fParam []float32
	var err error

	if cp.mfcc == nil {
		return nil, fmt.Errorf("MFCC info not initialized")
	}

	// take deltas from statics
	if cp.mfcc.IsDynamic {
		fdelta = make([]float32, width*(*row), width*(*row))
		err = cp.doDelta(fdelta, fstatic, row, width)
		if err != nil {
			return nil, err
		}

		iSOff = cp.mfcc.dynamicWinSize
		iDOff = 0
	}

	// take accelerations from deltaes
	if cp.mfcc.IsAcce {
		facce = make([]float32, width*(*row), width*(*row))
		err = cp.doDelta(facce, fdelta, row, width)
		if err != nil {
			return nil, err
		}

		iSOff = 2 * cp.mfcc.dynamicWinSize
		iDOff = cp.mfcc.dynamicWinSize
	}

	iSOff, iDOff = 0, 0

	// calculate the actual width of the conjunct parameter
	*col = 0

	if cp.mfcc.IsStatic {
		*col += cp.mfcc.mfccOrder
	}

	if cp.mfcc.IsDynamic {
		*col += cp.mfcc.mfccOrder
	}

	if cp.mfcc.IsAcce {
		*col += cp.mfcc.mfccOrder
	}

	// prepare for parameter buffer
	fParam = make([]float32, (*col)*(*row), (*col)*(*row))

	for i := 0; i < *row; i++ {

		if cp.mfcc.IsStatic {
			for j := 1; j < width; j++ {
				fParam[ipt] = fstatic[(i+iSOff)*width+j]
				ipt++
			}
		}

		if cp.mfcc.IsDynamic {
			for j := 1; j < width; j++ {
				fParam[ipt] = fdelta[(i+iDOff)*width+j]
				ipt++
			}
		}

		if cp.mfcc.IsAcce {
			for j := 1; j < width; j++ {
				fParam[ipt] = facce[i*width+j]
				ipt++
			}
		}
	}

	return fParam, nil
}

// ------------- Cepstral Mean Substraction & Variance Normalisation ------------
//This function normalizes the mfcc feature parameters into a Guassian
//distribute,which can reduce the influence of channel.
//    fParam   : buffer which stored feature parameters
//	iVecsize   : size of a feature vector which stored parameter
//  iVecNum    : number of feature vectors
func (cp *CParam) FeatureNorm(fParam [][]float32, iVecSize, iVecNum int) error {
	if iVecSize <= 0 {
		return fmt.Errorf("Dimension of GMM less than zero")
	}

	if iVecNum <= 0 {
		fmt.Errorf("Nb of frames less than zero")
	}

	var cmsMean []float32 = make([]float32, iVecSize, iVecSize)
	var cmsStdv []float32 = make([]float32, iVecSize, iVecSize)
	var tempMean, tempStdv float32 = 0, 0

	//Get the average value of the mV
	for i := 0; i < iVecSize/2; i++ {
		for j := 0; j < iVecNum; j++ {
			tempMean += fParam[j][i]
			tempStdv += fParam[j][i] * fParam[j][i]
		}

		cmsMean[i] = tempMean / float32(iVecNum)

		//Get the standard deviations
		cmsStdv[i] = tempStdv / float32(iVecNum)
		cmsStdv[i] -= cmsMean[i] * cmsMean[i]

		if cmsStdv[i] <= 0 {
			cmsStdv[i] = 1.0
		} else {
			cmsStdv[i] = float32(math.Sqrt(float64(cmsStdv[i])))
		}

		tempMean = 0
		tempStdv = 0
	}

	//subtract the average value
	for i := 0; i < iVecSize/2; i++ {
		for j := 0; j < iVecNum; j++ {
			fParam[j][i] = (fParam[j][i] - cmsMean[i]) / cmsStdv[i]
		}
	}

	return nil
}

func (cp *CParam) FeatureNorm2(fParam []float32, iVecSize, iVecNum int) error {
	if iVecSize <= 0 {
		return fmt.Errorf("Dimension of GMM less than zero")
	}

	if iVecNum <= 0 {
		fmt.Errorf("Nb of frames less than zero")
	}

	var cmsMean []float32 = make([]float32, iVecSize, iVecSize)
	var cmsStdv []float32 = make([]float32, iVecSize, iVecSize)
	var tempMean, tempStdv float32 = 0, 0

	//Get the average value of the mV
	for i := 0; i < iVecSize/2; i++ {
		for j := 0; j < iVecNum; j++ {
			tempMean += fParam[j*iVecSize+i]
			tempStdv += fParam[j*iVecSize+i] * fParam[j*iVecSize+i]
		}

		cmsMean[i] = tempMean / float32(iVecNum)

		//Get the standard deviations
		cmsStdv[i] = tempStdv / float32(iVecNum)
		cmsStdv[i] -= cmsMean[i] * cmsMean[i]

		if cmsStdv[i] <= 0 {
			cmsStdv[i] = 1.0
		} else {
			cmsStdv[i] = float32(math.Sqrt(float64(cmsStdv[i])))
		}

		tempMean = 0
		tempStdv = 0
	}

	//subtract the average value
	for i := 0; i < iVecSize/2; i++ {
		for j := 0; j < iVecNum; j++ {
			fParam[j*iVecSize+i] = (fParam[j*iVecSize+i] - cmsMean[i]) / cmsStdv[i]
		}
	}

	return nil
}

//------------- Decibel Normalization -------------------------------------------
func (cp *CParam) dBNorm(sampleBuffer []float32, sampleCount int64) {
	var sampleMax float32 = -math.MaxFloat32
	for i := int64(0); i < sampleCount; i++ {
		if sampleBuffer[i] > sampleMax {
			sampleMax = sampleBuffer[i]
		}
	}

	for i := int64(0); i < sampleCount; i++ {
		sampleBuffer[i] = float32(math.Pow(10, constant.DB/20.0)) * float32(math.Pow(2, 15)-1) * sampleBuffer[i] / sampleMax
	}
}

//------------- IsZeroGlobalMean --------------------------------------------------
func (cp *CParam) IsZeroGlobalMean(data []float32, sampleCount int64) {
	var mean float32 = 0.0
	for i := int64(0); i < sampleCount; i++ {
		mean += data[i]
	}
	mean /= float32(sampleCount)

	for i := int64(0); i < sampleCount; i++ {
		y := data[i] - mean
		if y > 32767 {
			y = 32767
		}
		if y < -32767 {
			y = -32767
		}
		if y > 0 {
			data[i] = float32(int16(y + 0.5))
		} else {
			data[i] = float32(int16(y - 0.5))
		}
	}
}

//------------- Norm static feature ---------------------------------------------
func (cp *CParam) energyNorm(p_FeatBuf []float32, p_nVecSize, p_nFrameNum int) {
	var maxe, mine float32
	var ft []float32 = p_FeatBuf
	var index int = 0
	maxe = ft[index]

	for i := 0; i < p_nFrameNum; i++ {
		if ft[index] > maxe {
			maxe = ft[index]
		}

		index += p_nVecSize
	}

	mine = (maxe - float32(cp.mfcc.SilFloor)*float32(math.Log(10.0))) / 10.0

	for i := 0; i < p_nFrameNum; i++ {
		if ft[index] < mine {
			mine = ft[index]
		}

		ft[index] = 1.0 - (maxe-ft[index])*float32(cp.mfcc.EnergyScale)
		p_FeatBuf[index] = 1.0 - (maxe-p_FeatBuf[index])*float32(cp.mfcc.EnergyScale)
		index += p_nVecSize
	}
}

//------------- Differential Power Spectrum -------------------------------------
func (cp *CParam) DPSCC(pointNB int) {
	fttIndex := pointNB
	for j := 0; j < fttIndex; j++ {
		if j < fttIndex-1 {
			cp.filterBank.fftRealValue[j] = math.Abs(cp.filterBank.fftRealValue[j] - cp.filterBank.fftRealValue[j+1])
		} else {
			cp.filterBank.fftRealValue[j] = 0
		}
	}
}

//------------- Predictive Differential Amplitude Spectrum ----------------------
func (cp *CParam) PDASCC(pointNB int) {
	fttIndex := pointNB

	//	1.预测差分
	var WINLEN int = 6
	var damplitude []float64 = make([]float64, fttIndex, fttIndex)
	for j := 0; j < fttIndex; j++ {
		dmax := -math.MaxFloat64
		for w := 0; j+w < fttIndex && w <= WINLEN; w++ {
			dsin := math.Sin((float64(w) * constant.PI) / float64(2*WINLEN))
			dcur := cp.filterBank.fftRealValue[j+w] * dsin
			if dcur > dmax {
				dmax = dcur
			}
		}
		damplitude[j] = dmax
	}

	var alpha float64 = 1.05
	var dDright []float64 = make([]float64, fttIndex, fttIndex)
	var dDleft []float64 = make([]float64, fttIndex, fttIndex)

	for j := 0; j < fttIndex-1; j++ {
		if damplitude[j] > cp.filterBank.fftRealValue[j] && damplitude[j+1] < cp.filterBank.fftRealValue[j+1] {
			dDright[j] = cp.filterBank.fftRealValue[j] - alpha*cp.filterBank.fftRealValue[j+1]
		} else if damplitude[j] <= cp.filterBank.fftRealValue[j] && damplitude[j+1] >= cp.filterBank.fftRealValue[j+1] {
			dDright[j] = alpha*cp.filterBank.fftRealValue[j] - cp.filterBank.fftRealValue[j+1]
		} else {
			dDright[j] = cp.filterBank.fftRealValue[j] - cp.filterBank.fftRealValue[j+1]
		}
	}

	dDright[fttIndex-1] = 0.0

	for j := fttIndex - 1; j > 0; j-- {
		if damplitude[j] < cp.filterBank.fftRealValue[j] && damplitude[j-1] < cp.filterBank.fftRealValue[j-1] {
			dDleft[j] = cp.filterBank.fftRealValue[j] - alpha*cp.filterBank.fftRealValue[j-1]
		} else if damplitude[j] >= cp.filterBank.fftRealValue[j] && damplitude[j-1] >= cp.filterBank.fftRealValue[j-1] {
			dDleft[j] = alpha*cp.filterBank.fftRealValue[j] - cp.filterBank.fftRealValue[j-1]
		} else {
			dDleft[j] = cp.filterBank.fftRealValue[j] - cp.filterBank.fftRealValue[j-1]
		}
	}

	if damplitude[0] < cp.filterBank.fftRealValue[0] {
		dDleft[0] = (1.0 - alpha) * cp.filterBank.fftRealValue[0]
	} else {
		dDleft[0] = (alpha - 1.0) * cp.filterBank.fftRealValue[0]
	}

	// 2.累积过程

	var left []float64 = make([]float64, fttIndex, fttIndex)
	var right []float64 = make([]float64, fttIndex, fttIndex)
	for i := 1; i < fttIndex; i++ {
		left[i] = left[i-1] + dDleft[i-1]
	}

	for i := fttIndex - 2; i >= 0; i-- {
		right[i] = right[i+1] + dDright[i+1]
	}

	for i := 0; i < fttIndex; i++ {
		cp.filterBank.fftRealValue[i] = (left[i] + right[i]) / 2.0
	}
}

//------------- Feature Warping -------------------------------------------------
// nWinSize=300
// vSize=nStep=特征维数
// nInNum 输入特征帧数
// data 输入特征
// nOutNum 输出特征帧数
func (cp *CParam) warping(data []float32, vSize int, nInNum *int, nStep, nWinSize int) error {

	cp.createWarpTable()
	var nOutNum int = *nInNum - nWinSize
	if nOutNum <= 0 {
		return fmt.Errorf("nOutNum can not <= 0")
	}

	var warpBuf []float32 = make([]float32, nOutNum*nStep, nOutNum*nStep)
	var warpFrmNo int = 0
	var pDataIvt []float32 = make([]float32, nOutNum*nStep, nOutNum*nStep)

	for i := 0; i < nStep; i++ {
		//		var dst []float32 = make([]float32, nOutNum*nStep+i*nInNum, nOutNum*nStep+i*nInNum)
		for j := 0; j < *nInNum; j++ {
			//			dst[j] = data[j*nStep+i]
			if j < nOutNum*nStep {
				pDataIvt[j] = data[j*nStep+i]
			}
		}
	}

	var halfwin int = nWinSize >> 1
	var minus_res []float32 = make([]float32, nWinSize, nWinSize)

	for i := halfwin; i+halfwin < *nInNum; i++ {
		for k := 0; k < vSize; k++ {

			var p []float32 = append(pDataIvt,
				make([]float32, nOutNum*nStep+k*(*nInNum)-nOutNum*nStep)...)

			var curValue float32 = p[i]
			t := halfwin - i
			nIndex := 2*halfwin - 1

			for m := i - halfwin; m < i; m++ {
				minus_res[t+m] = p[m] - curValue
			}

			t = halfwin - i - 1

			for m := i + 1; m < i+halfwin; m++ {
				minus_res[t+m] = p[m] - curValue
			}

			var ui []uint = make([]uint, nWinSize, nWinSize)
			for i := 0; i < nWinSize; i++ {
				ui[i] = uint(minus_res[i])
			}

			for m := 0; m < 2*halfwin-1; m++ {
				nIndex -= int((ui[m] >> 31))
			}

			warpBuf[warpFrmNo*nStep+k] = cp.warpTable[nIndex]
		}

		for k := vSize; k < nStep; k++ {
			warpBuf[warpFrmNo*nStep+k] = data[i*nStep+k]
		}
		warpFrmNo++
	}

	*nInNum = warpFrmNo

	copy(data[:warpFrmNo*nStep], warpBuf[:warpFrmNo*nStep])

	return nil
}

//------------- Rasta-filtering -------------------------------------------------
/************************************************************************/
/*
	data : static mfcc
	vSize: order of mfcc
	nNum : frame number
	nStep: order of mfcc
*/
/************************************************************************/
func (cp *CParam) rastaFiltering(data []float32, vSize int, nNum *int, nStep int) error {
	if *nNum <= 4 {
		return fmt.Errorf("Not eoungh features for Rasta filtering")
	}

	var RastaBuf []float32 = make([]float32, (*nNum)*vSize, (*nNum)*vSize)
	for i := 0; i < *nNum-4; i++ {
		if i == 0 {
			for j := 0; j < vSize; j++ {
				RastaBuf[i*vSize+j] = 0.1 * (2.0*data[(i+4)*nStep+j] + data[(i+3)*nStep+j] -
					data[(i+1)*nStep+j] - 2.0*data[i*nStep+j])
			}
		} else {
			for j := 0; j < vSize; j++ {
				RastaBuf[i*vSize+j] = 0.1*(2.0*data[(i+4)*nStep+j]+data[(i+3)*nStep+j]-
					data[(i+1)*nStep+j]-2.0*data[i*nStep+j]) +
					float32(cp.mfcc.RastaCoff)*RastaBuf[(i-1)*vSize+j]
			}
		}
	}

	for i := 0; i < *nNum-4; i++ {
		for j := 0; j < vSize; j++ {
			data[i*nStep+j] = RastaBuf[i*vSize+j]
		}
	}

	*nNum = *nNum - 4

	return nil
}

//------------- Create warping talbe --------------------------------------------
func (cp *CParam) createWarpTable() {
	var TableBegin, TableEnd, presice float64 = -10.0, 10.0, 1.0e-5
	cp.warpTable = make([]float32, cp.warpWinLength, cp.warpWinLength)
	var rankBuf []float64 = make([]float64, cp.warpWinLength, cp.warpWinLength)

	for i := 0; i < cp.warpWinLength; i++ {
		rankBuf[i] = float64(float64(cp.warpWinLength)-0.5-float64(i)) / float64(cp.warpWinLength)
	}

	var integral float64 = 0.0
	var Index int = cp.warpWinLength - 1

	for x := float64(TableBegin); x <= TableEnd; x += presice {
		integral += float64(math.Exp(-x*x/2.0) / math.Sqrt(2*constant.PI) * presice)
		if integral >= rankBuf[Index] {
			cp.warpTable[Index] = float32(x)
			Index--
			if Index < 0 {
				break
			}
		}
	}
	return
}

// mel -> frequency
func (cp *CParam) freq(mel float32) float32 {
	return float32(700 * (math.Exp(float64(mel)/float64(1127)) - 1))
}

// frequency -> mel

func (cp *CParam) mel(freq float32) float32 {
	return float32(1127 * math.Log(1+float64(freq)/float64(700)))
}

// Hamming window
// - Arguments -
//     dVector : vector to be windowed
//        iLen : length of the vector
func (cp *CParam) doHamming(dVector []float64, iLen int) {
	var a float64
	if cp.hammingWinSize != nil && iLen != len(cp.hammingWinSize) {
		cp.hammingWinSize = nil
	}

	if cp.hammingWinSize == nil {
		cp.hammingWinSize = make([]float64, iLen, iLen)
		a = float64(2) * constant.PI / float64(iLen-1)
		for i := 0; i < iLen; i++ {
			cp.hammingWinSize[i] = 0.54 - 0.46*math.Cos(a*float64(i))
		}
	}

	for i := 0; i < iLen; i++ {
		dVector[i] *= cp.hammingWinSize[i]
	}
}

// Pre-emphasize the input signal.
// - Arguments -
//           s : pointer to input vector
//        iLen : length of the input vector
//        preE : option for the emphasis filter default = 0.97
func (cp *CParam) preEmphasise(s []float64, iLen int) {
	cp.preEmphasise2(s, iLen, 0.97)
}

func (cp *CParam) preEmphasise2(s []float64, iLen int, preE float64) {
	for i := iLen; i >= 2; i-- {
		s[i-1] -= s[i-2] * preE
	}
	s[0] *= 1.0 - preE
}

// Take delta from the relatively static signal. the deltaes
//  are appended to the static signal. Therefore the given
//  buffer must be large enough. Note that some frames will
//  be discarded in this procedure, the memory is freed at
//  the same time.
func (cp *CParam) doDelta(fdest, fsource []float32, iLen *int, width int) error {
	var winSize int = cp.mfcc.dynamicWinSize
	if *iLen < 2*winSize+1 {
		return fmt.Errorf("iLen = %d less than %d", *iLen, 2*winSize+1)
	}

	var fnorm, fsum float32
	var fpback, fpforw []float32

	if !cp.mfcc.IsPolishDiff {

		for k := 1; k <= winSize; k++ {
			fnorm += float32(k * k)
		}
	} else {
		for k := 1; k <= winSize; k++ {
			fnorm += float32(winSize - k + 1)
		}
	}

	fnorm *= 2

	for i := 0; i < *iLen; i++ {
		for d := 0; d < width; d++ {
			fsum = 0
			for k := 1; k <= winSize; k++ {
				fpback = fsource[d+__max(i-k, 0)*width:]
				fpforw = fsource[d+__min(i+k, *iLen-1)*width:]
				var im float32
				if !cp.mfcc.IsPolishDiff {
					im = float32(k)
				} else {
					im = float32(winSize-k+1) / float32(k)
				}
				fsum = fsum + im*(fpforw[0]-fpback[0])
			}
			fdest[i*width+d] = fsum / fnorm
		}
	}

	return nil
}

// Lift the cepstral to the same amplitudes. It should be
//  called just after the dct procedure and before the
//  deletion of the 0th coefficient.
// - Arguments -
//     fVector : vector to be lifted, length is specified
//               in pmfccinfo.
func (cp *CParam) liftCepstral(dVector []float64) error {
	var L float32
	if cp.mfcc == nil {
		return fmt.Errorf("CParam mfcc can not be nil")
	}
	iLen := cp.mfcc.mfccOrder + 1
	L = cp.mfcc.CepstralLifter

	if cp.cepLifterWinSize == nil {
		cp.cepLifterWinSize = make([]float32, iLen, iLen)
		for i := 0; i < iLen; i++ {
			cp.cepLifterWinSize[i] = 1.0 + L/2.0*float32(math.Sin(constant.PI*float64(i)/float64(L)))
		}
	}

	for i := 0; i < iLen; i++ {
		dVector[i] *= float64(cp.cepLifterWinSize[i])
	}

	return nil
}

func __max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func __min(x, y int) int {
	if x > y {
		return y
	}
	return x
}
