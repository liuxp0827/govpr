package constant

const (
	//	Speaker Recognition ---------------
	REL_FACTOR = 16 // 自适应因子
	MAXLOP     = 1  // 自适应次数

	BIT_PER_SAMPLE = 16
	SAMPLERATE     = 16000 // 采样率
	MIN_FRAMES     = 300
	DB             = -3.0      // 归一化分贝量
	LOGZERO        = (-1.0E10) /* ~log(0) */
	LSMALL         = (-0.5E10) /* log values < LSMALL are set to LOGZERO */

)

const (
	PI = 3.14159265358979
)

const (
	VAR_FLOOR   = 0.005 // Variance floor, make sure the variance to be large enough	(old:0.005)
	VAR_CEILING = 10.0  // Variance ceiling
	MAX_LOOP    = 10
)

const (
	// Wave
	VOC_BLOCK_LEN int = 1600
	MIN_VOC_ENG   int = 500
	SHRT_MAX      int = 35535
)

var (
	DLOG2PAI float64 = 1.837877066
)

const (
	LOW_CUT_OFF      = 250
	HIGH_CUT_OFF     = 3800
	FILTER_BANK_SIZE = 24
	FRAME_LENGTH     = 20
	FRAME_SHIFTt     = 10
	MFCC_ORDER       = 16
	BSTATIC          = true
	BDYNAMIC         = true
	BACCE            = false

	CMSVN          = true
	DBNORM         = true
	ZEROGLOBALMEAN = true
	FEATWARP       = false
	DIFPOL         = false
	DPSCC          = false
	PDASCC         = false
	ENERGYNORM     = false
	RASTA          = false

	SIL_FLOOR                = 50
	ENERGY_SCALE             = 19
	FEATURE_WARPING_WIN_SIZE = 300
	RASTA_COFF               = 0.94
)
