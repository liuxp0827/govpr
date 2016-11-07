package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/liuxp0827/govpr/httpapi/models"
)

type AppInfoController struct {
	beego.Controller
}

func (this *AppInfoController) RegisterApp() {
	q := this.Ctx.Request.URL.Query()
	name := q.Get("devname")
	pwd := q.Get("password")
	appname := q.Get("appname")

	if len(name) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register App name can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	if len(pwd) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register App password can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	if len(appname) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register App appname can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	b, err := db.CheckDeveloper(name, pwd)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register App failed: %v", err)}
		this.ServeJSON(false)
		return
	}
	if !b {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register App failed: Developer %s is not exist, or password is wrong", name)}
		this.ServeJSON(false)
		return
	}

	err = db.AddAppInfo(name, appname)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register App %s failed: %v", appname, err)}
		this.ServeJSON(false)
		return
	}

	app, err := db.GetAppInfoByName(name, appname)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register App %s failed: %v", appname, err)}
		this.ServeJSON(false)
		return
	}

	this.Data["json"] = map[string]interface{}{
		"msg":    fmt.Sprintf("Register App successfully"),
		"appid":  app.AppId,
		"appkey": app.Key,
	}

	this.ServeJSON(false)
	return
}

func (this *AppInfoController) DeleteApp() {
	q := this.Ctx.Request.URL.Query()
	name := q.Get("devname")
	pwd := q.Get("password")
	appname := q.Get("appname")

	if len(name) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete App name can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	if len(pwd) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete App password can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	if len(appname) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete App appname can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	b, err := db.CheckDeveloper(name, pwd)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete App failed: %v", err)}
		this.ServeJSON(false)
		return
	}
	if !b {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete App failed: Developer %s is not exist, or password is wrong", name)}
		this.ServeJSON(false)
		return
	}

	err = db.DeleteAppInfoByName(name, appname)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete App failed: %v", err)}
		this.ServeJSON(false)
		return
	}

	this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete App successfully")}
	this.ServeJSON(false)
	return

}

func (this *AppInfoController) GetAppInfo() {
	q := this.Ctx.Request.URL.Query()
	name := q.Get("devname")
	pwd := q.Get("password")
	appname := q.Get("appname")

	if len(name) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Get App Info name can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	if len(pwd) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Get App Info password can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	if len(appname) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register App appname can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	b, err := db.CheckDeveloper(name, pwd)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Get App Info failed: %v", err)}
		this.ServeJSON(false)
		return
	}
	if !b {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Get App Info failed: Developer %s is not exist, or password is wrong", name)}
		this.ServeJSON(false)
		return
	}

	app, err := db.GetAppInfoByName(name, appname)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Get App Info %s failed: %v", appname, err)}
		this.ServeJSON(false)
		return
	}

	this.Data["json"] = map[string]interface{}{
		"msg":    fmt.Sprintf("Get App Info successfully"),
		"appid":  app.AppId,
		"appkey": app.Key,
	}

	this.ServeJSON(false)
	return
}
