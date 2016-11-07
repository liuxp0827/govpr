package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/liuxp0827/govpr/httpapi/log"
)

func init() {
	// register model
	orm.RegisterModel(new(AppInfo))
}

// AppInfo 应用信息
type AppInfo struct {
	Id        int64      // Id号
	AppId     string     `orm:"unique;size(50)"`         //appid
	Key       string     `orm:"size(50)"`                //秘钥
	Name      string     `orm:"unique;size(32)"`         //app名字
	Developer *Developer `orm:"rel(fk)"`                 //应用所属的开发者
	Created   time.Time  `orm:"auto_now_add;type(date)"` //应用创建时间
	Email     string     `orm:"size(100)"`               //与应用绑定的Email
	Token     string     `orm:"size(100)"`               // app授权token值
	Users     []*User    `orm:"reverse(many)"`
	lock      *sync.Mutex
}

func (this *Developer) NewAppInfo(appname string) *AppInfo {

	t := time.Now()
	appid := fmt.Sprintf("%d", t.UnixNano())
	key := appGenKey(t, appname)

	app := &AppInfo{
		AppId:     appid,
		Key:       key,
		Name:      appname,
		Developer: this,
		Created:   t,
		Email:     this.Email,
		Users:     make([]*User, 0),
		Token:     appGenToken(appid, key),
		lock:      &sync.Mutex{},
	}

	return app
}

func (this *Developer) AddAppInfo(app *AppInfo) error {

	o := orm.NewOrm()
	var a AppInfo
	err := o.QueryTable("app_info").Filter("name", app.Name).One(&a)
	if err == nil {
		return fmt.Errorf("Developer %s AddAppInfo %s failed: appname is exist", this.DeveloperName, app.Name)
	}

	id, err := o.Insert(app)
	if err != nil {
		log.Errorf("Developer %s AddAppInfo %s: %v", this.DeveloperName, app.Name, err)
		return err
	}

	if app.lock == nil {
		app.lock = &sync.Mutex{}
	}

	if app.Users != nil && len(app.Users) > 0 {
		app.Id = id
		for _, c := range app.Users {
			app.AddUser(c)
		}
	}

	if this.Apps == nil {
		this.Apps = make([]*AppInfo, 0)
	}
	this.Apps = append(this.Apps, app)

	err = UpdateDeveloper(this)
	if err != nil {
		return err
	}

	return err
}

func (this *Developer) UpdateAppInfo(app *AppInfo) error {

	o := orm.NewOrm()
	_, err := o.Update(app)
	if err != nil {
		return err
	}

	if app.Users != nil && len(app.Users) > 0 {
		for _, v := range app.Users {
			v.App = app
			app.AddAndUpdateUser(v)
		}
	}

	var exist bool = false
	for i, a := range this.Apps {
		if a.Id == app.Id {
			exist = true
			this.Apps[i] = a
		}
	}

	if !exist {
		this.Apps = append(this.Apps, app)
	}

	err = UpdateDeveloper(this)
	if err != nil {
		return err
	}

	return nil
}

func (this *Developer) DeleteAppInfo(appname string) error {

	o := orm.NewOrm()
	_, err := o.QueryTable("app_info").Filter("name", appname).Delete()
	if err != nil {
		log.Errorf("Developer %s DeleteAppInfo %s failed: %v", this.DeveloperName, appname, err)
		return err
	}

	index := -1
	length := len(this.Apps)
	for i := 0; i < length; i++ {
		if this.Apps[i].Name == appname {
			index = i
			break
		}
	}
	if index >= 0 {
		this.Apps[index] = this.Apps[length-1]
		this.Apps = this.Apps[:length-1]
	}

	err = UpdateDeveloper(this)
	if err != nil {
		return err
	}

	return nil
}

// 获得项目
func (this *Developer) GetAppInfoByName(appname string) (*AppInfo, error) {
	var app AppInfo
	o := orm.NewOrm()
	err := o.QueryTable("app_info").Filter("name", appname).One(&app)

	if err == orm.ErrMultiRows {
		return nil, err
	}
	if err == orm.ErrNoRows {
		return nil, err
	}

	app.Developer = this
	app.lock = &sync.Mutex{}
	app.Users, _, err = app.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("Developer %s GetAppInfoByName %s failed: %v", this.DeveloperName, appname, err)
	}
	return &app, nil
}

func (this *Developer) GetAppInfoByToken(token string) (*AppInfo, error) {
	var app AppInfo
	o := orm.NewOrm()
	err := o.QueryTable("app_info").Filter("token", token).One(&app)

	if err == orm.ErrMultiRows {
		return nil, err
	}
	if err == orm.ErrNoRows {
		return nil, err
	}

	app.Developer = this
	app.lock = &sync.Mutex{}
	app.Users, _, err = app.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("Developer %s GetAppInfoByToken %s failed: %v", this.DeveloperName, token, err)
	}
	return &app, nil
}

func GetAppInfoByToken(token string) (*AppInfo, error) {
	var app AppInfo
	o := orm.NewOrm()
	err := o.QueryTable("app_info").Filter("token", token).One(&app)

	if err == orm.ErrMultiRows {
		return nil, err
	}
	if err == orm.ErrNoRows {
		return nil, err
	}

	app.lock = &sync.Mutex{}
	app.Users, _, err = app.GetAllUsers()
	if err != nil {
		log.Errorf("GetAppInfoByToken %s failed: %v", token, err)
		return nil, fmt.Errorf("GetAppInfoByToken %s failed: %v", token, err)
	}
	return &app, nil
}

func (this *Developer) hasPermissionForAppInfo(appid string, appkey string) (bool, *AppInfo, error) {
	var app AppInfo
	o := orm.NewOrm()
	err := o.QueryTable("app_info").Filter("app_id", appid).Filter("key", appkey).One(&app)
	if err != nil {
		return false, nil, err
	}

	if appkey != app.Key {
		return false, nil, nil
	}

	app.Developer = this
	app.lock = &sync.Mutex{}
	app.Users, _, err = app.GetAllUsers()
	if err != nil {
		return false, nil, fmt.Errorf("Developer %s hasPermissionForAppInfo appid: %s, appkey: %s failed: %v", this.DeveloperName, appid, appkey, err)
	}

	return true, &app, nil
}

// 获得项目
func (this *Developer) GetAppInfoById(id int64) (*AppInfo, error) {
	var app AppInfo
	o := orm.NewOrm()
	err := o.QueryTable("project").Filter("id", id).One(&app)
	if err == orm.ErrMultiRows {
		return nil, err
	}
	if err == orm.ErrNoRows {
		return nil, err
	}

	app.Developer = this
	app.lock = &sync.Mutex{}
	app.Users, _, err = app.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("Developer %s GetAppInfoById %s failed: %v", err)
	}

	return &app, nil
}

// 获得所有项目
func (this *Developer) GetAllAppInfo() (appInfos []*AppInfo, count int64, err error) {
	o := orm.NewOrm()
	count, err = o.QueryTable("app_info").Filter("developer_id", this.Id).All(&appInfos)
	if err != nil {
		log.Errorf("Developer %s GetAllAppInfo failed: %v", this.DeveloperName, err)
		return nil, -1, err
	}

	for _, p := range appInfos {
		p.lock = &sync.Mutex{}
		p.Developer = this
	}
	return
}

func appGenKey(t time.Time, appname string) string {
	return baseCoder.EncodeToString([]byte(strconv.FormatInt(t.UnixNano(), 10) + appname))
}

func appGenToken(appid, key string) string {
	str := fmt.Sprintf("%s&%s", appid, key)
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
