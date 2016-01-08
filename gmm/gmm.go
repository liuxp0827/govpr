package gmm

import (
	"fmt"
	"govpr/constant"
	"govpr/fileIO"
	"math"
	"os"
	"sort"
	"zeus/log"
)

const (
	MEAN_ONLY = iota
	KLDIVERGENCE
)

type NormParam struct {
	mean   float64 // = 0.0
	stdVar float64 // = 1.0
}

type ZNormParam struct {
	mic NormParam
	tel NormParam
}

type GMM struct {
	BgsvLoaded    bool
	DGsv          []float64
	BCout         bool
	BReadSeparate bool // format of reading features: separate or whole

	BTopLoaded    bool    // whether top distribution parameter is loaded
	ITopDistribNB int     // top distribution number
	TopList       [][]int // top distribution for all frames

	BMLoaded bool // whether the model is loaded
	BPLoaded bool // whether the param are loaded

	IFrames int         // number of total frames
	FParam  [][]float32 // feature buffer

	Param_list_file string

	BDiag          bool        // if the covariance matrix diagonal
	IVectorSize    int         // Vector size
	INumMixtures   int         // Mixtures of the GMM
	DLDet          []float64   // determinant of the covariance matrix [mixture]
	DMixtureWeight []float64   // weight of each mixture[mixture]
	DMean          [][]float64 // mean vector [mixture,dimension]
	DCovar         [][]float64 // covariance (diagonal) [mixture,dimension]
}

func NewGMM() *GMM {
	gmm := &GMM{
		BgsvLoaded:    false,
		DGsv:          make([]float64, 0),
		BCout:         true,
		BReadSeparate: false,

		BTopLoaded:    false,
		ITopDistribNB: 5,

		BMLoaded: false,
		BPLoaded: false,

		FParam: make([][]float32, 0),

		BDiag:          true,
		DMixtureWeight: make([]float64, 0),
		DMean:          make([][]float64, 0),
		DCovar:         make([][]float64, 0),
	}
	return gmm
}

func (g *GMM) Close() {
	g.CleanUpTop()
	g.CleanUpMdl()
	g.CleanUpPar()
	g.CleanUpGSV()
}

/* Model file access routines */
//func (g *GMM) ProtoModel(iDim, iMixs int) int {
//	return 0
//}

func (g *GMM) DupModel(gmm *GMM) {
	if !g.BMLoaded {
		g.INumMixtures = gmm.INumMixtures
		g.IVectorSize = gmm.IVectorSize
		g.DLDet = make([]float64, g.INumMixtures, g.INumMixtures)
		g.DMixtureWeight = make([]float64, g.INumMixtures, g.INumMixtures)
		g.DMean = make([][]float64, g.INumMixtures, g.INumMixtures)
		g.DCovar = make([][]float64, g.INumMixtures, g.INumMixtures)

		for i := 0; i < g.INumMixtures; i++ {
			g.DLDet[i] = gmm.DLDet[i]
			g.DMixtureWeight[i] = gmm.DMixtureWeight[i]
			g.DMean[i] = make([]float64, g.IVectorSize, g.IVectorSize)
			g.DCovar[i] = make([]float64, g.IVectorSize, g.IVectorSize)
			for j := 0; j < g.IVectorSize; j++ {
				g.DMean[i][j] = gmm.DMean[i][j]
				g.DCovar[i][j] = gmm.DCovar[i][j]
			}
		}
		g.BMLoaded = true
	} else {

		for i := 0; i < g.INumMixtures; i++ {
			g.DLDet[i] = gmm.DLDet[i]
			g.DMixtureWeight[i] = gmm.DMixtureWeight[i]

			for j := 0; j < g.IVectorSize; j++ {
				g.DMean[i][j] = gmm.DMean[i][j]
				g.DCovar[i][j] = gmm.DCovar[i][j]
			}
		}
	}
}

func (g *GMM) LoadModel(filename string) error {

	reader, err := fileIO.NewFileIO(filename, 0)
	if err != nil {
		return err
	}

	g.INumMixtures, err = reader.ReadInt()
	if err != nil {
		return err
	}

	g.IVectorSize, err = reader.ReadInt()
	if err != nil {
		return err
	}

	if !g.BMLoaded {
		g.DLDet = make([]float64, g.INumMixtures, g.INumMixtures)
		g.DMixtureWeight = make([]float64, g.INumMixtures, g.INumMixtures)
		g.DMean = make([][]float64, g.INumMixtures, g.INumMixtures)
		g.DCovar = make([][]float64, g.INumMixtures, g.INumMixtures)
		for i := 0; i < g.INumMixtures; i++ {
			g.DMean[i] = make([]float64, g.IVectorSize, g.IVectorSize)
			g.DCovar[i] = make([]float64, g.IVectorSize, g.IVectorSize)
		}
	}

	for i := 0; i < g.INumMixtures; i++ {
		g.DLDet[i] = 0.0
	}

	for i := 0; i < g.INumMixtures; i++ {
		g.DMixtureWeight[i], err = reader.ReadDouble()
		if err != nil {
			return err
		}
	}

	for i := 0; i < g.INumMixtures; i++ {
		_, err = reader.ReadDouble() // not used
		if err != nil {
			return err
		}

		_, err = reader.ReadDouble() // not used
		if err != nil {
			return err
		}

		_, err = reader.ReadChar() // not used
		if err != nil {
			return err
		}

		for j := 0; j < g.IVectorSize; j++ {
			g.DCovar[i][j], err = reader.ReadDouble()
			if err != nil {
				return err
			}

			g.DLDet[i] = math.Log(g.DCovar[i][j])
		}

		for j := 0; j < g.IVectorSize; j++ {
			g.DMean[i][j], err = reader.ReadDouble()
			if err != nil {
				return err
			}
		}
	}

	reader.Close()
	g.BMLoaded = true
	return nil

}

func (g *GMM) LoadModelByZNorm(filename string, zn *ZNormParam) error {
	reader, err := fileIO.NewFileIO(filename, 0)
	if err != nil {
		return err
	}

	g.INumMixtures, err = reader.ReadInt()
	if err != nil {
		return err
	}

	g.IVectorSize, err = reader.ReadInt()
	if err != nil {
		return err
	}

	if !g.BMLoaded {
		g.DLDet = make([]float64, g.INumMixtures, g.INumMixtures)
		g.DMixtureWeight = make([]float64, g.INumMixtures, g.INumMixtures)
		g.DMean = make([][]float64, g.INumMixtures, g.INumMixtures)
		g.DCovar = make([][]float64, g.INumMixtures, g.INumMixtures)
		for i := 0; i < g.INumMixtures; i++ {
			g.DMean[i] = make([]float64, g.IVectorSize, g.IVectorSize)
			g.DCovar[i] = make([]float64, g.IVectorSize, g.IVectorSize)
		}
	}

	for i := 0; i < g.INumMixtures; i++ {
		g.DLDet[i] = 0.0
	}

	for i := 0; i < g.INumMixtures; i++ {
		g.DMixtureWeight[i], err = reader.ReadDouble()
		if err != nil {
			return err
		}
	}

	for i := 0; i < g.INumMixtures; i++ {
		_, err = reader.ReadDouble() // not used
		if err != nil {
			return err
		}

		_, err = reader.ReadDouble() // not used
		if err != nil {
			return err
		}

		_, err = reader.ReadChar() // not used
		if err != nil {
			return err
		}

		for j := 0; j < g.IVectorSize; j++ {
			g.DCovar[i][j], err = reader.ReadDouble()
			if err != nil {
				return err
			}

			g.DLDet[i] = math.Log(g.DCovar[i][j])
		}

		for j := 0; j < g.IVectorSize; j++ {
			g.DMean[i][j], err = reader.ReadDouble()
			if err != nil {
				return err
			}
		}
	}

	zn = new(ZNormParam)
	zn.mic.mean, err = reader.ReadDouble()
	if err != nil {
		return nil
	}
	zn.mic.stdVar, err = reader.ReadDouble()
	if err != nil {
		return nil
	}
	zn.tel.mean, err = reader.ReadDouble()
	if err != nil {
		return nil
	}
	zn.tel.stdVar, err = reader.ReadDouble()
	if err != nil {
		return nil
	}

	reader.Close()
	g.BMLoaded = true
	return nil
}

func (g *GMM) SaveModel(filename string) error {
	writer, err := fileIO.NewFileIO(filename, 1)
	if err != nil {
		return err
	}

	err = writer.WriteInt(g.INumMixtures)
	if err != nil {
		return err
	}

	writer.WriteInt(g.IVectorSize)
	if err != nil {
		return err
	}

	for i := 0; i <= g.INumMixtures; i++ {
		err = writer.WriteDouble(g.DMixtureWeight[i])
		if err != nil {
			return err
		}
	}

	for i := 0; i < g.INumMixtures; i++ {
		err = writer.WriteDouble(0.0) // not used
		if err != nil {
			return err
		}

		err = writer.WriteDouble(0.0) // not used
		if err != nil {
			return err
		}

		err = writer.WriteChar(byte(0)) // not used
		if err != nil {
			return err
		}

		for j := 0; j < g.IVectorSize; j++ {
			err = writer.WriteDouble(g.DCovar[i][j])
			if err != nil {
				return err
			}
		}

		for j := 0; j < g.IVectorSize; j++ {
			err = writer.WriteDouble(g.DMean[i][j])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *GMM) SaveModelByZNorm(filename string, zn *ZNormParam) error {

	writer, err := fileIO.NewFileIO(filename, 1)
	if err != nil {
		return err
	}

	err = writer.WriteInt(g.INumMixtures)
	if err != nil {
		return err
	}

	writer.WriteInt(g.IVectorSize)
	if err != nil {
		return err
	}

	for i := 0; i <= g.INumMixtures; i++ {
		err = writer.WriteDouble(g.DMixtureWeight[i])
		if err != nil {
			return err
		}
	}

	for i := 0; i < g.INumMixtures; i++ {
		err = writer.WriteDouble(0.0) // not used
		if err != nil {
			return err
		}

		err = writer.WriteDouble(0.0) // not used
		if err != nil {
			return err
		}

		err = writer.WriteChar(byte(0)) // not used
		if err != nil {
			return err
		}

		for j := 0; j < g.IVectorSize; j++ {
			err = writer.WriteDouble(g.DCovar[i][j])
			if err != nil {
				return err
			}
		}

		for j := 0; j < g.IVectorSize; j++ {
			err = writer.WriteDouble(g.DMean[i][j])
			if err != nil {
				return err
			}
		}
	}

	err = writer.WriteDouble(zn.mic.mean)
	if err != nil {
		return err
	}

	err = writer.WriteDouble(zn.mic.stdVar)
	if err != nil {
		return err
	}

	err = writer.WriteDouble(zn.tel.mean)
	if err != nil {
		return err
	}

	err = writer.WriteDouble(zn.tel.stdVar)
	if err != nil {
		return err
	}

	return nil
}

func (g *GMM) CleanUpMdl() {
	if !g.BMLoaded {
		return
	}

	/* 貌似不需要 */
	for ii := 0; ii < g.INumMixtures; ii++ {
		g.DMean[ii] = nil
		g.DCovar[ii] = nil
	}

	g.DMean = nil
	g.DCovar = nil
	g.DLDet = nil
	g.DMixtureWeight = nil
	g.BMLoaded = false
}

func (g *GMM) CleanUpPar() {
	if !g.BPLoaded {
		return
	}

	/* 貌似不需要 */
	for i := 0; i < g.IFrames; i++ {
		g.FParam[i] = nil
	}

	g.FParam = nil
	g.BPLoaded = false
}

func (g *GMM) CleanUpTop() {
	if !g.BTopLoaded {
		g.TopList = nil
	}

	for i := 0; i < g.IFrames; i++ {
		g.TopList[i] = nil
	}
	g.TopList = nil
	g.BTopLoaded = false
}

func (g *GMM) CopyFParam(gmm *GMM) error {
	if !gmm.BMLoaded || !gmm.BPLoaded {
		return fmt.Errorf("source GMM has not loaded model or feature parameter")
	}

	if g.BPLoaded {
		g.CleanUpPar()
	}

	g.BPLoaded = true
	g.IFrames = gmm.IFrames
	g.IVectorSize = gmm.IVectorSize
	g.FParam = make([][]float32, gmm.IFrames, gmm.IFrames)
	for i := 0; i < g.IFrames; i++ {
		g.FParam[i] = make([]float32, g.IVectorSize, g.IVectorSize)
	}

	for i := 0; i < g.IFrames; i++ {
		for j := 0; j < g.IVectorSize; j++ {
			g.FParam[i][j] = gmm.FParam[i][j]
		}
	}

	return nil
}

/* Model file access routines */

//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////

/* Model estimation routines */

func (g *GMM) EM(mixtures int) (int, error) {
	var dlogfrmprob, rubbish, lastrubbish float64
	var dsumgama, dlogmixw, dgama, dmixw []float64
	var threshold float64 = 1e-5
	var DMean, dvar [][]float64
	var iloop int = 0

	DMean = make([][]float64, mixtures, mixtures)
	dvar = make([][]float64, mixtures, mixtures)
	for i := 0; i < mixtures; i++ {
		DMean[i] = make([]float64, g.IVectorSize, g.IVectorSize)
		dvar[i] = make([]float64, g.IVectorSize, g.IVectorSize)
	}

	dmixw = make([]float64, mixtures, mixtures)
	dlogmixw = make([]float64, mixtures, mixtures)
	dgama = make([]float64, mixtures, mixtures)
	dsumgama = make([]float64, mixtures, mixtures)

	rubbish = .0

	doing := func() (int, error) {
		lastrubbish = rubbish
		rubbish = .0
		for i := 0; i < mixtures; i++ {

			// speed up
			if g.DMixtureWeight[i] <= 0 {
				dlogmixw[i] = constant.LOGZERO
			} else {
				dlogmixw[i] = math.Log(g.DMixtureWeight[i])
			}

			// clean up temporary values
			dmixw[i] = .0
			dsumgama[i] = .0
			for j := 0; j < g.IVectorSize; j++ {
				DMean[i][j] = .0
				dvar[i][j] = .0
			}
		}

		if g.BReadSeparate {

			g.IFrames = 0
			EMFileIO, err := fileIO.NewFileIO(g.Param_list_file, 0)
			if err != nil {
				return 0, err
			}

		FOR:
			_, err = EMFileIO.ReadLine()
			for err == nil {
				var frames int
				var mfccBuffer [][]float32
				g.LoadParamFileBuf(EMFileIO.Line, &mfccBuffer, &frames)

				for ii := 0; ii < frames; ii++ {
					dlogfrmprob = constant.LOGZERO
					for ij := 0; ij < mixtures; ij++ {
						dgama[ij] = g.LMixProb(mfccBuffer[ii], ij)
						dgama[ij] += dlogmixw[ij]
						dlogfrmprob = g.LogAdd(dgama[ij], dlogfrmprob)
					}
					rubbish += dlogfrmprob
					for ij := 0; ij < mixtures; ij++ {
						dgama[ij] -= dlogfrmprob
						dgama[ij] = math.Exp(dgama[ij])
						dsumgama[ij] += dgama[ij]

						dmixw[ij] += dgama[ij]
						for ik := 0; ik < g.IVectorSize; ik++ {
							DMean[ij][ik] += dgama[ij] * float64(mfccBuffer[ii][ik])
							dvar[ij][ik] += dgama[ij] * float64(mfccBuffer[ii][ik]) * float64(mfccBuffer[ii][ik])
						}
					}
				}
				g.IFrames += frames
				goto FOR
			}

		} else {
			for i := 0; i < g.IFrames; i++ {
				dlogfrmprob = constant.LOGZERO
				for j := 0; j < mixtures; j++ {
					dgama[j] = g.LMixProb(g.FParam[i], j)
					dgama[j] += dlogmixw[j]
					dlogfrmprob = g.LogAdd(dgama[j], dlogfrmprob)
				}

				rubbish += dlogfrmprob

				for j := 0; j < constant.TOP_MIXS; j++ {
					dgama[j] -= dlogfrmprob
					dgama[j] = math.Exp(dgama[j])
					dsumgama[j] += dgama[j]
					// update weights
					dmixw[j] += dgama[j]
					for k := 0; k < g.IVectorSize; k++ {
						DMean[j][k] += dgama[j] * float64(g.FParam[i][k])
						dvar[j][k] += dgama[j] * float64(g.FParam[i][k]) * float64(g.FParam[i][k])
					}
				}
			}
		}

		rubbish /= float64(g.IFrames) // rubbish = LLR

		//-----------------------------------------------
		// M-step
		//-----------------------------------------------
		// update weight

		for i := 0; i < mixtures; i++ {
			if dsumgama[i] == .0 {
				return -1, nil
			}

			g.DMixtureWeight[i] = dmixw[i] / float64(g.IFrames)

			for j := 0; j < g.IVectorSize; j++ {
				g.DMean[i][j] = DMean[i][j] / dsumgama[i]
				g.DCovar[i][j] = dvar[i][j] / dsumgama[i]
				g.DCovar[i][j] -= g.DMean[i][j] * g.DMean[i][j]

				if g.DCovar[i][j] < constant.VAR_FLOOR {
					g.DCovar[i][j] = constant.VAR_FLOOR
				}

				if g.DCovar[i][j] < constant.VAR_CEILING {
					g.DCovar[i][j] = constant.VAR_CEILING
				}
			}
		}

		for i := 0; i < mixtures; i++ {
			g.DLDet[i] = .0
			for j := 0; j < g.IVectorSize; j++ {
				g.DLDet[i] += math.Log(g.DCovar[i][j])
			}
		}

		iloop++
		if g.BCout {
			log.Infof("loop: %d, Average Log Likelihood: %f, Increment: %f", iloop, rubbish, rubbish-lastrubbish)
		}
		return 0, nil
	}

DO:
	ret, err := doing()
	if err != nil {
		return 0, err
	} else if ret == -1 {
		return -1, nil
	}

	for iloop < constant.MAX_LOOP && math.Abs((rubbish-lastrubbish)/(lastrubbish+0.01)) > threshold {
		goto DO
	}

	if g.BCout {
		if iloop >= constant.MAX_LOOP {
			log.Info("Break at loop %d", iloop)
		} else {
			log.Info("Converged at loop %d", iloop)
		}
	}

	return iloop, nil
}

//func (g *GMM) MaPFromUBM(mixtures int) int {
//	var dlogfrmprob, rubbish, lastrubbish float64
//	var dsumgama, dlogmixw, dgama []float64
//	var threshold float64 = 1e-5
//	var DMean [][]float64
//	var iloop int = 0
//
//	DMean = make([][]float64, mixtures, mixtures)
//	for i := 0; i < mixtures; i++ {
//		DMean[i] = make([]float64, g.IVectorSize, g.IVectorSize)
//	}
//
//	dlogmixw = make([]float64, mixtures, mixtures)
//	dgama = make([]float64, mixtures, mixtures)
//	dsumgama = make([]float64, mixtures, mixtures)
//
//	g.TopDistribs(g.FParam, constant.TOP_MIXS)
//
//	rubbish = .0
//
//DO:
//	func() {
//		lastrubbish = rubbish
//		for i := 0; i < mixtures; i++ {
//			// speed up
//			if g.DMixtureWeight[i] <= 0 {
//				dlogmixw[i] = constant.LOGZERO
//			} else {
//				dlogmixw[i] = math.Log(g.DMixtureWeight[i])
//			}
//
//			// clean up temporary values
//			dsumgama[i] = .0
//			for j := 0; j < g.IVectorSize; j++ {
//				DMean[i][j] = .0
//			}
//		}
//
//		for i := 0; i < g.IFrames; i++ {
//			dlogfrmprob = constant.LOGZERO
//			for j := 0; j < constant.TOP_MIXS; j++ {
//				dgama[j] = g.LMixProb(g.FParam[i], g.TopList[i][j])
//				dgama[j] += dlogmixw[g.TopList[i][j]]
//				dlogfrmprob = g.LogAdd(dgama[j], dlogfrmprob)
//			}
//
//			rubbish += dlogfrmprob
//			for j := 0; j < constant.TOP_MIXS; j++ {
//				dgama[j] -= dlogfrmprob
//				dgama[j] = math.Exp(dgama[j])
//				dsumgama[g.TopList[i][j]] += dgama[j]
//				for k := 0; k < g.IVectorSize; k++ {
//					DMean[g.TopList[i][j]][k] += dgama[j] * float64(g.FParam[i][k])
//				}
//			}
//		}
//
//		rubbish /= float64(g.IFrames) // rubbish = LLR
//
//		//-----------------------------------------------
//		// M-step
//		//-----------------------------------------------
//		// update weight
//
//		for i := 0; i < mixtures; i++ {
//			if dsumgama[i] == .0 {
//				continue
//			}
//
//			for j := 0; j < g.IVectorSize; j++ {
//				g.DMean[i][j] = DMean[i][j] / dsumgama[i]
//			}
//		}
//
//		iloop++
//		if g.BCout {
//			log.Infof("loop: %d, Average Log Likelihood: %f, Increment: %f", iloop, rubbish, rubbish-lastrubbish)
//		}
//	}()
//
//	for iloop < constant.MAX_LOOP && math.Abs((rubbish-lastrubbish)/(lastrubbish+0.01)) > threshold {
//		goto DO
//	}
//
//	if g.BCout {
//		if iloop >= constant.MAX_LOOP {
//			log.Info("Break at loop %d", iloop)
//		} else {
//			log.Info("Converged at loop %d", iloop)
//		}
//	}
//
//	return iloop
//}

//func (g *GMM) updateDet() int {
//	return 0
//}
//
//func (g *GMM) vqCheckZero(counter *int) bool {
//	return false
//}
//
//func (g *GMM) vqNearest(vector, distance *float64) int {
//	return 0
//}

// Init the mean vectors of VQ procedure
// Return Value:
//   1 :  Success
//   0 :  Nothing to init
//func (g *GMM) vqInit(temperature []float64) error {
//	if !g.BMLoaded {
//		return fmt.Errorf("model not loaded")
//	}
//
//	var iGlobalFrames int = 0
//	var amp float64 = 0.001
//	dGlobalMean := make([]float64, g.IVectorSize, g.IVectorSize)
//	iCount := make([]int, g.INumMixtures, g.INumMixtures)
//
//	if g.BReadSeparate {
//		g.IFrames = 0
//
//	}
//
//	return nil
//}

// Euclidean Distance
// (alternative: pre-calculate the total variance of each feature,
// then weighting each feature with the variance)
//func (g *GMM) vqDistortion(vec1, vec2 []float64) float64 {
//	var result float64
//	for i := 0; i < g.IVectorSize; i++ {
//		result += (vec1[i] - vec2[i]) * (vec1[i] - vec2[i])
//	}
//	return result
//}

/* Model estimation routines */

//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////

/* Observation file routines */

func (g *GMM) LoadParamFileBuf(filename string, databuf *[][]float32, sample *int) (int, error) {

	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}

	fheader, err := readHTKHeader(file)
	if err != nil {
		return 0, err
	}

	*sample = int(fheader.nSamples)
	var vectsize = fheader.sampSize / 4

	fileio, err := fileIO.NewFileIO2(file, 0)
	if err != nil {
		return 0, err
	}

	*databuf = make([][]float32, *sample, *sample)
	for i := 0; i < *sample; i++ {
		(*databuf)[i] = make([]float32, vectsize, vectsize)
		for j := 0; j < int(vectsize); j++ {
			(*databuf)[i][j], err = fileio.ReadFloat32()
			if err != nil {
				return 0, err
			}
		}
	}

	fileio.Close()

	return int(vectsize), nil
}

/* Observation file routines */

//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////

/* Likelihood calculation routines */

func (g *GMM) LProb(featureBuf [][]float32, start, length int64) float64 {
	var dgama, dlogfrmprob, sum float64 = .0, .0, .0
	dlogmixw := make([]float64, g.INumMixtures, g.INumMixtures)
	for ij := 0; ij < g.INumMixtures; ij++ {
		if g.DMixtureWeight[ij] <= 0 {
			dlogmixw[ij] = constant.LOGZERO
		} else {
			dlogmixw[ij] = math.Log(g.DMixtureWeight[ij])
		}
	}

	for ii := int64(start); ii < (start + length); ii++ {
		dlogfrmprob = constant.LOGZERO
		for jj := 0; jj < g.INumMixtures; jj++ {
			dgama = g.LMixProb(featureBuf[ii], g.TopList[ii][jj]) + dlogmixw[g.TopList[ii][jj]]
			dlogfrmprob = g.LogAdd(dgama, dlogfrmprob)
		}
		sum += dlogfrmprob
	}

	dlogmixw = nil
	return sum
}

func (g *GMM) LTopProb(featureBuf [][]float32, start, length int64) float64 {
	var dgama, dlogfrmprob, sum float64 = .0, .0, .0
	dlogmixw := make([]float64, g.INumMixtures, g.INumMixtures)

	for ij := 0; ij < g.INumMixtures; ij++ {
		if g.DMixtureWeight[ij] <= 0 {
			dlogmixw[ij] = constant.LOGZERO
		} else {
			dlogmixw[ij] = math.Log(g.DMixtureWeight[ij])
		}
	}

	for ii := int64(start); ii < (start + length); ii++ {
		dlogfrmprob = constant.LOGZERO
		for jj := 0; jj < g.ITopDistribNB; jj++ {
			dgama = g.LMixProb(featureBuf[ii], g.TopList[ii][jj]) + dlogmixw[g.TopList[ii][jj]]
			dlogfrmprob = g.LogAdd(dgama, dlogfrmprob)
		}
		sum += dlogfrmprob
	}

	dlogmixw = nil
	return sum
}

// Calculate the output probility of the given vector according
//   to the Triggers (Enable or disable particular dimensions)
func (g *GMM) LVectorProb(feature_Vector []float32) float64 {

	if !g.BMLoaded {
		panic("Model not loaded")
	}
	var dresult, dsum float64
	dresult = constant.LOGZERO
	for ii := 0; ii < g.INumMixtures; ii++ {
		dsum = g.LMixProb(feature_Vector, ii) + math.Log(g.DMixtureWeight[ii])
		dresult = g.LogAdd(dresult, dsum)
	}

	return dresult
}

func (g *GMM) LVectorProbTri(feature_Vector []float32, trigger_Vector []int) float64 {
	if !g.BMLoaded {
		panic("Model not loaded")
	}

	var isum int
	var dresult, dsum, ddet, dtmp float64
	dresult = constant.LOGZERO
	for ii := 0; ii < g.INumMixtures; ii++ {
		isum = 0
		dsum = 0.0
		ddet = 1.0
		for ij := 0; ij < g.IVectorSize; ij++ {
			if trigger_Vector[ij] > 0 {
				dtmp = float64(feature_Vector[ij]) - g.DMean[ii][ij]
				dtmp = dtmp * dtmp
				dsum += dtmp / g.DCovar[ii][ij]
				isum++
			}
		}
		dsum = -float64(isum)*constant.DLOG2PAI - math.Log(ddet) - dsum
		dsum /= 2
		dsum += math.Log(g.DMixtureWeight[ii])
		dresult = g.LogAdd(dresult, dsum)
	}

	return dresult
}

func (g *GMM) LVectorTopProb(featureVector [][]float32, iFrmIndex int) float64 {
	if !g.BMLoaded {
		panic("Model not loaded")
	}
	var dsum float64
	var dresult float64 = constant.LOGZERO

	for ii := 0; ii < g.ITopDistribNB; ii++ {
		dsum = g.LMixProb(featureVector[iFrmIndex], g.TopList[iFrmIndex][ii]) +
			math.Log(g.DMixtureWeight[g.TopList[iFrmIndex][ii]])
		dresult = g.LogAdd(dresult, dsum)
	}

	return dresult
}

/* Likelihood calculation routines */

//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////

func (g *GMM) FrameOcc(frame []float32, mix_index int) float64 {
	return g.LogAdd(math.Log(g.DMixtureWeight[mix_index]), g.LMixProb(frame, mix_index))
}

// Routine for adding two log-values in a linear scale, return the log
// result in double
func (g *GMM) LogAdd(lvar1 float64, lvar2 float64) float64 {
	var diff, z float64
	var minLogExp = -math.Log(-(constant.LOGZERO))
	if lvar1 < lvar2 {
		lvar1, lvar2 = lvar2, lvar1
	}

	diff = lvar2 - lvar1
	if diff < minLogExp {
		if lvar1 < constant.LSMALL {
			return constant.LOGZERO
		} else {
			return lvar1
		}
	} else {
		z = math.Exp(diff)
		return lvar1 + math.Log(1.0+z)
	}
}

// Return the log likelihood of the given vector to the given
// mixture (the score should be multiplied with the mixture weight).
//    FParam :  pointer to the feature vector in float
//  MixIndex :  Mixture index
// Return Value:
//    The log-likelihood in double
func (g *GMM) LMixProb(buffer []float32, mixIndex int) float64 {
	if !g.BMLoaded {
		panic("Model not loaded")
	}
	var dsum, dtmp float64
	dsum = .0
	for ii := 0; ii < g.IVectorSize; ii++ {
		dtmp = float64(buffer[ii]) - g.DMean[mixIndex][ii]
		dsum += dtmp * dtmp / g.DCovar[mixIndex][ii]
	}

	dsum = -(float64(g.IVectorSize) * constant.DLOG2PAI) - g.DLDet[mixIndex] - dsum
	dsum /= 2

	if dsum < constant.LOGZERO {
		return constant.LOGZERO
	} else {
		return dsum
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////

/* top distributions */

func (g *GMM) TopDistribs(buffer [][]float32, topDistribs int) bool {
	if !g.BMLoaded {
		panic("Model not loaded")
	}

	if g.BTopLoaded {
		g.CleanUpTop()
	}

	if topDistribs > g.INumMixtures {
		topDistribs = g.INumMixtures
	}

	var dsum float64

	g.ITopDistribNB = topDistribs

	distribs := make([]*Distributor, g.INumMixtures, g.INumMixtures)
	dlogmixw := make([]float64, g.INumMixtures, g.INumMixtures)
	for ii := 0; ii < g.INumMixtures; ii++ {
		if g.DMixtureWeight[ii] <= 0 {
			dlogmixw[ii] = constant.LOGZERO
		} else {
			dlogmixw[ii] = math.Log(g.DMixtureWeight[ii])
		}
	}

	g.TopList = make([][]int, g.IFrames, g.IFrames)
	for i := 0; i < g.IFrames; i++ {
		g.TopList[i] = make([]int, g.ITopDistribNB, g.ITopDistribNB)
	}

	for ik := 0; ik < g.IFrames; ik++ {
		for ii := 0; ii < g.INumMixtures; ii++ {
			dsum = g.LMixProb(buffer[ik], ii) + dlogmixw[ii]
			distribs[ii].index = ii
			distribs[ii].score = dsum
		}

		sort.Sort(Distributors(distribs))

		for ii := 0; ii < g.ITopDistribNB; ii++ {
			g.TopList[ik][ii] = distribs[g.INumMixtures-1-ii].index
		}
	}
	distribs = nil
	dlogmixw = nil

	return false
}

//func (g *GMM) ORBP(modelPath string, interval int) bool {
//	return false
//}

func (g *GMM) CopyTopDistribs(fromGMM *GMM) error {
	if !fromGMM.BMLoaded || !fromGMM.BPLoaded {
		return fmt.Errorf("source GMM has not loaded model or feature parameter")
	}

	if g.BTopLoaded {
		g.CleanUpTop()
	}

	g.BTopLoaded = true
	g.IFrames = fromGMM.IFrames
	g.ITopDistribNB = fromGMM.ITopDistribNB

	g.TopList = make([][]int, fromGMM.IFrames)

	for i := 0; i < fromGMM.IFrames; i++ {
		g.TopList[i] = make([]int, fromGMM.ITopDistribNB, fromGMM.ITopDistribNB)
		for j := 0; j < fromGMM.ITopDistribNB; j++ {
			g.TopList[i][j] = fromGMM.TopList[i][j]
		}
	}

	return nil
}

/* top distributions */

//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////

/* GMM super vector */

func (g *GMM) GMM2GSV() {
	if !g.BgsvLoaded {
		g.DGsv = make([]float64, g.INumMixtures*g.IVectorSize, g.INumMixtures*g.IVectorSize)
	}

	var k int
	for i := 0; i < g.INumMixtures; i++ {
		for j := 0; j < g.IVectorSize; j++ {
			g.DGsv[k] = g.DMean[i][j]
			k++
		}
	}

	g.BgsvLoaded = true
}

// GSVType = MEAN_ONLY,
// GSVType = KLDIVERGENCE,
func (g *GMM) GSVnorm(GSVType int) {
	if GSVType == MEAN_ONLY {
		return
	}

	for i := 0; i < g.INumMixtures; i++ {
		for j := 0; j < g.IVectorSize; j++ {
			if GSVType == KLDIVERGENCE {
				g.DGsv[i*g.IVectorSize+j] *= math.Sqrt(g.DMixtureWeight[i] / g.DCovar[i][j])
			} else {
				g.DGsv[i*g.IVectorSize+j] /= math.Sqrt(g.DCovar[i][j])
			}
		}
	}
}

func (g *GMM) CleanUpGSV() {
	if g.BgsvLoaded {
		g.DGsv = nil
	}
}

/* GMM super vector */

//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////

