package main

import (
	"govpr"
	"io/ioutil"
	"govpr/log"
)

var (
	model_path string = "model/"
	vpr_dir    string = "vpr/"
)

type engineRet struct {
	FinishTrain bool
	Score       float32
	ErrCode     int
	Type        int
}

func setEngineRet(finishTrain bool, score float32, errCode, Type int) *engineRet {
	return &engineRet{
		FinishTrain: finishTrain,
		Score:       score,
		ErrCode:     errCode,
		Type:        Type,
	}
}

type engine struct {
	vprEngine *govpr.VPREngine
	err       *engineRet
}

func NewEngine() *engine {
	return &engine{
		vprEngine: govpr.NewVPREngine(0, 16000, 50, "vpr/", "model/", "test"),
		err:       nil,
	}
}

func (this *engine) DestroyEngine() {
	this.vprEngine = nil
}

func (this *engine) TrainSpeech(count int, buffers [][]byte, texts []string, userid string) {

	log.Info("start vpr TrainSpeech")
	vprTrainSpeech(this, count, buffers)
}

func (this *engine) RecSpeech(buffer []byte, text string, userid string) {
	log.Info("start vpr RecSpeech")
	vprRecSpeech(this, buffer)
}

func vprTrainSpeech(x *engine, count int, buffers [][]byte) error {
	engine := x.vprEngine

	var err error
	for i := 0; i < count; i++ {
		err = engine.AddTrainBuffer(buffers[i])
		if err != nil {
			return err
		}
	}

	defer engine.ClearTrainBuffer()
	defer engine.ClearAllBuffer()

	err = engine.TrainModel()
	if err != nil {
		return err
	}

	log.Info("vpr TrainSpeech")

	return nil
}

func vprRecSpeech(x *engine, buffer []byte) (err error) {
	engine := x.vprEngine

	err = engine.AddVerifyBuffer(buffer)
	defer engine.ClearVerifyBuffer()
	if err != nil {
		log.Error(err)
		return
	}

	err = engine.VerifyModel()
	if err != nil {
		log.Error(err)
		return
	}

	Score := engine.GetScore()
	log.Infof("得分：%f", Score)
	return
}

func main() {
	vprEngine := NewEngine()
	//trainlist := []string{
	//	"wav/train/01_32468975.wav",
	//	"wav/train/02_58769423.wav",
	//	"wav/train/03_59682734.wav",
	//	"wav/train/04_64958273.wav",
	//	"wav/train/05_65432978.wav",
	//}
	//
	//text := []string{
	//	"32468975",
	//	"58769423",
	//	"59682734",
	//	"64958273",
	//	"65432978",
	//}
	//
	//buffer := make([][]byte, 0)
	//
	//for _, file := range trainlist {
	//	buf, err := loadWaveData(file)
	//	if err != nil {
	//		log.Error(err)
	//		return
	//	}
	//	buffer = append(buffer, buf)
	//}

	verifData, err := loadWaveData("wav/verify/34986527.wav")
	if err != nil {
		log.Error(err)
		return
	}

	//vprEngine.TrainSpeech(5, buffer, text, "test")
	//log.Info("vpr TrainSpeech successful")
	vprEngine.RecSpeech(verifData, "34986527", "test")

}

func loadWaveData(file string) ([]byte, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	data = data[44:]
	return data, nil
}
