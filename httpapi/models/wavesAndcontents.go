package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/astaxie/beego"
)

func (this *User) GetAllWavesAndContents() error {
	this.Waves = this.Waves[:0]
	this.Contents = this.Contents[:0]
	train_data_path := beego.AppConfig.DefaultString("model_path", "mod/") +
		this.Token + "_" + this.UserId

	fileInfos, err := ioutil.ReadDir(train_data_path)
	if err != nil {
		return err
	}

	for _, v := range fileInfos {
		if !v.IsDir() && strings.HasPrefix(v.Name(), "_0") {
			data, err := ioutil.ReadFile(train_data_path + "/" + v.Name())
			if err != nil {
				return err
			}
			this.Waves = append(this.Waves, data)
			this.Contents = append(this.Contents, v.Name()[4:])
		}
	}
	return nil
}

func (this *User) addWavesAndContents(wave []byte, content string, step int) error {

	if wave == nil || len(wave) <= 0 {
		return fmt.Errorf(" user %s length of upload wave is 0, please reupload", this.UserId)
	}

	// 训练语音绝对路径
	train_data := beego.AppConfig.DefaultString("model_path", "mod/") +
		this.Token + "_" + this.UserId + "/" + "_0" + fmt.Sprintf("%d", step) + "_" + content

	// 训练语音父目录
	train_data_path := path.Dir(train_data)

	if !IsExist(train_data) {

		// 创建父路径
		err := os.MkdirAll(train_data_path, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		os.Remove(train_data)
	}

	// ***************** 若之前第 step 条 语音已经存在, 则先删除，后再保存 ***************************
	fileInfos, err := ioutil.ReadDir(train_data_path)
	if err != nil {
		return err
	}
	for _, v := range fileInfos {
		if !v.IsDir() && strings.HasPrefix(v.Name(), "_0"+fmt.Sprintf("%d", step)+"_") {
			os.Remove(train_data_path + "/" + v.Name())
		}
	}
	// ***************** 若之前第 step 条 语音已经存在, 则先删除，后再保存 ***************************

	err = ioutil.WriteFile(train_data, wave, 0666)
	if err != nil {
		return err
	}

	if this.Contents == nil {
		this.Contents = make([]string, 5, 5)
	}

	this.Contents[step-1] = content

	return nil
}

func (this *User) clearWavesAndContents() error {

	train_data_path := beego.AppConfig.DefaultString("model_path", "mod/") +
		this.Token + "_" + this.UserId

	fileInfos, err := ioutil.ReadDir(train_data_path)
	if err != nil {
		return err
	}

	for _, v := range fileInfos {
		if !v.IsDir() && strings.HasPrefix(v.Name(), "_0") {
			os.Remove(train_data_path + "/" + v.Name())
		}
	}

	this.Waves = make([][]byte, 5, 5)
	this.Contents = make([]string, 5, 5)
	return nil
}
