package controllers

import (
	"io/ioutil"
	"strconv"

	"github.com/liuxp0827/govpr/httpapi/constants"
	"github.com/astaxie/beego"
	"github.com/liuxp0827/govpr/log"
	"github.com/liuxp0827/govpr/httpapi/models"
)

// Operations about Users
type UserController struct {
	beego.Controller
}

// @Title registerUser
// @Description create Users
// @Success 200 {string, string, string} ret, errCode, msg
// @Failure 403 body is empty
// @router /registerUser [post]
func (this *UserController) RegisterUser() {

	userid := this.Input().Get("userid")
	token := this.Input().Get("token")

	if userid == "" {
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_REGISTER_USER, "errCode": constants.ERROR_USER_ILLEGAL,
			"msg": "get userid " + userid + " failed, userid is illegal"}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()

	err := db.AddUser(token, userid)
	if err != nil {
		if err.Error() == "token" {
			log.Warnf("用户账号[%s]: 添加用户失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_REGISTER_USER, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}
		log.Warnf("用户账号[%s]: 添加用户失败, 用户已存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_REGISTER_USER, "errCode": constants.ERROR_USER_EXISTENT, "msg": "register userid " + userid + " failed, " + err.Error()}
		this.ServeJSON(false)

	} else {
		log.Infof("用户账号[%s]: 添加用户成功", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_REGISTER_USER, "errCode": constants.SUCCESS_REGISTER_USER, "msg": "register userid " + userid + " success"}
		this.ServeJSON(false)
	}
}

// @Title detectVerify
// @Description detect User whether can do verify
// @Success 200 {string, string, string} ret, errCode, msg
// @Failure 403 body is empty
// @router /detectquery [post]
func (this *UserController) DetectVerify() {
	userid := this.Input().Get("userid")
	token := this.Input().Get("token")

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
			log.Warnf("用户账号[%s]: 验证检测失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DETECT_QUERY, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}
		// log.Errorf("用户账号[%s]: err %s", userid, err.Error())
		log.Warnf("用户账号[%s]: 验证检测失败, 用户不存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DETECT_QUERY, "errCode": constants.ERROR_USER_NONEXISTENT, "msg": "detect verify userid " + userid + " failed, " + err.Error()}
		this.ServeJSON(false)
		return
	}

	if u.IsTrain == false {
		log.Warnf("用户账号[%s]: 验证检测失败, 模型不存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DETECT_QUERY, "errCode": constants.ERROR_MODEL_NONEXISTENT, "msg": "detect verify userid " + userid + " failed, model is not exist"}
		this.ServeJSON(false)
	} else {
		log.Infof("用户账号[%s]: 验证检测通过", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_DETECT_QUERY, "errCode": constants.SUCCESS_DETECT_QUERY, "msg": "detect verify userid " + userid + " success"}
		this.ServeJSON(false)
	}
}

// @Title deleteUser
// @Description delete User
// @Success 200 {string, string, string} ret, errCode, msg
// @Failure 403 body is empty
// @router /deleteUser [post]
func (this *UserController) DeleteUser() {
	userid := this.Input().Get("userid")
	token := this.Input().Get("token")

	if userid == "" {
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DELETE_USER, "errCode": constants.ERROR_USER_ILLEGAL,
			"msg": "get userid " + userid + " failed, userid is illegal"}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	err := db.DeleteUser(token, userid)
	if err != nil {
		if err.Error() == "token" {
			log.Warnf("用户账号[%s]: 删除用户失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DELETE_USER, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}
		log.Warnf("用户账号[%s]: 删除用户失败, 用户不存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DELETE_USER, "errCode": constants.ERROR_USER_NONEXISTENT, "msg": "delete userid failed, " + err.Error()}
		this.ServeJSON(false)
	} else {
		log.Infof("用户账号[%s]: 删除用户成功", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_DELETE_USER, "errCode": constants.SUCCESS_DELETE_USER, "msg": "delete userid " + userid + " success"}
		this.ServeJSON(false)
	}

}

// @Title add Sample
// @Description add User's sample for train models
// @Success 200 {string, string, string} ret, errCode, msg
// @Failure 403 body is empty
// @router /addsample [post]
func (this *UserController) AddSample() {
	userid := this.Input().Get("userid")
	token := this.Input().Get("token")
	content := this.Input().Get("content")
	step, _ := strconv.Atoi(this.Input().Get("step"))

	if userid == "" {
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_ADDSAMPLE, "errCode": constants.ERROR_USER_ILLEGAL,
			"msg": "get userid " + userid + " failed, userid is illegal"}
		this.ServeJSON(false)
		return
	}

	if step > 5 || step < 1 {
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_ADDSAMPLE, "errCode": constants.ERROR_URL_PARAM_ILLEGAL,
			"msg": "get userid " + userid + " failed, url param 'step' must between 1 and 5"}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	u, err := db.GetUserById(token, userid)
	if err != nil {
		if err.Error() == "token" {
			log.Warnf("用户账号[%s]: 添加语音数据失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_ADDSAMPLE, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}

		log.Warnf("用户账号[%s]: 添加语音数据失败, 用户不存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_ADDSAMPLE, "errCode": constants.ERROR_USER_NONEXISTENT, "msg": "get userid " + userid + " failed, " + err.Error()}
		this.ServeJSON(false)
		return
	}

	if u.IsTrain == true {
		log.Warnf("用户账号[%s]: 添加语音数据失败, 模型已存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_ADDSAMPLE, "errCode": constants.ERROR_MODEL_EXISTENT, "msg": "the model of userid " + userid + " is existed."}
		this.ServeJSON(false)
		return
	}

	file, _, err := this.Ctx.Request.FormFile("file")
	if err != nil {
		log.Errorf("FormFile: %s", err.Error())
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("Close: %s", err.Error())
			return
		}
	}()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("ReadAll: %s", err.Error())
		return
	}

	var lengthOfData int
	if data == nil || len(data) <= 10000 {
		log.Errorf("用户账号[%s]: 添加语音数据失败, 语音数据为空", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_ADDSAMPLE, "errCode": constants.ERROR_SAMPLE_IS_NULL, "msg": "upload sample is null, please reupload."}
		this.ServeJSON(false)
		return
	}

	lengthOfData = len(data)

	err = db.AddWavesAndContents(token, userid, data, content, step)
	if err != nil {
		if err.Error() == "token" {
			log.Warnf("用户账号[%s]: 添加语音数据失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_ADDSAMPLE, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}

		log.Errorf("用户账号[%s]: 添加语音数据失败, 添加语音数据到数据库有误", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_ADDSAMPLE, "errCode": constants.ERROR_ADDSAMPLE_FAILED, "msg": "userid " + userid + "add sample failed, " + err.Error()}
		this.ServeJSON(false)
		return
	}

	log.Infof("用户账号[%s]: 训练文本内容: %s, 当前训练步骤: %d, 添加语音数据成功, 语音长度为: %d", userid, content, step, lengthOfData)
	this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_ADDSAMPLE, "errCode": constants.SUCCESS_ADDSAMPLE, "step": step, "msg": "step " + strconv.Itoa(step) + ": wav upload success"}
	this.ServeJSON(false)

}

// @Title clearSamples
// @Description clear User's samples
// @Success 200 {string, string, string} ret, errCode, msg
// @Failure 403 body is empty
// @router /clearSamples [post]
func (this *UserController) ClearSamples() {
	userid := this.Input().Get("userid")
	token := this.Input().Get("token")

	if userid == "" {
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_CLEAR_SAMPLES, "errCode": constants.ERROR_USER_ILLEGAL,
			"msg": "get userid " + userid + " failed, userid is illegal"}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	_, err := db.GetUserById(token, userid)
	if err != nil {
		if err.Error() == "token" {
			log.Warnf("用户账号[%s]: 删除用户语音数据失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_CLEAR_SAMPLES, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}

		log.Warnf("用户账号[%s]: 删除用户语音数据失败, 用户不存在", userid)
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_CLEAR_SAMPLES, "errCode": constants.ERROR_USER_NONEXISTENT, "msg": "get userid " + userid + " failed"}
		this.ServeJSON(false)
	} else {
		err = db.ClearWavesAndContents(token, userid)
		if err != nil {
			if err.Error() == "token" {
				log.Warnf("用户账号[%s]: 删除用户语音数据失败, 没有应用权限", userid)
				this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_CLEAR_SAMPLES, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
				this.ServeJSON(false)
				return
			}

			log.Errorf("用户账号[%s]: 删除用户语音数据失败, 删除语音数据有误", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_CLEAR_SAMPLES, "errCode": constants.ERROR_CLEAR_SAMPLES_FAILED, "msg": "clear userid " + userid + " waves and contents failed"}
			this.ServeJSON(false)
		} else {
			log.Infof("用户账号[%s]: 删除用户语音数据成功", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_CLEAR_SAMPLES, "errCode": constants.SUCCESS_CLEAR_SAMPLES, "msg": "clear userid " + userid + " success"}
			this.ServeJSON(false)
		}
	}
}

// @Title detectregister
// @Description detect User for add sample
// @Success 200 {string, string, string} ret, errCode, msg
// @Failure 403 body is empty
// @router /detectregister [post]
func (this *UserController) DetectRegister() {
	userid := this.Input().Get("userid")
	token := this.Input().Get("token")

	if userid == "" {
		this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DETECT_REGISTER, "errCode": constants.ERROR_USER_ILLEGAL,
			"msg": "get userid " + userid + " failed, userid is illegal"}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	p, err := db.GetUserById(token, userid)

	if err != nil {

		if err.Error() == "token" {
			log.Warnf("用户账号[%s]: 登记检测失败, 没有应用权限", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DETECT_REGISTER, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
			this.ServeJSON(false)
			return
		}

		log.Errorf("用户账号[%s]: 用户不存在: %s", userid, err.Error())

		err = db.AddUser(token, userid)
		if err != nil {
			if err.Error() == "token" {
				log.Warnf("用户账号[%s]: 添加用户失败, 没有应用权限", userid)
				this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_REGISTER_USER, "errCode": constants.ERROR_APP_TOKEN, "msg": "register userid " + userid + " failed, " + "app token error"}
				this.ServeJSON(false)
				return
			}

			log.Warnf("用户账号[%s]: 添加用户失败, 用户已存在,%s", userid, err.Error())
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_REGISTER_USER, "errCode": constants.ERROR_USER_EXISTENT, "msg": "register userid " + userid + " failed, " + err.Error()}
			this.ServeJSON(false)
		} else {
			log.Infof("用户账号[%s]: 添加用户成功", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_DETECT_REGISTER, "errCode": constants.SUCCESS_REGISTER_USER, "msg": "register userid " + userid + " success"}
			this.ServeJSON(false)
		}
	} else {
		if p.IsTrain == false {
			log.Infof("用户账号[%s]: 登记检测通过", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.SUCCESS_DETECT_REGISTER, "errCode": constants.SUCCESS_DETECT_REGISTER, "msg": "detect register userid " + userid + " success"}
			this.ServeJSON(false)
		} else {
			log.Warnf("用户账号[%s]: 登记检测失败, 模型已训练", userid)
			this.Data["json"] = map[string]interface{}{"ret": constants.FAILED_DETECT_REGISTER, "errCode": constants.ERROR_MODEL_EXISTENT, "msg": "detect register userid " + userid + " failed, model is exist"}
			this.ServeJSON(false)
		}
	}
}
