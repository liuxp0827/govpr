package math

import (
	"fmt"
	"github.com/liuxp0827/govpr/constant"
	"math"
)

//type float64 float64

var dct1 []float64
var dct2 []float64

var fft1 []float64
var fft2 []float64

//------------------- Fast Fourier Transformation ------------------
// routine of fft
// - Arguments -
//      ar  	: pointer to the sequence of the real part
//      ai 		: pointer to the sequence of the imaginary part
//      length  : length of the vector, it must be 2^N, otherwise
//               the vector must be padded with zero.
func FFT(ar, ai []float64, length int) error {

	if ar == nil || len(ar) == 0 || ai == nil || len(ai) == 0 || length <= 0 {
		return fmt.Errorf("invalid param")
	}

	m := int(math.Logb(float64(length)))
	if math.Pow(2.0, float64(m)) != float64(length) {
		return fmt.Errorf("invalid length")
	}

	var temr, temi, x, y, temr1, temi1 float64
	if fft1 == nil {
		fft1 = make([]float64, length, length)
	} else if len(fft1) != length {
		fft1 = make([]float64, length, length)
	}

	if fft2 == nil {
		fft2 = make([]float64, length, length)

	} else if len(fft2) != length {
		fft2 = make([]float64, length, length)
	}

	// fill buffers with precalculated values
	temr = 2 * constant.PI / float64(length)
	for i := 0; i < length; i++ {
		fft1[i] = math.Sin(temr * float64(i))
		fft2[i] = math.Cos(temr * float64(i))
	}

	fftSort(ar, ai, length)

	var b int = 1

	for l := 1; l <= m; l++ {
		for j := 0; j < b; j++ {

			p := (1 << uint(m-l)) * j
			x = fft2[p]
			y = -fft1[p]
			p = 1 << uint(l)

			for k := j; k < length; k += p {
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

// Discrete Cosine Transform
// - Arguments -
//		 data : pointer to the real vector
//     width : only the first "width" elements will
//                be computed, if width <= 0, all elements
//                will be computed and returned
func DCT(data []float64, width *int) error {
	var ioffset int
	var length int = len(data)
	var length2 int = length << 1
	var icfft [2]int = [2]int{1, -1}
	var dfactor, dc0 float64 = math.Sqrt(2.0 / float64(length)), 1.0 / math.Sqrt(2.0)
	if *width <= 0 {
		*width = length
	}

	if dct1 == nil {
		dct1 = make([]float64, length, length)
	} else if len(dct1) != length {
		dct1 = make([]float64, length, length)
	}

	if dct2 == nil {
		dct2 = make([]float64, length2, length2)
	} else if len(dct2) != length2 {
		dct2 = make([]float64, length2, length2)
	}

	// Create Cosine Table & Copy data to the source buff
	for i := 0; i < length2; i++ {
		dct2[i] = math.Cos(float64(i) * constant.PI / float64(length2))
		if i < length {
			dct1[i] = data[i]
		}
	}

	//-------------------------
	// Main Procedure
	for k := 0; k < *width; k++ {
		data[k] = 0.0
		for n := 0; n < length; n++ {
			ioffset = ((n << 1) + 1) * k
			data[k] += dct1[n] * float64(icfft[(ioffset/length2)%2]) * dct2[ioffset%length2]
		}
		if k == 0 {
			data[k] *= dc0
		}
		data[k] *= dfactor
	}

	return nil
}

// sort routine for FFT
func fftSort(realx, imagex []float64, length int) {
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
