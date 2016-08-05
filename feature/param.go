package feature

type Param struct {
	lowCutOff              uint // low cut-off
	highCutOff             uint // high cut-off
	filterBankSize         int  // # num of filter-bank
	frameLength            int  // # frame length
	frameShift             int  // 10 # frame shift
	mfccOrder              int  // 16 # mfcc order
	isStatic               bool // t	# static mfcc
	isDynamic              bool // t	# dynamic mfcc
	isAcce                 bool // f	# acce mfcc
	cmsvn                  bool // t	# cmsvn
	isZeroGlobalMean       bool // t # zero global mean
	isDBNorm               bool // t # decibel normalization
	isDiffPolish           bool // f	# polish differential formula
	isDiffPowerSpectrum    bool // f	# differentail power spectrum
	isPredDiffAmplSpectrum bool // f	# predictive differential amplitude spectrum
	isEnergyNorm           bool
	silFloor               int16
	energyscale            int16
	isFeatWarping          bool
	featWarpWinSize        int16
	isRasta                bool
	rastaCoff              float64
}
