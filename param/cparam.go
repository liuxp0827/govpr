package param

import (
	"fmt"
	"govpr/cmath"
	"govpr/constant"
	"govpr/waveIO"
	"math"
)

type CParam struct {
	FBInfo           *FBankInfo
	MfccInfo         *MFCCInfo
	m_pfCepLiftWin   []float32
	m_iLenCepLiftWin int
	m_pdHammingWin   []float64 // vector of the hamming window
	m_iLenHammingWin int       // length of the hamming window

	m_nWarpWinSize   int       // warping window size
	m_pfWarpTableBuf []float32 // warping probability table
}

func NewCParam() *CParam {
	return &CParam{
		m_nWarpWinSize: 300,
	}
}

// Initialize the filter bank info struct. User should
//  call this function before calling wav2MFCC().
// - Arguments -
//   iSampRate : Sample rate (samples per second)
//    fWinTime : Frame length in ms
//      iNumFB : Number of filter banks
//     iFLoCut : Low cut-off frequency (in Hz) default = -2
//     iFHiCut : High cut-off frequency (in Hz)default = -1
// - Return -
//    true if successful, false if failed somewhere.
func (cp *CParam) InitFBank(iSampRate, fWinTime int, iNumFB int) error {
	return cp.InitFBank2(iSampRate, fWinTime, iNumFB, -2, -1)
}

func (cp *CParam) InitFBank2(iSampRate, fWinLen int, iNumFB, iFLoCut, iFHiCut int) error {
	var fmello, fmelhi, fmelstep, fCurFreq, fa, fb, fc float32

	if iFLoCut >= iFHiCut || iFHiCut > (iSampRate>>1) || iFLoCut > (iSampRate>>1) {
		return fmt.Errorf("Low and High cut off frequencies set incorrectly")
	}

	// check number of filter bancks
	if cp.MfccInfo != nil {
		if cp.MfccInfo.IOrder+1 > cp.FBInfo.INumFB {
			return fmt.Errorf("param order nb greater than filter bank nb")
		}
	}

	cp.FBInfo = new(FBankInfo)

	// given by arguments
	cp.FBInfo.ISampRate = iSampRate
	cp.FBInfo.IFrameSize = int(float32(iSampRate) * float32(fWinLen) * 1e-3)
	cp.FBInfo.INumFB = iNumFB

	// calculated from arguments
	cp.FBInfo.IfftN = 2
	for cp.FBInfo.IFrameSize > cp.FBInfo.IfftN {
		cp.FBInfo.IfftN <<= 1
	}

	Nby2 := cp.FBInfo.IfftN >> 1
	cp.FBInfo.FFRes = float32(iSampRate / cp.FBInfo.IfftN)

	// the low and high cut-off indices
	if iFLoCut < 0 {
		cp.FBInfo.Iklo = 0
		iFLoCut = 0
	} else {
		cp.FBInfo.Iklo = int(iFLoCut / int(cp.FBInfo.FFRes))
	}

	if iFHiCut < 0 {
		iFHiCut = iSampRate >> 1
		cp.FBInfo.Ikhi = Nby2
	} else {
		cp.FBInfo.Ikhi = int(iFHiCut / int(cp.FBInfo.FFRes))
		if cp.FBInfo.Ikhi > Nby2 {
			cp.FBInfo.Ikhi = Nby2
		}
	}

	cp.FBInfo.PCF = nil
	cp.FBInfo.PloChan = nil
	cp.FBInfo.PloWt = nil
	cp.FBInfo.Pdatar = nil
	cp.FBInfo.Pdatai = nil

	// the center frequencies
	cp.FBInfo.PCF = make([]float32, iNumFB+1, iNumFB+1)

	fmello = cp.mel(float32(iFLoCut))
	fmelhi = cp.mel(float32(iFHiCut))
	fmelstep = (fmelhi - fmello) / float32(iNumFB+1)
	cp.FBInfo.PCF[0] = float32(iFLoCut) // the zero index is the low cut-off
	for i := 1; i < iNumFB; i++ {
		cp.FBInfo.PCF[i] = fmello + fmelstep*float32(i)
		cp.FBInfo.PCF[i] = cp.freq(cp.FBInfo.PCF[i])
	}

	// lower channel indices
	cp.FBInfo.PloChan = make([]int16, Nby2, Nby2)
	ichan := 0

	for i := 0; i < Nby2; i++ {
		fCurFreq = float32(i) * cp.FBInfo.FFRes
		if i < cp.FBInfo.Iklo || i > cp.FBInfo.Ikhi {
			cp.FBInfo.PloChan[i] = -1
		} else {
			for cp.FBInfo.PCF[ichan] <= fCurFreq && ichan <= iNumFB {
				ichan++
			}
		}
	}

	// lower channel weights
	cp.FBInfo.PloWt = make([]float32, Nby2, Nby2)
	for i := 0; i < Nby2; i++ {
		fCurFreq = float32(i) * cp.FBInfo.FFRes
		if cp.FBInfo.PloChan[i] == -1 {
			cp.FBInfo.PloWt[i] = 0.0
		} else {
			if int(cp.FBInfo.PloChan[i]) < iNumFB {
				fc = cp.FBInfo.PCF[cp.FBInfo.PloChan[i]+1]
			} else {
				fc = float32(iFHiCut)
			}

			fa = 1 / (cp.FBInfo.PCF[cp.FBInfo.PloChan[i]] - fc)
			fb = -fa * fc
			cp.FBInfo.PloWt[i] = fa*fCurFreq + fb
		}
	}

	// alloc memory for data buffer
	cp.FBInfo.Pdatar = make([]float64, cp.FBInfo.IfftN, cp.FBInfo.IfftN)
	cp.FBInfo.Pdatai = make([]float64, cp.FBInfo.IfftN, cp.FBInfo.IfftN)

	// the defaults
	cp.FBInfo.BTakeLogs = true
	cp.FBInfo.BUsePower = false
	cp.FBInfo.BPreEmph = true
	cp.FBInfo.BUseHamming = true
	return nil
}

// Uninitialize the filter bank info. Free the memory
//  alloced in InitFBank(). Caller of InitFBank should
//  call this function when all has been done.
func (cp *CParam) UnInitFBank() {
	cp.FBInfo = nil
	cp.MfccInfo = nil
	cp.m_pdHammingWin = nil
	cp.m_iLenHammingWin = 0
	cp.m_pfCepLiftWin = nil
	cp.m_iLenCepLiftWin = 0
}

// Initialize the mfcc info struct. User should call
//  this function before calling wav2MFCC().
// - Arguments -
//      iOrder : MFCC order (except the 0th)
//  fFrameRate : MFCC frame rate in ms
// - Return -
//    true if successful, false if failed somewhere.
func (cp *CParam) InitMFCC(iOrder int, fFrmRate float32) error {

	cp.MfccInfo = new(MFCCInfo)

	if cp.FBInfo != nil {
		if iOrder+1 > cp.FBInfo.INumFB {
			return fmt.Errorf("param order nb greater than filter bank nb")
		}
	}

	cp.MfccInfo.IOrder = iOrder     // mfcc order, except the 0th
	cp.MfccInfo.FFrmRate = fFrmRate // frame rate in ms

	// the defaults
	cp.MfccInfo.IDeltaWin = 2
	cp.MfccInfo.BFBank = false
	cp.MfccInfo.B0 = false
	cp.MfccInfo.BD0 = false
	cp.MfccInfo.BA0 = false
	cp.MfccInfo.BStatic = true
	cp.MfccInfo.BDelta = true
	cp.MfccInfo.BAcce = false
	cp.MfccInfo.BCepLift = true
	cp.MfccInfo.FCepsLifter = 22

	return nil
}

// Uninitialize the MFCC info. Free the memory alloced
//  in InitMFCC(). Caller of InitFBank should call this
//  function when all has been done.
func (cp *CParam) UnInitMFCC() {
	cp.MfccInfo = nil
}

// wave -> mfcc.
// - Arguments -
//       pdata : buffer for wave data sequence
//        iLen : length of the wave sequence
//      fParam : buffer for storing the converted parameters,
//               memory is alloced within the function, so the
//               CALLER is RESPONSIBLE to free the memory.
//        iCol : width of the param vector
//        iRow : length of the param vector sequence
// - Return -
//    true if successful, false if failed somewhere.
func (cp *CParam) WAV2MFCC(pdata []float32, wavinfo waveIO.WavInfo, fParam []float32, iCol, iRow *int) error {

	if cp.MfccInfo.ZeroGlobalMean {
		cp.ZeroGlobalMean(pdata, wavinfo.Length)
	}

	if cp.MfccInfo.BdBNorm {
		cp.dBNorm(pdata, wavinfo.Length)
	}

	if cp.FBInfo == nil || cp.MfccInfo == nil {
		return fmt.Errorf("Filter bank info and MFCC info not initialized")
	}

	var iWidth, iFrameRate int
	var Nby2 int = cp.FBInfo.IfftN >> 1
	var melfloor float32 = float32(1.0)
	var fstatic []float32
	var err error

	var cmath *cmath.CMath = cmath.NewCMath()

	// calculate number of rows (frames)
	iFrameRate = int(1e-3 * float32(cp.MfccInfo.FFrmRate) * float32(cp.FBInfo.ISampRate))
	if iFrameRate > cp.FBInfo.IFrameSize {
		return fmt.Errorf("Sample point equal to zero")
	}

	*iRow = int((wavinfo.Length - int64(cp.FBInfo.IFrameSize-iFrameRate)) / int64(iFrameRate))

	// buffer for raw static params (include the 0th coef)
	iWidth = cp.MfccInfo.IOrder + 1
	fstatic = make([]float32, (*iRow)*iWidth, (*iRow)*iWidth)

	// buffer for filter banks
	var pFBank []float64 = make([]float64, cp.FBInfo.INumFB)

	for ii := 0; ii < *iRow; ii++ {

		for ij := 0; ij < cp.FBInfo.IFrameSize; ij++ {
			cp.FBInfo.Pdatar[ij] = float64(pdata[ii*iFrameRate+ij])
		}

		// Do pre-emphasis
		if cp.FBInfo.BPreEmph {
			cp.preEmphasise(cp.FBInfo.Pdatar, cp.FBInfo.IFrameSize)
		}

		// Do hamming
		if cp.FBInfo.BUseHamming {
			cp.doHamming(cp.FBInfo.Pdatar, cp.FBInfo.IFrameSize)
		}

		// take fft
		err = cmath.FFT(cp.FBInfo.Pdatar, cp.FBInfo.Pdatai, cp.FBInfo.IfftN)
		if err != nil {
			return err
		}

		for ij := 0; ij < Nby2; ij++ {
			cp.FBInfo.Pdatar[ij] =
				cp.FBInfo.Pdatar[ij]*cp.FBInfo.Pdatar[ij] +
					cp.FBInfo.Pdatai[ij]*cp.FBInfo.Pdatai[ij]
		}

		//	Differential Power Spectrum (paper:电话信道下多说话人识别研究,author:邓菁)
		if cp.FBInfo.BUsePower == true && cp.MfccInfo.BDPSCC == true {
			cp.DPSCC(Nby2)
		}

		// use power or amp
		if !cp.FBInfo.BUsePower {
			for ij := 0; ij < Nby2; ij++ {
				cp.FBInfo.Pdatar[ij] = math.Sqrt(cp.FBInfo.Pdatar[ij])
			}
		}

		// Predictive Differential Amplitude Spectrum (paper:电话信道下多说话人识别研究,author:邓菁)
		if cp.FBInfo.BUsePower == false && cp.MfccInfo.BPDASCC == true {
			cp.PDASCC(Nby2)
		}

		// accumulate filter banks
		for ij := 0; ij < Nby2; ij++ {
			if cp.FBInfo.PloChan[ij] < 0 {
				continue
			}

			// accumulate the lower bank
			if cp.FBInfo.PloChan[ij] != 0 {
				pFBank[cp.FBInfo.PloChan[ij]-1] +=
					cp.FBInfo.Pdatar[ij] * float64(cp.FBInfo.PloWt[ij])
			}
			// accumulate the upper bank
			if int(cp.FBInfo.PloChan[ij]) < cp.FBInfo.INumFB {
				pFBank[cp.FBInfo.PloChan[ij]] +=
					cp.FBInfo.Pdatar[ij] * float64(1-cp.FBInfo.PloWt[ij])
			}
		}

		// take logs
		if cp.FBInfo.BTakeLogs {
			for ij := 0; ij < cp.FBInfo.INumFB; ij++ {
				if pFBank[ij] >= float64(melfloor) {
					pFBank[ij] = math.Log(pFBank[ij])
				} else {
					pFBank[ij] = math.Log(float64(melfloor))
				}
			}
		}

		// take dct
		if !cp.MfccInfo.BFBank {
			err = cmath.DCT2(pFBank, cp.FBInfo.INumFB, iWidth)
			if err != nil {
				return err
			}

			// Liftering
			if cp.MfccInfo.BCepLift {
				err = cp.cepLift(pFBank)
				if err != nil {
					return err
				}
			}
		}

		// copy data
		for ij := 0; ij < iWidth; ij++ {
			fstatic[ii*iWidth+ij] = float32(pFBank[ij])
		}
	}

	if cp.MfccInfo.BFeatWarping {
		err = cp.FeatureWarping(fstatic, iWidth, iRow, iWidth, int(cp.MfccInfo.FeatWarpWinSize))
		if err != nil {
			return err
		}
	}

	if cp.MfccInfo.BRasta {
		err := cp.DoRasta(fstatic, iWidth, iRow, iWidth)
		if err != nil {
			return err
		}
	}

	if cp.MfccInfo.BEnergyNorm {
		cp.EnergyNorm(fstatic, iWidth, *iRow)
	}

	fParam, err = cp.Static2Full(fstatic, &iWidth, iRow)

	iCol = &iWidth

	return nil
}

// Calculate the requested parameters from the raw static coefficients.
//  This function convert raw static coef. to conjunct coef. (0th, static,
//  delta, acce or any conbinations of them) specified in m_pmfccinfo.
//  In calculating deltas, some heading and tailing frames may be discarded
//  in this function. The caller is responsible to to free the memory of
//  the returned dbuffer.
// - Arguments -
//     fStatic : buffer for raw static coef. (include the 0th)
//        iCol : width of the raw static coef., and is changed
//               by this function to be the width of the conjunct
//               params.
//        iRow : frames of the raw static coef. and is recalculated
//               in this function to be the actual number of frames
//                of the conjunct params.
//  - Return value -
//    buffer address of the conjunc parameters.
func (cp *CParam) Static2Full(fstatic []float32, iCol, iRow *int) ([]float32, error) {
	var iWidth int = *iCol
	var iSOff, iDOff, ipt int
	var fdelta, facce, fParam []float32
	var err error

	if cp.MfccInfo == nil {
		return nil, fmt.Errorf("MFCC info not initialized")
	}

	// take deltas from statics
	if cp.MfccInfo.BDelta || cp.MfccInfo.BD0 {
		fdelta = make([]float32, iWidth*(*iRow), iWidth*(*iRow))
		err = cp.doDelta(fdelta, fstatic, iRow, iWidth)
		if err != nil {
			return nil, err
		}

		iSOff = cp.MfccInfo.IDeltaWin
		iDOff = 0
	}

	// take accelerations from deltaes
	if cp.MfccInfo.BAcce || cp.MfccInfo.BA0 {
		facce = make([]float32, iWidth*(*iRow), iWidth*(*iRow))
		err = cp.doDelta(facce, fdelta, iRow, iWidth)
		if err != nil {
			return nil, err
		}

		iSOff = 2 * cp.MfccInfo.IDeltaWin
		iDOff = cp.MfccInfo.IDeltaWin
	}

	iSOff, iDOff = 0, 0

	// calculate the actual width of the conjunct parameter
	*iCol = 0
	if cp.MfccInfo.B0 {
		*iCol += 1
	}
	if cp.MfccInfo.BStatic {
		*iCol += cp.MfccInfo.IOrder
	}
	if cp.MfccInfo.BD0 {
		*iCol += 1
	}
	if cp.MfccInfo.BDelta {
		*iCol += cp.MfccInfo.IOrder
	}
	if cp.MfccInfo.BA0 {
		*iCol += 1
	}
	if cp.MfccInfo.BAcce {
		*iCol += cp.MfccInfo.IOrder
	}

	// prepare for parameter buffer
	fParam = make([]float32, (*iCol)*(*iRow), (*iCol)*(*iRow))

	for ii := 0; ii < *iRow; ii++ {
		if cp.MfccInfo.B0 {
			fParam[ipt] = fstatic[(ii+iSOff)*iWidth+0]
			ipt++
		}

		if cp.MfccInfo.BStatic {
			for ij := 1; ij < iWidth; ij++ {
				fParam[ipt] = fstatic[(ii+iSOff)*iWidth+ij]
				ipt++
			}
		}

		if cp.MfccInfo.BD0 {
			fParam[ipt] = fdelta[(ii+iDOff)*iWidth+0]
			ipt++
		}

		if cp.MfccInfo.BDelta {
			for ij := 1; ij < iWidth; ij++ {
				fParam[ipt] = fdelta[(ii+iDOff)*iWidth+ij]
				ipt++
			}
		}

		if cp.MfccInfo.BA0 {
			fParam[ipt] = facce[ii*iWidth+0]
			ipt++
		}

		if cp.MfccInfo.BAcce {
			for ij := 1; ij < iWidth; ij++ {
				fParam[ipt] = facce[ii*iWidth+ij]
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
// ---return---
//  false if there are errors when calculating the parameters, true if successful
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
/************************************************************************/
/*
  [7/10/2011 chenwl] decibel Normalization
  电平归一化：X'=[10^(dB/20)]*[2^(n-1)-1]*X/Xmax
  X   ：原始采样点量化大小
  Xmax：原始采样点量化最大值
  n   ：量化位数
  dB  ：欲归一化的分贝值
*/
/************************************************************************/
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

//------------- ZeroGlobalMean --------------------------------------------------
func (cp *CParam) ZeroGlobalMean(pdata []float32, sampleCount int64) {
	var mean float32 = 0.0
	for i := int64(0); i < sampleCount; i++ {
		mean += pdata[i]
	}
	mean /= float32(sampleCount)

	for i := int64(0); i < sampleCount; i++ {
		y := pdata[i] - mean
		if y > 32767 {
			y = 32767
		}
		if y < -32767 {
			y = -32767
		}
		if y > 0 {
			pdata[i] = float32(int16(y + 0.5))
		} else {
			pdata[i] = float32(int16(y - 0.5))
		}
	}
}

//------------- Norm static feature ---------------------------------------------
func (cp *CParam) EnergyNorm(p_FeatBuf []float32, p_nVecSize, p_nFrameNum int) {
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

	mine = (maxe - float32(cp.MfccInfo.SilFloor)*float32(math.Log(10.0))) / 10.0

	for i := 0; i < p_nFrameNum; i++ {
		if ft[index] < mine {
			mine = ft[index]
		}

		ft[index] = 1.0 - (maxe-ft[index])*float32(cp.MfccInfo.Energyscale)
		p_FeatBuf[index] = 1.0 - (maxe-p_FeatBuf[index])*float32(cp.MfccInfo.Energyscale)
		index += p_nVecSize
	}
}

//------------- Differential Power Spectrum -------------------------------------
func (cp *CParam) DPSCC(pointNB int) {
	Nby2 := pointNB
	for j := 0; j < Nby2; j++ {
		if j < Nby2-1 {
			cp.FBInfo.Pdatar[j] = math.Abs(cp.FBInfo.Pdatar[j] - cp.FBInfo.Pdatar[j+1])
		} else {
			cp.FBInfo.Pdatar[j] = 0
		}
	}
}

//------------- Predictive Differential Amplitude Spectrum ----------------------
func (cp *CParam) PDASCC(pointNB int) {
	Nby2 := pointNB

	//	1.预测差分
	var WINLEN int = 6
	var damplitude []float64 = make([]float64, Nby2, Nby2)
	for j := 0; j < Nby2; j++ {
		dmax := -math.MaxFloat64
		for w := 0; j+w < Nby2 && w <= WINLEN; w++ {
			dsin := math.Sin((float64(w) * constant.PI) / float64(2*WINLEN))
			dcur := cp.FBInfo.Pdatar[j+w] * dsin
			if dcur > dmax {
				dmax = dcur
			}
		}
		damplitude[j] = dmax
	}

	var alpha float64 = 1.05
	var dDright []float64 = make([]float64, Nby2, Nby2)
	var dDleft []float64 = make([]float64, Nby2, Nby2)

	for j := 0; j < Nby2-1; j++ {
		if damplitude[j] > cp.FBInfo.Pdatar[j] && damplitude[j+1] < cp.FBInfo.Pdatar[j+1] {
			dDright[j] = cp.FBInfo.Pdatar[j] - alpha*cp.FBInfo.Pdatar[j+1]
		} else if damplitude[j] <= cp.FBInfo.Pdatar[j] && damplitude[j+1] >= cp.FBInfo.Pdatar[j+1] {
			dDright[j] = alpha*cp.FBInfo.Pdatar[j] - cp.FBInfo.Pdatar[j+1]
		} else {
			dDright[j] = cp.FBInfo.Pdatar[j] - cp.FBInfo.Pdatar[j+1]
		}
	}

	//	Dright的最右一个如何算？
	dDright[Nby2-1] = 0.0

	for j := Nby2 - 1; j > 0; j-- {
		if damplitude[j] < cp.FBInfo.Pdatar[j] && damplitude[j-1] < cp.FBInfo.Pdatar[j-1] {
			dDleft[j] = cp.FBInfo.Pdatar[j] - alpha*cp.FBInfo.Pdatar[j-1]
		} else if damplitude[j] >= cp.FBInfo.Pdatar[j] && damplitude[j-1] >= cp.FBInfo.Pdatar[j-1] {
			dDleft[j] = alpha*cp.FBInfo.Pdatar[j] - cp.FBInfo.Pdatar[j-1]
		} else {
			dDleft[j] = cp.FBInfo.Pdatar[j] - cp.FBInfo.Pdatar[j-1]
		}
	}
	//	Dleft的最左一个如何算？
	if damplitude[0] < cp.FBInfo.Pdatar[0] {
		dDleft[0] = (1.0 - alpha) * cp.FBInfo.Pdatar[0]
	} else {
		dDleft[0] = (alpha - 1.0) * cp.FBInfo.Pdatar[0]
	}

	// 2.累积过程

	var left []float64 = make([]float64, Nby2, Nby2)
	var right []float64 = make([]float64, Nby2, Nby2)
	for i := 1; i < Nby2; i++ {
		left[i] = left[i-1] + dDleft[i-1]
	}

	for i := Nby2 - 2; i >= 0; i-- {
		right[i] = right[i+1] + dDright[i+1]
	}

	for i := 0; i < Nby2; i++ {
		cp.FBInfo.Pdatar[i] = (left[i] + right[i]) / 2.0
	}
}

//------------- Feature Warping -------------------------------------------------
// m_nWinSize=300
// vSize=nStep=特征维数
// nInNum 输入特征帧数
// pdata 输入特征
// nOutNum 输出特征帧数
func (cp *CParam) FeatureWarping(pdata []float32, vSize int, nInNum *int, nStep, m_nWinSize int) error {

	cp.createWarpTable()
	var nOutNum int = *nInNum - m_nWinSize
	if nOutNum <= 0 {
		return fmt.Errorf("nOutNum can not <= 0")
	}

	var warpBuf []float32 = make([]float32, nOutNum*nStep, nOutNum*nStep)
	var warpFrmNo int = 0
	var pDataIvt []float32 = make([]float32, nOutNum*nStep, nOutNum*nStep)

	for i := 0; i < nStep; i++ {
		//		var dst []float32 = make([]float32, nOutNum*nStep+i*nInNum, nOutNum*nStep+i*nInNum)
		for j := 0; j < *nInNum; j++ {
			//			dst[j] = pdata[j*nStep+i]
			if j < nOutNum*nStep {
				pDataIvt[j] = pdata[j*nStep+i]
			}
		}
	}

	var halfwin int = m_nWinSize >> 1
	var minus_res []float32 = make([]float32, m_nWinSize, m_nWinSize)

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

			var ui []uint = make([]uint, m_nWinSize, m_nWinSize)
			for i := 0; i < m_nWinSize; i++ {
				ui[i] = uint(minus_res[i])
			}

			for m := 0; m < 2*halfwin-1; m++ {
				nIndex -= int((ui[m] >> 31))
			}

			warpBuf[warpFrmNo*nStep+k] = cp.m_pfWarpTableBuf[nIndex]
		}

		for k := vSize; k < nStep; k++ {
			warpBuf[warpFrmNo*nStep+k] = pdata[i*nStep+k]
		}
		warpFrmNo++
	}

	*nInNum = warpFrmNo

	copy(pdata[:warpFrmNo*nStep], warpBuf[:warpFrmNo*nStep])

	return nil
}

//------------- Rasta filtering -------------------------------------------------
/************************************************************************/
/*
	data : static mfcc
	vSize: order of mfcc
	nNum : frame number
	nStep: order of mfcc
*/
/************************************************************************/
func (cp *CParam) DoRasta(data []float32, vSize int, nNum *int, nStep int) error {
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
					float32(cp.MfccInfo.RastaCoff)*RastaBuf[(i-1)*vSize+j]
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
	cp.m_pfWarpTableBuf = make([]float32, cp.m_nWarpWinSize, cp.m_nWarpWinSize)
	var rankBuf []float64 = make([]float64, cp.m_nWarpWinSize, cp.m_nWarpWinSize)

	for i := 0; i < cp.m_nWarpWinSize; i++ {
		rankBuf[i] = float64(float64(cp.m_nWarpWinSize)-0.5-float64(i)) / float64(cp.m_nWarpWinSize)
	}

	var integral float64 = 0.0
	var Index int = cp.m_nWarpWinSize - 1

	for x := float64(TableBegin); x <= TableEnd; x += presice {
		integral += float64(math.Exp(-x*x/2.0) / math.Sqrt(2*constant.PI) * presice)
		if integral >= rankBuf[Index] {
			cp.m_pfWarpTableBuf[Index] = float32(x)
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
	return float32(700*math.Exp(float64(mel)/float64(1127)) - 1)
}

// frequency -> mel

func (cp *CParam) mel(freq float32) float32 {
	return float32(1127 * math.Log(float64(1+freq)/float64(700)))
}

// Hamming window
// - Arguments -
//     dVector : vector to be windowed
//        iLen : length of the vector
// - Return -
//   false if there's memory error, true if successful.
func (cp *CParam) doHamming(dVector []float64, iLen int) {
	var a float64
	if cp.m_pdHammingWin != nil && iLen != cp.m_iLenHammingWin {
		cp.m_pdHammingWin = nil
	}

	if cp.m_pdHammingWin == nil {
		cp.m_pdHammingWin = make([]float64, iLen, iLen)
		a = float64(2) * constant.PI / float64(iLen-1)
		for i := 0; i < iLen; i++ {
			cp.m_pdHammingWin[i] = 0.54 - 0.46*math.Cos(a*float64(i))
		}
		cp.m_iLenHammingWin = iLen
	}

	for i := 0; i < iLen; i++ {
		dVector[i] *= cp.m_pdHammingWin[i]
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
func (cp *CParam) doDelta(fdest, fsource []float32, iLen *int, iWidth int) error {
	var winSize int = cp.MfccInfo.IDeltaWin
	if *iLen < 2*winSize+1 {
		return fmt.Errorf("iLen = %d less than %d", *iLen, 2*winSize+1)
	}

	var fnorm, fsum float32
	var fpback, fpforw []float32

	if !cp.MfccInfo.BPolishDiff {

		for k := 1; k < winSize; k++ {
			fnorm += float32(k * k)
		}
	} else {
		for k := 1; k < winSize; k++ {
			fnorm += float32(winSize - k + 1)
		}
	}

	fnorm *= 2

	for i := 0; i < *iLen; i++ {
		for d := 0; d < iWidth; d++ {
			fsum = 0
			for k := 1; k <= winSize; k++ {
				fpback = fsource[d+__max(i-k, 0)*iWidth:]
				fpforw = fsource[d+__min(i+k, *iLen-1)*iWidth:]
				var im float32
				if !cp.MfccInfo.BPolishDiff {
					im = float32(k)
				} else {
					im = float32(winSize-k+1) / float32(k)
				}
				fsum = fsum + im*(fpforw[0]-fpback[0])
			}
			fdest[i*iWidth+d] = fsum / fnorm
		}
	}

	return nil
}

// Lift the cepstral to the same amplitudes. It should be
//  called just after the dct procedure and before the
//  deletion of the 0th coefficient.
// - Arguments -
//     fVector : vector to be lifted, length is specified
//               in m_pmfccinfo.
// - Return -
//   false if there's memory error, true if successful.
func (cp *CParam) cepLift(dVector []float64) error {
	var L float32
	if cp.MfccInfo == nil {
		return fmt.Errorf("CParam MfccInfo can not be nil")
	}
	iLen := cp.MfccInfo.IOrder + 1
	L = cp.MfccInfo.FCepsLifter

	if cp.m_pfCepLiftWin == nil {
		cp.m_pfCepLiftWin = make([]float32, iLen, iLen)
		for i := 0; i < iLen; i++ {
			cp.m_pfCepLiftWin[i] = 1.0 + L/2.0*float32(math.Sin(constant.PI*float64(i)/float64(L)))
		}
	}

	for i := 0; i < iLen; i++ {
		dVector[i] *= float64(cp.m_pfCepLiftWin[i])
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
