package govpr

import "fmt"


var (
	LSV_ERR_ENGINE_NOT_INIT      error = fmt.Errorf("engine not init")
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
	LSV_ERR_CONF_PARAM           error = fmt.Errorf("conf param error")
	LSV_ERR_NO_ACTIVE_SPEECH     error = fmt.Errorf("no active speech")
	LSV_ERR_INVALID_PARAM        error = fmt.Errorf("invalid param")
)

func NewError(err error, e string) error {
	return fmt.Errorf("%s: %s", err.Error(), e)
}
