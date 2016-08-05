package param

type FilterBank struct {
	sampleRate     int     // sample rate (samples / second)
	frameSize      int     // frame size (in samples)
	fttSize        int     // fft size (2^N)
	fttResolution  float32 // fft resolution ( (SampRate/2) / (fftN/2) = SampRate/fftN )
	filterBankSize int     // number of filterbanks
	start, end     int     // lopass to hipass cut-off fft indices

	centerFreqs            []float32 // array[1..filterBankSize+1] of centre freqs
	lowerFilterBanksIndex  []int16   // array[1..fttSize/2] of lower fbank index
	lowerFilterBanksWeight []float32 // array[1..fttSize/2] of lower fbank weighting
	fftRealValue           []float64 // array[1..fttSize] of fft bins (real part)
	fftComplexValue        []float64 // array[1..fttSize] of fft bins (image part)

	isUsePower      bool // use power rather than magnitude (d: false)
	isLogFBChannels bool // log filterbank channels (d: true)
	isPreEmphasize  bool // pre emphasize (d: true)
	isUseHamming    bool // Use Hamming window rather than rectangle window (d: true)
}
