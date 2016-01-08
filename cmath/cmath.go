package cmath

import (
	"fmt"
	"govpr/constant"
	"math"
)

//type float64 float64

type CMath struct {
	m_dctbuf1     []float64
	m_dctbuf2     []float64
	m_len_dctbuf1 int
	m_len_dctbuf2 int

	m_fftbuf1     []float64
	m_fftbuf2     []float64
	m_len_fftbuf1 int
	m_len_fftbuf2 int

	MinLogExp float64
}

func NewCMath() *CMath {
	return &CMath{
		MinLogExp: -math.Log(-constant.LOGZERO),
	}
}

//------------------- Log Arithmetic ------------------
// Log routines are very useful to the overflow and underflow
//  problems.
// The types LogFloat and float64 are used for representing
//  real numbers on a log scale.  LZERO is used for log(0)
//  in log arithmetic, any log real value <= LSMALL is
//  considered to be zero.

//------------------------------------------------------------------
// Convert log(x) to double, result is floored to
//  0.0 if x < LSMALL
func (cm *CMath) L2F(x float64) float64 {
	if x < constant.LSMALL {
		return 0.0
	}
	return math.Exp(x)
}

//------------------------------------------------------------------
// Return diff x - y on log scale, diff < LSMALL is
//  floored to LZERO
func (cm *CMath) LSub(x, y float64) float64 {
	if x < y {
		return constant.LSMALL
	}

	var diff, z float64

	diff = y - x
	if diff < cm.MinLogExp {
		if x < constant.LSMALL {
			return constant.LOGZERO
		} else {
			return x
		}
	} else {
		z = 1.0 - math.Exp(diff)
		if z < constant.MINLARG {
			return constant.LOGZERO
		} else {
			return x + math.Log(z)
		}
	}
}

//------------------------------------------------------------------
// Return sum x + y on log scale, sum < LSMALL
//  is floored to LZERO
func (cm *CMath) LAdd(x, y float64) float64 {
	if x < y {
		x, y = y, x
	}

	diff := y - x
	if diff < cm.MinLogExp {
		if x < constant.LSMALL {
			return constant.LOGZERO
		} else {
			return x
		}
	} else {
		z := math.Exp(diff)
		return x + math.Log(1.0+z)
	}
}

//------------------- Fast Fourier Transformation ------------------
// routine of fft
// - Arguments -
//      ar  	: pointer to the sequence of the real part
//      ai 		: pointer to the sequence of the imaginary part
//      Length  : length of the vector, it must be 2^N, otherwise
//               the vector must be padded with zero.
// - Return value -
//    true if successful, false if Length isn't padded to be 2^N
func (cm *CMath) FFT(ar, ai []float64, Length int) error {

	if ar == nil || len(ar) == 0 || ai == nil || len(ai) == 0 || Length <= 0 {
		return fmt.Errorf("invalid param")
	}

	m := int(math.Logb(float64(Length)))
	if math.Pow(2.0, float64(m)) != float64(Length) {
		//	if int(math.Pow(2.0,float64(m))) != Length {
		return fmt.Errorf("invalid length")
	}

	var temr, temi, x, y, temr1, temi1 float64
	if cm.m_fftbuf1 == nil {
		cm.m_fftbuf1 = make([]float64, Length, Length)
		cm.m_len_fftbuf1 = Length
	} else if cm.m_len_fftbuf1 != Length {
		cm.m_len_fftbuf1 = Length
		cm.m_fftbuf1 = make([]float64, Length, Length)
	}

	if cm.m_fftbuf2 == nil {
		cm.m_fftbuf2 = make([]float64, Length, Length)
		cm.m_len_fftbuf2 = Length
	} else if cm.m_len_fftbuf2 != Length {
		cm.m_len_fftbuf2 = Length
		cm.m_fftbuf2 = make([]float64, Length, Length)
	}

	// fill buffers with precalculated values
	temr = 2 * constant.PI / float64(Length)
	for l := 0; l < Length; l++ {
		cm.m_fftbuf1[l] = math.Sin(temr * float64(l))
		cm.m_fftbuf2[l] = math.Cos(temr * float64(l))
	}

	cm.fftsort(ar, ai, Length)
	b := 1
	for l := 1; l <= m; l++ {
		for j := 0; j <= b; j++ {
			p := (1 << uint(m-l)) * j
			x = cm.m_fftbuf2[p]
			y = -cm.m_fftbuf1[p]
			p = 1 << uint(l)

			for k := j; k < Length; k += p {
				temr = ar[k+b]*x - ai[k+b]*y
				temi = ar[k+b]*y + ai[k+b]*x
				temr1 = ar[k] + temr
				temi1 = ai[k] + temi
				ar[k+b] = ar[k] - temr
				ai[k+b] = ai[k] - temi
				ar[k] = temr1
				ai[k] = temi1
			}
		}
		b <<= 1
	}

	return nil
}

// routine of real dct
// - Arguments -
//		 dbuff : pointer to the real vector
//      iInLen : length of the input vector
//     iOutLen : only the first "iOutLen" elements will
//                be computed, if iOutLen <= 0, all elements
//                will be computed and returned
// - Return value -
//    true if successful, false if any error occured.
func (cm *CMath) DCT(dbuff []float64, iInLen int) error {
	return cm.DCT2(dbuff, iInLen, 0)
}

func (cm *CMath) DCT2(dbuff []float64, iInLen, iOutLen int) error {
	var ioffset int
	var iN2 int = iInLen << 1
	var icfft [2]int = [2]int{1, -1}
	var dfactor, dc0 float64 = math.Sqrt(2.0 / float64(iInLen)), 1.0 / math.Sqrt(2.0)
	if iOutLen <= 0 {
		iOutLen = iInLen
	}

	if cm.m_dctbuf1 == nil {
		cm.m_dctbuf1 = make([]float64, iInLen, iInLen)
		cm.m_len_dctbuf1 = iInLen
	} else if cm.m_len_dctbuf1 != iInLen {
		cm.m_len_dctbuf1 = iInLen
		cm.m_dctbuf1 = make([]float64, iInLen, iInLen)
	}

	if cm.m_dctbuf2 == nil {
		cm.m_dctbuf2 = make([]float64, iN2, iN2)
		cm.m_len_dctbuf2 = iN2
	} else if cm.m_len_dctbuf2 != iN2 {
		cm.m_len_dctbuf2 = iN2
		cm.m_dctbuf2 = make([]float64, iN2, iN2)
	}

	// Create Cosine Table & Copy data to the source buff
	for i := 0; i < iN2; i++ {
		cm.m_dctbuf2[i] = math.Cos(float64(i) * constant.PI / float64(iN2))
		if i < iInLen {
			cm.m_dctbuf1[i] = dbuff[i]
		}
	}

	//-------------------------
	// Main Procedure
	for k := 0; k < iOutLen; k++ {
		dbuff[k] = 0.0
		for n := 0; n < iInLen; n++ {
			ioffset = ((n << 1) + 1) * k
			dbuff[k] += cm.m_dctbuf1[n] * float64(icfft[(ioffset/iN2)%2]) * cm.m_dctbuf2[ioffset%iN2]
		}
		if k == 0 {
			dbuff[k] *= dc0
		}
		dbuff[k] *= dfactor
	}

	return nil
}

// sort routine for FFT
func (cm *CMath) fftsort(realx, imagex []float64, length int) {
	j := length >> 1
	for i := 1; i < length-1; i++ {
		if i < j {
			realx[i], realx[j] = realx[j], realx[i]
			imagex[i], imagex[j] = imagex[j], imagex[i]
		}

		k := length >> 1
		for j >= k {
			j = j - k
			k >>= 1
		}

		j = j + k
	}
}
