package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/liuxp0827/govpr/log"
)

func init() {
	// register model
	orm.RegisterModel(new(User))
}

type User struct {
	Id       int64
	App      *AppInfo `orm:"rel(fk)"`
	UserId   string   `orm:"unique;size(32)"` // 用户ID号
	Token    string   `orm:"size(100)"`       // app授权token值
	Waves    [][]byte `orm:"-"`
	Contents []string `orm:"-"` // 5条语音的文本内容
	IsTrain  bool     // 是否已经训练模型
	lock     *sync.Mutex
}

func (this *AppInfo) NewUser(token, userid string) *User {

	user := User{
		Token:    token,
		App:      this,
		UserId:   userid,
		Contents: make([]string, 5, 5),
		Waves:    make([][]byte, 5, 5),
		IsTrain:  false,
		lock:     &sync.Mutex{},
	}
	return &user
}

func (this *AppInfo) AddUser(user *User) error {
	o := orm.NewOrm()
	var u User
	err := o.QueryTable("user").Filter("user_id", user.UserId).Filter("token", user.Token).One(&u)
	if err == nil {
		return fmt.Errorf("AppInfo %s AddUser %s failed: userid is exist", this.Name, user.UserId)
	}

	_, err = o.Insert(user)
	return err
}

func (this *AppInfo) DeleteUser(userid, token string) error {
	o := orm.NewOrm()
	var u User
	err := o.QueryTable("user").Filter("user_id", userid).Filter("token", token).One(&u)
	if err != nil {
		log.Debugf("AppInfo %s DeleteUser %s is not exist, token %s", this.Name, userid, token)
		return nil
	}

	id, err := o.QueryTable("user").Filter("user_id", userid).Filter("token", token).Delete()
	if err != nil {
		return err
	}

	log.Debugf("AppInfo %s DeleteUser %s successful, id %d", this.Name, userid, id)

	train_data_path := beego.AppConfig.DefaultString("model_path", "mod/") + token + "_" + userid
	fileInfos, err := ioutil.ReadDir(train_data_path)
	if err != nil {
		log.Errorf("AppInfo %s DeleteUser %s failed: %v", this.Name, userid, err)
		return nil
	}

	for _, v := range fileInfos {
		if !v.IsDir() && strings.HasPrefix(v.Name(), "_0") {
			os.Remove(train_data_path + "/" + v.Name())
		}
	}

	return nil
}

func (this *AppInfo) GetUserByIdForTrain(userid, token string) (*User, error) {
	o := orm.NewOrm()
	var u User
	err := o.QueryTable("user").Filter("user_id", userid).Filter("token", token).One(&u)
	if err != nil {
		return nil, err
	}

	err = u.GetAllWavesAndContents()
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (this *AppInfo) GetUserById(userid, token string) (*User, error) {
	o := orm.NewOrm()
	var u User
	err := o.QueryTable("user").Filter("user_id", userid).Filter("token", token).One(&u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (this *AppInfo) AddAndUpdateUser(user *User) error {
	o := orm.NewOrm()
	var u User
	err := o.QueryTable("user").Filter("user_id", user.UserId).Filter("token", user.Token).One(&u)
	if err == nil {
		_, err := o.Update(u)
		if err != nil {
			log.Errorf("AppInfo %s AddAndUpdateUser: %v", this.Name, user.UserId, err)
			return err
		}
	}

	_, err = o.Insert(user)
	return err
}

func (this *AppInfo) UpdateUser(u *User) error {
	o := orm.NewOrm()
	_, err := o.Update(u)
	if err != nil {
		log.Errorf("AppInfo %s UpdateUser: %v", this.Name, u.UserId, err)
		return err
	}
	return nil
}

func (this *AppInfo) GetAllUsers() (users []*User, count int64, err error) {
	o := orm.NewOrm()
	count, err = o.QueryTable("user").Filter("app_id", this.Id).All(&users)
	if err != nil {
		log.Errorf("AppInfo %s GetAllUsers failed: %v", this.Name, err)
		return nil, -1, err
	}
	return
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
