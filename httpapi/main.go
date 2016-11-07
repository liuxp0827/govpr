package main

import (
	"github.com/astaxie/beego"
	"github.com/liuxp0827/govpr/httpapi/models"
	_ "github.com/liuxp0827/govpr/httpapi/routers"
	"github.com/liuxp0827/govpr/log"
)

func main() {

	mysqlUser := beego.AppConfig.String("mysql_user")
	mysqlPwd := beego.AppConfig.String("mysql_password")
	mysqlAddr := beego.AppConfig.String("mysql_addr")
	mysqlDB := beego.AppConfig.String("mysql_database")

	log.Infof("MySQL User Name: %s", mysqlUser)
	log.Infof("MySQL Addr: %s", mysqlAddr)
	log.Infof("MySQL Database: %s", mysqlDB)

	err := models.InitMysql(mysqlUser, mysqlPwd, mysqlAddr, mysqlDB)
	if err != nil {
		log.Fatal(err)
	}

	models.InitUserCache(beego.AppConfig.DefaultInt("local_cache_max_size", 500))
	beego.Run()
}
