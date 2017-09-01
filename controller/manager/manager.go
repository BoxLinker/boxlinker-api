package manager

import (
	"github.com/go-xorm/xorm"
	"github.com/BoxLinker/boxlinker-api/controller/models"

	"fmt"
	mAuth "github.com/BoxLinker/boxlinker-api/auth"
	"errors"
	settings "github.com/BoxLinker/boxlinker-api/settings/user"
	"github.com/BoxLinker/boxlinker-api"
	log "github.com/Sirupsen/logrus"
	"github.com/BoxLinker/boxlinker-api/controller/amqp"
)

type Manager interface {
	VerifyAuthToken(token string) (map[string]interface{}, error)
	VerifyUsernamePassword(username, password, hash string) (bool, error)
	GenerateToken(uid string, username string, exp ...int64) (string, error)

	// user
	CheckAdminUser() error
	GetUserByName(username string) (*models.User)
	GetUserById(id string) (*models.User)
	GetUsers(pageConfig boxlinker.PageConfig) ([]*models.User, error)
	SaveUser(user *models.User) error

	SaveUserToBeConfirmed(user *models.UserToBeConfirmed) error
	DeleteUserToBeConfirmed(uid string) error
	DeleteUsersToBeConfirmedByName(username string) error
	GetUserToBeConfirmed(id string, username string) (*models.UserToBeConfirmed, error)

	IsUserExists(username string) (bool, error)
	IsEmailExists(email string) (bool, error)
	UpdatePassword(id string, password string) (bool, error)

}

type DefaultManager struct {
	authenticator mAuth.Authenticator
	engine *xorm.Engine
	producer *amqp.Producer
}

type ManagerOptions struct {
	Authenticator mAuth.Authenticator
	DBUser string
	DBPassword string
	DBName string
	DBHost string
	DBPort int
}

func NewManager(config ManagerOptions) (Manager, error) {

	dbOptions := models.DBOptions{
		User: config.DBUser,
		Password: config.DBPassword,
		Name: config.DBName,
		Host: config.DBHost,
		Port: config.DBPort,
	}
	log.Infof("New Xorm Engine: %+v", dbOptions)
	engine, err := models.NewEngine(dbOptions)
	if err != nil {
		return nil, fmt.Errorf("new xorm engine err: %v", err)
	}

	return &DefaultManager{
		authenticator: config.Authenticator,
		engine: engine,
	}, nil
}

func (m DefaultManager) VerifyUsernamePassword(username, password, hash string) (bool, error) {
	//hash, err := mAuth.Hash(password)
	//if err != nil {
	//	return false, err
	//}
	return m.authenticator.Authenticate(username, password, hash)
}

func (m DefaultManager) GenerateToken(uid string, username string, exp ...int64) (string, error) {
	return m.authenticator.GenerateToken(uid, username, exp...)
}

func (m DefaultManager) GetUserToBeConfirmed(id string, username string) (*models.UserToBeConfirmed, error) {
	sess := m.engine.NewSession()
	defer sess.Close()

	u := &models.UserToBeConfirmed{
		Id: id,
		Name: username,
	}

	has, err := sess.Get(u)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	} else {
		return u, nil
	}
}

func (m DefaultManager) DeleteUsersToBeConfirmedByName(username string) error {
	sess := m.engine.NewSession()
	defer sess.Close()

	_, err := sess.And("name = ?", username).Delete(new(models.UserToBeConfirmed))
	return err
}

func (m DefaultManager) DeleteUserToBeConfirmed(uid string) error {
	sess := m.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}
	_, err := sess.ID(uid).Delete(new(models.UserToBeConfirmed))
	if err != nil {
		return err
	}
	return sess.Commit()
}
func (m DefaultManager) SaveUserToBeConfirmed(user *models.UserToBeConfirmed) error {
	sess := m.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}
	if _, err := sess.Insert(user); err != nil {
		sess.Rollback()
		return err
	}
	return sess.Commit()
}
func (m DefaultManager) SaveUser(user *models.User) error {
	sess := m.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}
	if _, err := sess.Insert(user); err != nil {
		sess.Rollback()
		return err
	}
	return sess.Commit()
}

func (m DefaultManager) VerifyAuthToken(token string) (map[string]interface{}, error) {
	errFailed := errors.New("Token 解析失败")
	success, data, err := mAuth.AuthToken(token)
	if err != nil {
		return nil, err
	}
	_username := data["username"]
	if _username == nil {
		return nil, errFailed
	}
	username := _username.(string)
	u := m.GetUserByName(username)
	if u == nil {
		return nil, fmt.Errorf("未找到用户: %s", username)
	}

	if !success {
		return nil, errFailed
	}


	return map[string]interface{}{
		"uid": u.Id,
		"username": u.Name,
	}, nil
}

func (m DefaultManager) CheckAdminUser() error {
	log.Debugf("CheckAdminUser ...")
	sess := m.engine.NewSession()
	defer sess.Close()
	adminUser := &models.User{
		Name: "admin",
	}
	if has, _ := sess.Get(adminUser); !has {
		pass, err := mAuth.Hash(settings.ADMIN_PASSWORD)
		if err != nil{
			return err
		}
		u := &models.User{
			Name: settings.ADMIN_NAME,
			Password: pass,
			Email: settings.ADMIN_EMAIL,
		}
		if _, err := sess.Insert(u); err != nil {
			sess.Rollback()
			return err
		} else {
			log.Infof("Add admin user: %s", u.Id)
		}
	} else {
		log.Infof("Admin user exists: %s", adminUser.Id)
	}
	return sess.Commit()
}

func (m DefaultManager) GetUsers(pageConfig boxlinker.PageConfig) (users []*models.User, err error) {
	fmt.Printf("pageConfig:> %+v", pageConfig)
	err = m.engine.Desc("created_unix").Limit(pageConfig.Limit(), pageConfig.Offset()).Find(&users)
	return
}

func (m DefaultManager) GetUserByName(username string) (*models.User) {
	u := &models.User{
		Name: username,
	}
	if found, _ := m.engine.Get(u); found {
		return u
	}
	return nil
}

func (m DefaultManager) GetUserById(id string) (*models.User) {
	u := &models.User{}
	if found, _ := m.engine.Id(id).Get(u); found {
		return u
	}
	return nil
}



func (m DefaultManager) IsUserExists(username string) (bool, error) {
	u := &models.User{
		Name: username,
	}
	if found, err := m.engine.Cols("name").Get(u); err != nil {
		return false, err
	} else if found {
		return true, nil
	} else {
		return false, nil
	}
}

func (m DefaultManager) IsEmailExists(email string) (bool, error) {
	u := &models.User{
		Email: email,
	}
	if found, err := m.engine.Cols("email").Get(u); err != nil {
		return false, err
	} else if found {
		return true, nil
	} else {
		return false, nil
	}
}

func (m DefaultManager) UpdatePassword(id string, password string) (bool, error) {
	sess := m.engine.NewSession()
	defer sess.Close()
	u := &models.User{
		Password: password,
	}
	_, err := m.engine.Id(id).Update(u)
	if err != nil {
		sess.Rollback()
		return false, err
	}
	return true, sess.Commit()
}

