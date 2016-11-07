package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/liuxp0827/govpr/httpapi/models"
)

type DeveloperController struct {
	beego.Controller
}

func (this *DeveloperController) RegisterDeveloper() {
	q := this.Ctx.Request.URL.Query()
	name := q.Get("name")
	pwd := q.Get("password")
	email := q.Get("email")

	if len(name) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register Developer name can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	if len(pwd) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register Developer password can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	if len(email) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register Developer email can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	err := db.AddDeveloper(name, pwd, email)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register Developer failed: %v", err)}
		this.ServeJSON(false)
		return
	}

	this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register Developer successfully")}
	this.ServeJSON(false)
	return
}

func (this *DeveloperController) DeleteDeveloper() {
	q := this.Ctx.Request.URL.Query()
	name := q.Get("name")
	pwd := q.Get("password")

	if len(name) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register Developer name can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	if len(pwd) == 0 {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Register Developer password can not be \"\"")}
		this.ServeJSON(false)
		return
	}

	db := models.NewDBEngine()
	b, err := db.CheckDeveloper(name, pwd)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete Developer failed: %v", err)}
		this.ServeJSON(false)
		return
	}
	if !b {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete Developer failed: %s is not exist, or password is wrong", name)}
		this.ServeJSON(false)
		return
	}

	err = db.DeleteDeveloperByName(name)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete Developer failed: %v", err)}
		this.ServeJSON(false)
		return
	}
	this.Data["json"] = map[string]interface{}{"msg": fmt.Sprintf("Delete Developer successfully")}
	this.ServeJSON(false)
	return
}
