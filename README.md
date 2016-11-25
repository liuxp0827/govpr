
## 简介
govpr是golang 实现的基于 GMM-UBM 说话人识别引擎(声纹识别),可用于语音验证,身份识别的场景.
目前暂时仅支持汉语数字的语音,语音格式为wav格式(比特率16000,16bits,单声道)

## 安装

go get github.com/liuxp0827/govpr

## 示例

如下是一个简单的示例. 可跳转至 [example](https://github.com/liuxp0827/govpr/blob/master/example)
查看详细的例子,示例中的语音为纯数字8位数字.语音验证后得到一个得分,可设置阈值来判断验证语音是否为注册训练者本人.
示例中,预设阈值1.0,语音验证得分>=1.0,可认定为是本人语音,语音验证得分<1.0则非本人语音.
![得分](https://github.com/liuxp0827/govpr/blob/master/example/result.jpg)

```go
package main

import (
	"github.com/liuxp0827/govpr"
	"github.com/liuxp0827/govpr/log"
	"github.com/liuxp0827/govpr/waveIO"
	"io/ioutil"
)

type engine struct {
	vprEngine *govpr.VPREngine
}

func NewEngine(sampleRate, delSilRange int, ubmFile, userModelFile string) (*engine, error) {
	vprEngine, err := govpr.NewVPREngine(sampleRate, delSilRange, true, ubmFile, userModelFile)
	if err != nil {
		return nil, err
	}
	return &engine{vprEngine: vprEngine}, nil
}

func (this *engine) DestroyEngine() {
	this.vprEngine = nil
}

func (this *engine) TrainSpeech(buffers [][]byte) error {

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

func (this *engine) RecSpeech(buffer []byte) (float64, error) {

	err := this.vprEngine.AddVerifyBuffer(buffer)
	defer this.vprEngine.ClearVerifyBuffer()
	if err != nil {
		log.Error(err)
		return -1.0, err
	}

	err = this.vprEngine.VerifyModel()
	if err != nil {
		log.Error(err)
		return -1.0, err
	}

	return this.vprEngine.GetScore(), nil
}

func main() {
	log.SetLevel(log.LevelDebug)

	vprEngine, err := NewEngine(16000, 50, "../ubm/ubm", "model/test.dat")
	if err != nil {
		log.Fatal(err)
	}

	trainlist := []string{
		"wav/train/01_32468975.wav",
		"wav/train/02_58769423.wav",
		"wav/train/03_59682734.wav",
		"wav/train/04_64958273.wav",
		"wav/train/05_65432978.wav",
	}

	trainBuffer := make([][]byte, 0)

	for _, file := range trainlist {
		buf, err := loadWaveData(file)
		if err != nil {
			log.Error(err)
			return
		}
		trainBuffer = append(trainBuffer, buf)
	}

	err = vprEngine.TrainSpeech(trainBuffer)
	if err != nil {
		log.Fatal(err)
	}

	var threshold float64 = 1.0

	selfverifyBuffer, err := waveIO.WaveLoad("wav/verify/self_34986527.wav")
	if err != nil {
		log.Fatal(err)
	}

	self_score, err := vprEngine.RecSpeech(selfverifyBuffer)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("self score %f, pass? %v", self_score, self_score >= threshold)

	otherverifyBuffer, err := waveIO.WaveLoad("wav/verify/other_38974652.wav")
	if err != nil {
		log.Fatal(err)
	}

	other_score, err := vprEngine.RecSpeech(otherverifyBuffer)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("other score %f, pass? %v", other_score, other_score >= threshold)
}

func loadWaveData(file string) ([]byte, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	// remove .wav header info 44 bits
	data = data[44:]
	return data, nil
}
```
