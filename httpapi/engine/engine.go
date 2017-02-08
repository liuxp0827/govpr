package engine

import (
	"github.com/astaxie/beego"
	"github.com/liuxp0827/govpr"
	"github.com/liuxp0827/govpr/log"
)

type engine struct {
	vprEngine *govpr.VPREngine
}

var (
	ubm_path string = beego.AppConfig.DefaultString("ubm_path", "vpr/ubm")
)

func NewEngine(sampleRate, delSilRange int, userModelFile string) (*engine, error) {

	vEngine, err := govpr.NewVPREngine(sampleRate, delSilRange, false, ubm_path, userModelFile)
	if err != nil {
		return nil, err
	}

	return &engine{
		vprEngine: vEngine,
	}, nil
}

func (this *engine) DestroyEngine() {
	this.vprEngine = nil
}

func (this *engine) TrainSpeech(c int, buffers [][]byte, texts []string, userid, token string) error {
	var err error
	count := len(buffers)
	for i := 0; i < count; i++ {
		err = this.vprEngine.AddTrainBuffer(buffers[i])
		if err != nil {
			log.Error(err)
			return err
		}
	}

	defer this.vprEngine.ClearTrainBuffer()
	defer this.vprEngine.ClearAllBuffer()

	err = this.vprEngine.TrainModel()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (this *engine) RecSpeech(buffer []byte, text string, userid, token string) (float64, error) {
	err := this.vprEngine.AddVerifyBuffer(buffer)
	defer this.vprEngine.ClearVerifyBuffer()
	if err != nil {
		return -1.0, err
	}

	err = this.vprEngine.VerifyModel()
	if err != nil {
		return -1.0, err
	}

	return this.vprEngine.GetScore(), nil
}
