package user

import (
	"net/http"
	"github.com/cabernety/boxlinker"
	log "github.com/Sirupsen/logrus"
	settings "github.com/cabernety/boxlinker/settings/user"
	"time"
)

type LoginForm struct {
	Username string		`json:"username"`
	Password string		`json:"password"`
}

func (f *LoginForm) validate() string {
	if f.Username == "" {
		return "您还没有填写用户名"
	}
	if f.Password == "" {
		return "您还没有填写用户名"
	}
	return ""
}

func (a *Api) Login(w http.ResponseWriter, r *http.Request){
	form := &LoginForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	log.Debugf("form: %v", form)
	if msg := form.validate(); msg != "" {
		boxlinker.Resp(w, 1, nil, msg)
		return
	}

	u := a.manager.GetUserByName(form.Username)
	success, err := a.manager.VerifyUsernamePassword(form.Username, form.Password, u.Password)
	if err != nil {
		boxlinker.Resp(w, 1, nil, err.Error())
		return
	}
	if !success {
		boxlinker.Resp(w, 1, nil, "failed")
		return
	}
	token, err := a.manager.GenerateToken(u.Id, u.Name)
	if err != nil {
		boxlinker.Resp(w, 1, nil, err.Error())
		return
	}
	cookie := &http.Cookie{
		Name:"X-Access-Token",
		Value: token,
		Expires: time.Now().Add(30*24*time.Hour),
		Domain: settings.COOKIE_DOMAIN,
	}
	http.SetCookie(w, cookie)
	boxlinker.Resp(w, 0, map[string]string{
		"X-Access-Token": token,
	})
}




