package feature

type ParamConf struct {
	bHTK            bool   // HTK if true or WAV if false
	lcut            uint // low cut-off
	hcut            uint // high cut-off
	nfb             int  // # num of filter-bank
	nflen           int	   // # frame length
	nfsft           int  // 10 # frame shift
	nmfcc           int  // 16 # mfcc order
	bs0             bool   // f	# static mfcc_0
	bs              bool   // t	# static mfcc
	bd0             bool   // f	# dynamic mfcc_0
	bd              bool   // t	# dynamic mfcc
	ba0             bool   // f	# acce mfcc_0
	ba              bool   // f	# acce mfcc
	cmsvn           bool   // t	# cmsvn
	ZeroGlobalMean  bool   // t # zero global mean
	bdBNorm         bool   // t # decibel normalization
	bDiffpolish     bool   // f	# polish differential formula
	bdpscc          bool   // f	# differentail power spectrum
	bpdascc         bool   // f	# predictive differential amplitude spectrum
	bEnergyNorm     bool
	silFloor        int16
	energyscale     int16
	bFeatWarping    bool
	featWarpWinSize int16
	bRasta          bool
	rastaCoff       float64
}
