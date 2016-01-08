package constant

const (
	//	Speaker Recognition ---------------
	REL_FACTOR = 16 //自适应因子
	MAXLOP     = 1  //自适应次数

	BIT_PER_SAMPLE   = 16
	SAMPLERATE       = 16000 //采样率
	MAXFRAMES        = 31000 //特征缓冲区最大帧数
	MIN_FRAMES       = 300
	MIN_TRAIN_FRAMES = 1000
	DB               = -3.0      //归一化分贝量
	LOGZERO          = (-1.0E10) /* ~log(0) */
	LSMALL           = (-0.5E10) /* log values < LSMALL are set to LOGZERO */
	MIN_AVG_FRAME    = 1         // Minimal frames per mixture
	DOUBLEZERO       = 1.0e-12

	IN_XML    = 1
	IN_BINARY = 2
	IN_HTK    = 3

	ENERGYTHREHOLD = 35 //  基于帧能量的端点检测阈值
)

const (
	PI      = 3.14159265358979
	MINEARG = (-708.3)  /* lowest exp() arg  = log(MINLARG) */
	MINLARG = 2.45E-308 /* lowest log() arg  = exp(MINEARG) */
)

const (
	//	GMM class -----------------------
	MAX_CUS_TAG = 15    // Maximum length of Tags in Custom GMM definition file
	MAX_HTK_TAG = 10    // Maximum length of Tags in HTK format GMM definition file
	TAG_CUS_NUM = 8     // Maximum types of Tags in GMM definition file
	TAG_HTK_NUM = 15    // Maxinum types of Tags in HTK format GMM definition file
	MAX_FILES   = 8000  // Maximum parameter files to be loaded when training
	VAR_FLOOR   = 0.005 // Variance floor, make sure the variance to be large enough	(old:0.005)
	VAR_CEILING = 10.0  // Variance ceiling

	MAX_LOOP = 10
)

const (
	// Wave
	VOC_BLOCK_LEN int   = 1600
	MIN_VOC_ENG   int   = 500
	INT16MAX      int16 = 32767
	SHRT_MAX      int   = 35535
)

var (
	TOP_MIXS int     = 30 // = 5
	DLOG2PAI float64 = 1.837877066
)

const (
	//	feature  -------------------------
	PARAMCONF_BHTK  = false
	PARAMCONF_LCUT  = 250
	PARAMCONF_HCUT  = 3800
	PARAMCONF_NFB   = 24
	PARAMCONF_NFLEN = 20
	PARAMCONF_NFSFT = 10
	PARAMCONF_NMFCC = 16
	PARAMCONF_BS0   = false
	PARAMCONF_BD0   = false
	PARAMCONF_BA0   = false
	PARAMCONF_BS    = true
	PARAMCONF_BD    = true
	PARAMCONF_BA    = false

	PARAMCONF_CMSVN          = true
	PARAMCONF_DBNORM         = true
	PARAMCONF_ZEROGLOBALMEAN = true
	PARAMCONF_FEATWARP       = false
	PARAMCONF_DIFPOL         = false
	PARAMCONF_DPSCC          = false
	PARAMCONF_PDASCC         = false
	PARAMCONF_ENERGYNORM     = false
	PARAMCONF_RASTA          = false

	PARAMCONF_SILFLOOR     = 50
	PARAMCONF_ENERGYSCALE  = 19
	PARAMCONF_FEATWARP_WIN = 300
	PARAMCONF_RASTA_COFF   = 0.94
)
