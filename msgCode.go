package govpr

import "fmt"

const (
	LSV_ENGINE_TYPE_VERIFICATION        = -17
	LSV_ENGINE_TYPE_TRAIN               = -18
	LSV_ENGINE_TYPE_ADAPTATION          = -19
	LSV_ENGINE_TYPE_IDENTIFICATION      = -20
	LSV_ENGINE_TYPE_OFFLINEVERIFICATION = -21

	LSV_ERR_VOICE_NOT_REGISTERED = -22

	LSV_ENGINE_TYPE_TD = -23
	LSV_ENGINE_TYPE_TI = -24
)

var (
	LSV_ERR_ENGINE_NOT_INIT      error = fmt.Errorf("")
	LSV_ERR_TIMEOUT              error = fmt.Errorf("timeout")
	LSV_ERR_NEED_MORE_SAMPLE     error = fmt.Errorf("need more sample ")
	LSV_ERR_ILLEGAL_HANDLE       error = fmt.Errorf("illegal handle")
	LSV_ERR_FILE_ERROR           error = fmt.Errorf("file error")
	LSV_ERR_NO_AVAILABLE_DATA    error = fmt.Errorf("no available data")
	LSV_ERR_VOICE_TOO_SHORT      error = fmt.Errorf("voice too short")
	LSV_ERR_TRAINING_FAILED      error = fmt.Errorf("train failed")
	LSV_ERR_VERIFY_FAILED        error = fmt.Errorf("verify failed")
	LSV_ERR_MODEL_NOT_FOUND      error = fmt.Errorf("model not found")
	LSV_ERR_MODEL_LOAD_FAILED    error = fmt.Errorf("model load failed")
	LSV_ERR_MEM_INSUFFICIENT     error = fmt.Errorf("memory insufficient")
	LSV_ERR_WORKINGDIR_NOT_FOUND error = fmt.Errorf("workDir not found")
	LSV_ERR_CONF_PARAM           error = fmt.Errorf("conf param error")
	LSV_ERR_NO_ACTIVE_SPEECH     error = fmt.Errorf("no active speech")
	LSV_ERR_INVALID_PARAM        error = fmt.Errorf("invalid param")
)

func NewError(err error, e string) error {
	return fmt.Errorf("%s: %s", err.Error(), e)
}
