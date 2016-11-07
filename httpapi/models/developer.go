package models

import (
	"errors"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/liuxp0827/govpr/log"
)

func init() {
	// register model
	orm.RegisterModel(new(Developer))
}

type Developer struct {
	Id            int64
	DeveloperName string     `orm:"unique;size(32)"`
	Password      string     `orm:"size(50)"`
	Email         string     `orm:"size(50)"`
	CreateTime    time.Time  `orm:"auto_now_add;type(date)"`
	Apps          []*AppInfo `orm:"reverse(many)"`
	lock          *sync.Mutex
}

func NewDeveloper(developername, password, email string) *Developer {
	t := time.Now()
	dev := &Developer{
		DeveloperName: developername,
		Password:      encryptBase64(password),
		Email:         email,
		CreateTime:    t,
		Apps:          make([]*AppInfo, 0),
		lock:          &sync.Mutex{},
	}
	return dev
}

func AddDeveloper(d *Developer) error {
	o := orm.NewOrm()
	_, err := o.Insert(d)
	if err != nil {
		log.Errorf("AddDeveloper %s failed: %v", d.DeveloperName, err)
		return err
	}

	if d.lock == nil {
		d.lock = &sync.Mutex{}
	}

	return err
}

func GetDeveloperByName(devname string) (*Developer, error) {
	var dev Developer
	o := orm.NewOrm()
	err := o.QueryTable("developer").Filter("developer_name", devname).One(&dev)
	if err != nil {
		return nil, err
	}

	dev.lock = &sync.Mutex{}
	dev.Apps, _, err = dev.GetAllAppInfo()
	if err != nil {
		log.Errorf("%s GetDeveloperByName: %v", devname, err)
	}

	return &dev, nil
}

func DeleteDeveloperByName(devname string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("developer").Filter("developer_name", devname).Delete()
	if err != nil {
		return err
	}
	return nil
}

func UpdateDeveloper(p *Developer) error {
	if p.lock == nil {
		p.lock = &sync.Mutex{}
	}

	o := orm.NewOrm()

	_, err := o.Update(p)
	if err != nil {
		log.Errorf("Developer %s UpdateDeveloper: %v", p.DeveloperName, err)
		return err
	}

	return nil
}

func GetAllDevelopers() (developers []*Developer, count int64, err error) {

	o := orm.NewOrm()
	count, err = o.QueryTable("developer").All(&developers)
	if err != nil {
		log.Error("GetAllDevelopers:", err)
		return nil, -1, err
	}

	for _, d := range developers {
		d.lock = &sync.Mutex{}
		d.Apps, _, err = d.GetAllAppInfo()
		if err != nil {
			log.Error("GetAllDevelopers:", err)
			return nil, -1, err
		}
	}
	return
}

func checkDeveloper(developername, password string) (bool, error) {
	dev, err := GetDeveloperByName(developername)
	if dev == nil {
		return false, err
	}
	if encryptBase64(password) != dev.Password {
		return false, nil
	}
	return true, nil
}

func checkEmailIsExist(email string) bool {
	var dev Developer
	o := orm.NewOrm()
	err := o.QueryTable("developer").Filter("email", email).One(&dev)
	if err == nil {
		return true
	}
	return false
}

func (developer *Developer) updatePassword(oldPassword, newPassword string) (bool, error) {
	if developer.Password != oldPassword {
		return false, errors.New("The password is incorrect.")
	}
	developer.Password = encryptBase64(newPassword)
	return true, UpdateDeveloper(developer)
}

func encryptBase64(password string) string {
	return baseCoder.EncodeToString([]byte(password))
}
