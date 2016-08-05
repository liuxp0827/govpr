package gmm

import (
	"fmt"
	"github.com/liuxp0827/govpr/constant"
	"github.com/liuxp0827/govpr/file"
	"github.com/liuxp0827/govpr/log"
	"math"
)

const (
	MEAN_ONLY = iota
	KLDIVERGENCE
)

type NormParam struct {
	mean   float64 // = 0.0
	stdVar float64 // = 1.0
}

type GMM struct {
	Frames      int         // number of total frames
	FeatureData [][]float32 // feature buffer

	VectorSize      int         // Vector size
	Mixtures        int         // Mixtures of the GMM
	deterCovariance []float64   // determinant of the covariance matrix [mixture]
	MixtureWeight   []float64   // weight of each mixture[mixture]
	Mean            [][]float64 // mean vector [mixture,dimension]
	Covar           [][]float64 // covariance (diagonal) [mixture,dimension]
}

func NewGMM() *GMM {
	gmm := &GMM{
		FeatureData:   make([][]float32, 0),
		MixtureWeight: make([]float64, 0),
		Mean:          make([][]float64, 0),
		Covar:         make([][]float64, 0),
	}
	return gmm
}

/* Model file access routines */

func (g *GMM) DupModel(gmm *GMM) {
	g.Mixtures = gmm.Mixtures
	g.VectorSize = gmm.VectorSize
	g.deterCovariance = make([]float64, g.Mixtures, g.Mixtures)
	g.MixtureWeight = make([]float64, g.Mixtures, g.Mixtures)
	g.Mean = make([][]float64, g.Mixtures, g.Mixtures)
	g.Covar = make([][]float64, g.Mixtures, g.Mixtures)

	for i := 0; i < g.Mixtures; i++ {
		g.deterCovariance[i] = gmm.deterCovariance[i]
		g.MixtureWeight[i] = gmm.MixtureWeight[i]
		g.Mean[i] = make([]float64, g.VectorSize, g.VectorSize)
		g.Covar[i] = make([]float64, g.VectorSize, g.VectorSize)
		for j := 0; j < g.VectorSize; j++ {
			g.Mean[i][j] = gmm.Mean[i][j]
			g.Covar[i][j] = gmm.Covar[i][j]
		}
	}
}

func (g *GMM) LoadModel(filename string) error {
	reader, err := file.NewVPRFile(filename)
	if err != nil {
		return err
	}

	g.Mixtures, err = reader.GetInt()
	if err != nil {
		log.Error(err)
		return err
	}

	g.VectorSize, err = reader.GetInt()
	if err != nil {
		log.Error(err)
		return err
	}

	g.deterCovariance = make([]float64, g.Mixtures, g.Mixtures)
	g.MixtureWeight = make([]float64, g.Mixtures, g.Mixtures)
	g.Mean = make([][]float64, g.Mixtures, g.Mixtures)
	g.Covar = make([][]float64, g.Mixtures, g.Mixtures)
	for i := 0; i < g.Mixtures; i++ {
		g.Mean[i] = make([]float64, g.VectorSize, g.VectorSize)
		g.Covar[i] = make([]float64, g.VectorSize, g.VectorSize)
	}

	for i := 0; i < g.Mixtures; i++ {
		g.deterCovariance[i] = 0.0
	}

	for i := 0; i < g.Mixtures; i++ {
		g.MixtureWeight[i], err = reader.GetFloat64()
		if err != nil {
			log.Error(err)
			return err
		}
	}

	for i := 0; i < g.Mixtures; i++ {
		_, err = reader.GetFloat64() // not used
		if err != nil {
			log.Error(err)
			return err
		}

		_, err = reader.GetFloat64() // not used
		if err != nil {
			log.Error(err)
			return err
		}

		_, err = reader.GetByte() // not used
		if err != nil {
			log.Error(err)
			return err
		}

		for j := 0; j < g.VectorSize; j++ {
			g.Covar[i][j], err = reader.GetFloat64()
			if err != nil {
				log.Error(err)
				return err
			}

			g.deterCovariance[i] += math.Log(g.Covar[i][j])
		}

		for j := 0; j < g.VectorSize; j++ {
			g.Mean[i][j], err = reader.GetFloat64()
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}
	reader.Close()
	return nil
}

func (g *GMM) SaveModel(filename string) error {
	writer, err := file.NewVPRFile(filename)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.PutInt(g.Mixtures)
	if err != nil {
		return err
	}

	_, err = writer.PutInt(g.VectorSize)
	if err != nil {
		return err
	}

	for i := 0; i < g.Mixtures; i++ {
		_, err = writer.PutFloat64(g.MixtureWeight[i])
		if err != nil {
			return err
		}
	}

	for i := 0; i < g.Mixtures; i++ {
		_, err = writer.PutFloat64(0.0) // not used
		if err != nil {
			return err
		}

		_, err = writer.PutFloat64(0.0) // not used
		if err != nil {
			return err
		}

		err = writer.PutByte(byte(0)) // not used
		if err != nil {
			return err
		}

		for j := 0; j < g.VectorSize; j++ {
			_, err = writer.PutFloat64(g.Covar[i][j])
			if err != nil {
				return err
			}
		}

		for j := 0; j < g.VectorSize; j++ {
			_, err = writer.PutFloat64(g.Mean[i][j])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *GMM) CopyFeatureData(gmm *GMM) error {

	g.Frames = gmm.Frames
	g.VectorSize = gmm.VectorSize
	g.FeatureData = make([][]float32, gmm.Frames, gmm.Frames)
	for i := 0; i < g.Frames; i++ {
		g.FeatureData[i] = make([]float32, g.VectorSize, g.VectorSize)
	}

	for i := 0; i < g.Frames; i++ {
		for j := 0; j < g.VectorSize; j++ {
			g.FeatureData[i][j] = gmm.FeatureData[i][j]
		}
	}

	return nil
}

/* Model file access routines */

/* Model estimation routines */

func (g *GMM) EM(mixtures int) (int, error) {
	var dlogfrmprob, rubbish, lastrubbish float64
	var dsumgama, dlogmixw, dgama, dmixw []float64
	var threshold float64 = 1e-5
	var mean, covar [][]float64
	var loop int = 0

	mean = make([][]float64, mixtures, mixtures)
	covar = make([][]float64, mixtures, mixtures)
	for i := 0; i < mixtures; i++ {
		mean[i] = make([]float64, g.VectorSize, g.VectorSize)
		covar[i] = make([]float64, g.VectorSize, g.VectorSize)
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
			if g.MixtureWeight[i] <= 0 {
				dlogmixw[i] = constant.LOGZERO
			} else {
				dlogmixw[i] = math.Log(g.MixtureWeight[i])
			}

			// clean up temporary values
			dmixw[i] = .0
			dsumgama[i] = .0
			for j := 0; j < g.VectorSize; j++ {
				mean[i][j] = .0
				covar[i][j] = .0
			}
		}

		for i := 0; i < g.Frames; i++ {
			dlogfrmprob = constant.LOGZERO
			for j := 0; j < mixtures; j++ {
				dgama[j] = g.LMixProb(g.FeatureData[i], j)
				dgama[j] += dlogmixw[j]
				dlogfrmprob = g.LogAdd(dgama[j], dlogfrmprob)
			}

			rubbish += dlogfrmprob

			for j := 0; j < mixtures; j++ {
				dgama[j] -= dlogfrmprob
				dgama[j] = math.Exp(dgama[j])
				dsumgama[j] += dgama[j]

				// update weights
				dmixw[j] += dgama[j]
				for k := 0; k < g.VectorSize; k++ {
					mean[j][k] += dgama[j] * float64(g.FeatureData[i][k])
					covar[j][k] += dgama[j] * float64(g.FeatureData[i][k]) * float64(g.FeatureData[i][k])
				}
			}
		}

		rubbish /= float64(g.Frames) // rubbish = LLR

		for i := 0; i < mixtures; i++ {
			if dsumgama[i] == .0 {
				return -1, nil
			}

			g.MixtureWeight[i] = dmixw[i] / float64(g.Frames)

			for j := 0; j < g.VectorSize; j++ {
				g.Mean[i][j] = mean[i][j] / dsumgama[i]
				g.Covar[i][j] = covar[i][j] / dsumgama[i]
				g.Covar[i][j] -= g.Mean[i][j] * g.Mean[i][j]

				if g.Covar[i][j] < constant.VAR_FLOOR {
					g.Covar[i][j] = constant.VAR_FLOOR
				}

				if g.Covar[i][j] > constant.VAR_CEILING {
					g.Covar[i][j] = constant.VAR_CEILING
				}
			}
		}

		for i := 0; i < mixtures; i++ {
			g.deterCovariance[i] = .0
			for j := 0; j < g.VectorSize; j++ {
				g.deterCovariance[i] += math.Log(g.Covar[i][j])
			}
		}
		loop++
		log.Debugf("loop: %02d, Average Log Likelihood: %f, Increment: %f", loop, rubbish, rubbish-lastrubbish)
		return 0, nil
	}

DO:
	ret, err := doing()
	if err != nil {
		return 0, err
	} else if ret == -1 {
		return 0, fmt.Errorf("error train loop")
	}

	for loop < constant.MAX_LOOP && math.Abs((rubbish-lastrubbish)/(lastrubbish+0.01)) > threshold {
		goto DO
	}

	if loop >= constant.MAX_LOOP {
		log.Debugf("Break at loop %d", loop)
	} else {
		log.Debugf("Converged at loop %d", loop)
	}

	return loop, nil
}

/* Model estimation routines */

//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////

/* Likelihood calculation routines */

func (g *GMM) LProb(featureBuf [][]float32, start, length int64) float64 {
	var dgama, dlogfrmprob, sum float64 = .0, .0, .0
	dlogmixw := make([]float64, g.Mixtures, g.Mixtures)
	for ij := 0; ij < g.Mixtures; ij++ {
		if g.MixtureWeight[ij] <= 0 {
			dlogmixw[ij] = constant.LOGZERO
		} else {
			dlogmixw[ij] = math.Log(g.MixtureWeight[ij])
		}
	}

	for ii := int64(start); ii < (start + length); ii++ {
		dlogfrmprob = constant.LOGZERO
		for jj := 0; jj < g.Mixtures; jj++ {
			dgama = g.LMixProb(featureBuf[ii], jj) + dlogmixw[jj]
			dlogfrmprob = g.LogAdd(dgama, dlogfrmprob)
		}
		sum += dlogfrmprob
	}

	dlogmixw = nil
	return sum
}

/* Likelihood calculation routines */

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
//    FeatureData :  pointer to the feature vector in float
//  MixIndex :  Mixture index
// Return Value:
//    The log-likelihood in float64
func (g *GMM) LMixProb(buffer []float32, mixIndex int) float64 {
	if g.Mean == nil || g.deterCovariance == nil {
		panic("Model not loaded")
	}

	var dsum, dtmp float64
	dsum = .0
	for ii := 0; ii < g.VectorSize; ii++ {
		dtmp = float64(buffer[ii]) - g.Mean[mixIndex][ii]
		dsum += dtmp * dtmp / g.Covar[mixIndex][ii]
	}

	dsum = -(float64(g.VectorSize) * constant.DLOG2PAI) - g.deterCovariance[mixIndex] - dsum
	dsum /= 2

	if dsum < constant.LOGZERO {
		return constant.LOGZERO
	} else {
		return dsum
	}
}
