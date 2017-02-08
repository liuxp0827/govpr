package controllers

import (
	"fmt"
	"io/ioutil"

	"github.com/astaxie/beego"
	"github.com/liuxp0827/govpr/httpapi/constants"
	"github.com/liuxp0827/govpr/httpapi/engine"
	"github.com/liuxp0827/govpr/httpapi/models"
	"github.com/liuxp0827/govpr/log"
)

var model_dir string = beego.AppConfig.DefaultString("model_dir", "mod/")

// Operations about Models
type ModelController struct {
	beego.Controller
}

// @Title trainModel
// @Description train User's model
// @Success 200 {string, string, string} ret, errCode, msg
// @Failure 403 body is empty
// @router /trainModel [post]
func (this *ModelController) TrainModel() {
	userid := this.Input().Get("userid")
	token := this.Input().Get("token")

	if userid == "" {
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DETECT_QUERY, "errCode": constants.ERROR_USER_ILLEGAL,
			"msg": "get userid " + userid + " failed, userid is illegal"}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()

	usr, err := db.GetUserByIdForTrain(token, userid)

	if err != nil {
		if err.Error() == "token" {
			log.Warnf("用户账号[%s]: 训练自适应模型失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_TRAIN_MODEL, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}

		log.Warnf("用户账号[%s] GetUserById failed: %v", userid, err)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_TRAIN_MODEL, "errCode": constants.ERROR_USER_NONEXISTENT, "msg": "train model failed, get  userid " + userid + " failed, " + err.Error()}
		this.ServeJSON(false)
		return
	}

	if usr.IsTrain {
		log.Warnf("用户账号[%s]: 训练自适应模型失败, 模型已存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_TRAIN_MODEL, "errCode": constants.ERROR_MODEL_EXISTENT, "msg": "train model failed, the model has existed."}
		this.ServeJSON(false)
		return
	}

	lengths := len(usr.Waves)
	if lengths < 5 {
		log.Errorf("用户账号[%s]: 训练自适应模型失败, 训练数据不足", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_TRAIN_MODEL, "errCode": constants.ERROR_SAMPLES_NOT_ENOUGH, "msg": "train userid " + userid + " model failed, count of train data is not enough, count must be greater than 5."}
		this.ServeJSON(false)
		return
	}

	for i := 0; i < lengths; i++ {
		if usr.Waves[i] == nil || len(usr.Waves[i]) <= 5000 {
			log.Errorf("用户账号[%s]: 训练自适应模型失败, 第%d条训练数据不足", userid, i+1)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_TRAIN_MODEL, "errCode": constants.ERROR_SAMPLES_NOT_ENOUGH, "msg": "train userid " + userid + " model failed, train data[" + fmt.Sprintf("%d", i+1) + "] is not enough."}
			this.ServeJSON(false)
			return
		}
	}

	x, err := engine.NewEngine(16000, 50, model_dir+token+"_"+userid+"/"+userid+".dat")
	if err != nil {
		log.Errorf("用户账号[%s]: 训练自适应模型失败, 训练过程有误, %v", userid, err)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_TRAIN_MODEL, "errCode": constants.ERROR_TRAIN_MODEL_FAILED, "msg": fmt.Sprintf("train userid %s model failed: %v", userid, err)}
		this.ServeJSON(false)
		return
	}

	err = x.TrainSpeech(lengths, usr.Waves, usr.Contents, usr.UserId, usr.Token)
	x.DestroyEngine()

	if err != nil {
		log.Errorf("用户账号[%s]: 训练自适应模型失败, 训练过程有误, %v", userid, err)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_TRAIN_MODEL, "errCode": constants.ERROR_TRAIN_MODEL_FAILED, "msg": fmt.Sprintf("train userid %s model failed: %v", userid, err)}
		this.ServeJSON(false)
		return
	}

	err = db.UpdateIsTrained(token, userid, true)
	if err != nil {
		log.Errorf("用户账号[%s]: 训练自适应模型失败,更新数据库失败", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_TRAIN_MODEL, "errCode": constants.ERROR_TRAIN_MODEL_FAILED, "msg": "train userid " + userid + " model failed, update database failed, " + err.Error()}
		this.ServeJSON(false)
		return
	}
	log.Infof("用户账号[%s]: 训练自适应模型成功", userid)
	this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_TRAIN_MODEL, "errCode": constants.SUCCESS_TRAIN_MODEL, "msg": "userid " + userid + " train model success."}
	this.ServeJSON(false)
	return

}

// @Title deleteModel
// @Description delete User's model
// @Success 200 {string, string, string} ret, errCode, msg
// @Failure 403 body is empty
// @router /deleteModel [post]
func (this *ModelController) DeleteModel() {
	userid := this.Input().Get("userid")
	token := this.Input().Get("token")

	if userid == "" {
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DETECT_QUERY, "errCode": constants.ERROR_USER_ILLEGAL,
			"msg": "get userid " + userid + " failed, userid is illegal"}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	_, err := db.GetUserById(token, userid)
	if err != nil {
		if err.Error() == "token" {
			log.Warnf("用户账号[%s]: 删除自适应模型失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DELETE_MODEL, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}

		log.Warnf("用户账号[%s]: 删除自适应模型失败, 用户不存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DELETE_MODEL, "errCode": constants.ERROR_USER_NONEXISTENT, "msg": "train model failed, get  userid " + userid + " failed, " + err.Error()}
		this.ServeJSON(false)
		return
	}

	db.UpdateIsTrained(token, userid, false)
	log.Infof("用户账号[%s]: 删除自适应模型成功", userid)
	this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_DELETE_MODEL, "errCode": constants.SUCCESS_DELETE_MODEL, "msg": "userid " + userid + " delete model success."}
	this.ServeJSON(false)
	return

}

// @Title verifyModel
// @Description verify User's model
// @Success 200 {string, string, string} ret, errCode, msg
// @Failure 403 body is empty
// @router /verifyModel [post]
func (this *ModelController) VerifyModel() {

	userid := this.Input().Get("userid")
	token := this.Input().Get("token")
	content := this.Input().Get("content")

	if userid == "" {
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DETECT_QUERY, "errCode": constants.ERROR_USER_ILLEGAL,
			"msg": "get userid " + userid + " failed, userid is illegal"}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	u, err := db.GetUserById(token, userid)
	if err != nil {
		if err.Error() == "token" {
			log.Warnf("用户账号[%s]: 验证语音数据失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_VERIFY_MODEL, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}
		log.Warnf("用户账号[%s]: 验证语音数据失败, 用户不存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_VERIFY_MODEL, "errCode": constants.ERROR_USER_NONEXISTENT, "msg": "verify model failed, get userid " + userid + " failed, " + err.Error()}
		this.ServeJSON(false)
		return
	}

	file, _, err := this.Ctx.Request.FormFile("file")
	if err != nil {
		log.Errorf("FormFile: %s", err.Error())
		return
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("ReadAll: %s", err.Error())
		return
	}

	if data == nil || len(data) <= 0 {
		log.Errorf("用户账号[%s]: 验证语音数据失败, 语音数据为空", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_VERIFY_MODEL, "errCode": constants.ERROR_SAMPLE_IS_NULL, "msg": "upload sample is null, please reupload."}
		this.ServeJSON(false)
		return
	}

	x, err := engine.NewEngine(16000, 50, model_dir+token+"_"+userid+"/"+userid+".dat")
	if err != nil {
		log.Errorf("用户账号[%s]: 验证语音数据失败, 验证过程有误, %v", userid, err)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_VERIFY_MODEL, "errCode": constants.ERROR_VERIFY_MODEL_FAILED, "msg": fmt.Sprintf("verify userid %s failed: %v", userid, err)}
		this.ServeJSON(false)
		return
	}

	score, err := x.RecSpeech(data, content, u.UserId, u.Token)
	x.DestroyEngine()
	if err != nil {
		log.Errorf("用户账号[%s]: 验证语音数据失败, 验证过程有误, %v", userid, err)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_VERIFY_MODEL, "errCode": constants.ERROR_VERIFY_MODEL_FAILED, "msg": fmt.Sprintf("verify userid %s failed: %v", userid, err)}
		this.ServeJSON(false)
		return
	}

	log.Infof("用户账号[%s]: 验证口令: %s, 最终得分: %f", userid, content, score)
	this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_VERIFY_MODEL, "score": score, "errCode": constants.SUCCESS_VERIFY_MODEL, "msg": "verify userid " + userid + " success."}
	this.ServeJSON(false)

	return

}
