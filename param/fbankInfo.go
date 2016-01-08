package param

type FBankInfo struct {
	ISampRate  int     // sample rate (samples / second)
	IFrameSize int     // frame size (in samples)
	IfftN      int     // fft size (2^N)
	FFRes      float32 // fft resolution ( (SampRate/2) / (fftN/2) = SampRate/fftN )
	INumFB     int     // number of filterbanks
	Iklo, Ikhi int     // lopass to hipass cut-off fft indices

	PCF     []float32 // array[1..iNumFB+1] of centre freqs
	PloChan []int16   // array[1..ifftN/2] of lower fbank index
	PloWt   []float32 // array[1..ifftN/2] of lower fbank weighting
	Pdatar  []float64 // array[1..ifftN] of fft bins (real part)
	Pdatai  []float64 // array[1..ifftN] of fft bins (image part)

	BUsePower   bool // use power rather than magnitude (d: false)
	BTakeLogs   bool // log filterbank channels (d: true)
	BPreEmph    bool // pre emphasize (d: true)
	BUseHamming bool // Use Hamming window rather than rectangle window (d: true)
}
