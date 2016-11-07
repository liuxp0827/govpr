package models

import (
	"encoding/base64"
	"fmt"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/liuxp0827/govpr/log"
)

const (
	base64Table = "123QRSTUabcdVWXYZHijKLAWDCABDstEFGuvwxyzGHIJklmnopqr234560178912"
)

var (
	baseCoder *base64.Encoding
)

func init() {
	baseCoder = base64.NewEncoding(base64Table)
}

func InitMysql(user, pwd, addr, db string) error {

	// dsn root:123456@tcp(103.27.5.136:3306)/github.com/liuxp0827/govpr/httpapi?charset=utf8
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", user, pwd, addr, db)
	//dsn := "root:123456@tcp(103.27.5.136:3306)/github.com/liuxp0827/govpr/httpapi?charset=utf8"
	log.Infof("InitMysql DSN: %s", dsn)
	err := orm.RegisterDataBase("default", "mysql", dsn, 30, 100)
	if err != nil {
		return err
	}

	// create table
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		return err
	}
	log.Infof("Synchronize DataBase success ...")
	return nil
}

type DBEngine struct {
}

func NewDBEngine() *DBEngine {
	return &DBEngine{}
}

///////////////////////////////////////////////////////////////////
////////////////////////   Developer    ///////////////////////////
///////////////////////////////////////////////////////////////////

func (this *DBEngine) AddDeveloper(devname, password, email string) error {
	dev := NewDeveloper(devname, password, email)
	return AddDeveloper(dev)
}

func (this *DBEngine) CheckDeveloper(devname, password string) (bool, error) {
	return checkDeveloper(devname, password)
}

func (this *DBEngine) FindDeveloperByName(devname string) (*Developer, error) {
	return GetDeveloperByName(devname)
}

func (this *DBEngine) UpdateDeveloperByName(dev *Developer) error {
	return UpdateDeveloper(dev)
}

func (this *DBEngine) DeleteDeveloperByName(devname string) error {
	_, err := GetDeveloperByName(devname)
	if err != nil {
		return err
	}
	return DeleteDeveloperByName(devname)
}

func (this *DBEngine) UpdateDeveloperPassword(devname, oldPassword, newPassword string) (bool, error) {
	u, err := GetDeveloperByName(devname)
	if err != nil {
		return false, err
	}

	return u.updatePassword(oldPassword, newPassword)
}

func (this *DBEngine) CheckEmailIsExist(email string) bool {
	return checkEmailIsExist(email)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////     AppInfo    ////////////////////////////
//////////////////////////////////////////////////////////////////////

func (this *DBEngine) AddAppInfo(devname, appName string) error {
	developer, err := GetDeveloperByName(devname)
	if err != nil {
		return err
	}

	app := developer.NewAppInfo(appName)

	err = developer.AddAppInfo(app)
	return err
}

func (this *DBEngine) HasPermissionForAppInfo(devname, appid, appkey string) (bool, *AppInfo, error) {
	developer, err := GetDeveloperByName(devname)
	if err != nil || developer == nil {
		return false, nil, err
	}

	return developer.hasPermissionForAppInfo(appid, appkey)
}

func (this *DBEngine) GetAppInfoByName(devname, appname string) (*AppInfo, error) {
	developer, err := GetDeveloperByName(devname)
	if err != nil {
		return nil, err
	}
	return developer.GetAppInfoByName(appname)
}

func (this *DBEngine) GetAppInfoByToken(devname, token string) (*AppInfo, error) {
	app, err := GetAppInfoByToken(token)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (this *DBEngine) UpdateAppInfoByName(devname string, app *AppInfo) error {
	developer, err := GetDeveloperByName(devname)
	if err != nil {
		return err
	}

	return developer.UpdateAppInfo(app)
}

func (this *DBEngine) DeleteAppInfoByName(devname, appname string) error {
	developer, err := GetDeveloperByName(devname)
	if err != nil {
		return err
	}

	return developer.DeleteAppInfo(appname)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////    User    ////////////////////////////
//////////////////////////////////////////////////////////////////////

func (this *DBEngine) AddUser(token, id string) error {
	app, err := GetAppInfoByToken(token)
	if app == nil || err != nil {
		return fmt.Errorf("token")
	}

	user := app.NewUser(app.Token, id)
	return app.AddUser(user)
}

func (this *DBEngine) DeleteUser(token, id string) error {
	app, err := GetAppInfoByToken(token)
	if app == nil || err != nil {
		return fmt.Errorf("token")
	}
	err = app.DeleteUser(id, token)
	if err == nil {
		UserCache.Remove(fmt.Sprintf("%s#%s", token, id))
	}
	return err
}

func (this *DBEngine) GetUserById(token, id string) (*User, error) {

	u, ok := UserCache.Get(fmt.Sprintf("%s#%s", token, id))
	if !ok {
		app, err := GetAppInfoByToken(token)
		if app == nil || err != nil {
			return nil, fmt.Errorf("token")
		}

		u, err = app.GetUserById(id, token)
		if err != nil {
			return nil, err
		} else {
			if u != nil {
				UserCache.Add(fmt.Sprintf("%s#%s", token, id), u)
			}
		}
	}

	if u != nil {
		return u, nil
	}

	return nil, fmt.Errorf("user is nil")
}

func (this *DBEngine) GetUserByIdForTrain(token, id string) (*User, error) {
	app, err := GetAppInfoByToken(token)
	if app == nil || err != nil {
		return nil, fmt.Errorf("token")
	}
	return app.GetUserByIdForTrain(id, token)
}

func (this *DBEngine) UpdateIsTrained(token, id string, isTrain bool) error {
	app, err := GetAppInfoByToken(token)
	if app == nil || err != nil {
		return fmt.Errorf("token")
	}

	user, err := app.GetUserById(id, token)
	if user == nil || err != nil {
		return err
	}

	user.IsTrain = isTrain

	err = app.UpdateUser(user)
	if err == nil {
		UserCache.Modify(fmt.Sprintf("%s#%s", token, id), user)
	}
	return err
}

func (this *DBEngine) AddWavesAndContents(token, id string, wave []byte, content string, step int) error {
	user, ok := UserCache.Get(fmt.Sprintf("%s#%s", token, id))
	if !ok {
		app, err := GetAppInfoByToken(token)
		if app == nil || err != nil {
			return fmt.Errorf("token")
		}

		user, err = app.GetUserById(id, token)
		if err != nil {
			return err
		} else {
			if user != nil {
				UserCache.Add(fmt.Sprintf("%s#%s", token, id), user)
			}
		}
	}

	if user == nil {
		return fmt.Errorf("user is nil")
	}

	return user.addWavesAndContents(wave, content, step)
}

func (this *DBEngine) ClearWavesAndContents(token, id string) error {
	user, ok := UserCache.Get(fmt.Sprintf("%s#%s", token, id))
	if !ok {
		app, err := GetAppInfoByToken(token)
		if app == nil || err != nil {
			return fmt.Errorf("token")
		}

		user, err := app.GetUserById(id, token)
		if err != nil {
			return err
		} else {
			if user != nil {
				UserCache.Add(fmt.Sprintf("%s#%s", token, id), user)
			}
		}
	}

	if user == nil {
		return fmt.Errorf("user is nil")
	}

	return user.clearWavesAndContents()
}
