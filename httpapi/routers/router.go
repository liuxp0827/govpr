package routers

import (
	"github.com/astaxie/beego"
	"github.com/liuxp0827/govpr/httpapi/controllers"
)

func init() {
	beego.Router("/trainmodel", &controllers.ModelController{}, "post:TrainModel")
	beego.Router("/verifymodel", &controllers.ModelController{}, "post:VerifyModel")
	beego.Router("/deletemodel", &controllers.ModelController{}, "post:DeleteModel")

	beego.Router("/registeruser", &controllers.UserController{}, "post:RegisterUser")
	beego.Router("/deleteuser", &controllers.UserController{}, "post:DeleteUser")
	beego.Router("/detectquery", &controllers.UserController{}, "post:DetectVerify")
	beego.Router("/addsample", &controllers.UserController{}, "post:AddSample")
	beego.Router("/clearsamples", &controllers.UserController{}, "post:ClearSamples")
	beego.Router("/detectregister", &controllers.UserController{}, "post:DetectRegister")

	beego.Router("/registerdeveloper", &controllers.DeveloperController{}, "get,post:RegisterDeveloper")
	beego.Router("/deletedeveloper", &controllers.DeveloperController{}, "get,post:DeleteDeveloper")

	beego.Router("/registerapp", &controllers.AppInfoController{}, "get,post:RegisterApp")
	beego.Router("/deleteapp", &controllers.AppInfoController{}, "get,post:DeleteApp")
	beego.Router("/appinfo", &controllers.AppInfoController{}, "get,post:GetAppInfo")
}
