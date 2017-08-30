package user

import (
	"net/http"
	"github.com/cabernety/boxlinker"
	"github.com/cabernety/boxlinker/controller/models"
	"regexp"
	"fmt"
	"github.com/cabernety/boxlinker/auth"
)

type RegForm struct {
	Username 	string 	`json:"username"`
	Password 	string 	`json:"password"`
	Email 		string 	`json:"email"`
}

func (f *RegForm) validate() map[string]int {
	m := make(map[string]int)
	if f.Username == "" {
		m["username"] = boxlinker.STATUS_FIELD_REQUIRED
	}

	if f.Password == "" {
		m["password"] = boxlinker.STATUS_FIELD_REQUIRED
	} else if len(f.Password) < 6 {
		m["password"] = boxlinker.STATUS_FIELD_REGEX_FAILED
	}

	if f.Email == "" {
		m["email"] = boxlinker.STATUS_FIELD_REQUIRED
	} else {
		if ok, err := regexp.MatchString("[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+", f.Email); err != nil {
			fmt.Printf("regexp err: %v", err)
		} else if !ok {
			m["email"] = boxlinker.STATUS_FIELD_REGEX_FAILED
		}
	}
	if len(m) != 0 {
		return m
	}
	return nil
}

func (a *Api) Reg(w http.ResponseWriter, r *http.Request) {
	form := &RegForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	if msg := form.validate(); msg != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, msg)
		return
	}
	if found, err := a.manager.IsUserExists(form.Username); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	} else if found {
		boxlinker.Resp(w, boxlinker.STATUS_USER_EXISTS, nil)
		return
	}

	if found, err := a.manager.IsEmailExists(form.Email); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	} else if found {
		boxlinker.Resp(w, boxlinker.STATUS_EMAIL_EXISTS, nil)
		return
	}


	//pass, err := auth.Hash(form.Password)
	pass, err := auth.Hash(form.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := a.manager.SaveUser(&models.User{
		Name: form.Username,
		Password: string(pass),
		Email: form.Email,
	}); err != nil {
		boxlinker.Resp(w, 1, nil, err.Error())
		return
	}
	boxlinker.Resp(w, 0, nil, "success")
}
