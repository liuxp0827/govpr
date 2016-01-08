package param

type MFCCInfo struct {
	IOrder          int     // MFCC order (except the 0th)
	IDeltaWin       int     // length dynamic window when take delta params, default is 2
	BFBank          bool    // output fbank other than MFCC
	B0              bool    // output MFCC_0 or not
	BD0             bool    // output delta MFCC_0 or not
	BA0             bool    // output acceleration MFCC_0 or not
	BStatic         bool    // static coefficients or not
	BDelta          bool    // delta coefficients or not
	BAcce           bool    // acceleration coefficients or not
	BCepLift        bool    // lift the cepstral or not
	FFrmRate        float32 // frame rate in ms
	FCepsLifter     float32 // cepstral lifter. It's invalid when bFBank is set to true
	BPolishDiff     bool    // polish differential formula
	BdBNorm         bool    // decibel normalization
	BDPSCC          bool    // Differential Power Spectrum
	BPDASCC         bool    // Predictive Differential Amplitude Spetrum
	ZeroGlobalMean  bool
	BEnergyNorm     bool
	SilFloor        int16
	Energyscale     int16
	BFeatWarping    bool
	FeatWarpWinSize int16
	BRasta          bool
	RastaCoff       float64
}
