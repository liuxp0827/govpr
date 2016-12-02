package param

type Mfcc struct {
	mfccOrder      int  // MFCC order (except the 0th)
	dynamicWinSize int  // length dynamic window when take delta params, default is 2
	isFilter       bool // output fbank other than MFCC

	IsStatic             bool    // static coefficients or not
	IsDynamic            bool    // dynamic coefficients or not
	IsAcce               bool    // acceleration coefficients or not
	IsLiftCepstral       bool    // lift the cepstral or not
	FrameRate            float32 // frame rate in ms
	CepstralLifter       float32 // cepstral lifter. It's invalid when bFBank is set to true
	IsPolishDiff         bool    // polish differential formula
	IsDBNorm             bool    // decibel normalization
	IsDiffPowerSpectrum  bool    // Differential Power Spectrum
	IsPredDiffAmpSpetrum bool    // Predictive Differential Amplitude Spetrum
	IsZeroGlobalMean     bool
	IsEnergyNorm         bool
	SilFloor             int16
	EnergyScale          int16
	IsFeatWarping        bool
	FeatWarpWinSize      int16
	IsRasta              bool
	RastaCoff            float64
}
