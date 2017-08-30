package models

import (
	"github.com/go-xorm/xorm"
	"fmt"
	"net/url"
	"strings"
	log "github.com/Sirupsen/logrus"
	"github.com/go-xorm/core"
	_ "github.com/go-sql-driver/mysql"
)

var (
	tables []interface{}
)

func init(){
	tables = append(tables, new(User))
}


type DBOptions struct {
	User string
	Password string
	Host string
	Port int
	Name string
}


func NewEngine(config DBOptions) (*xorm.Engine, error){
	var Param string = "?"
	if strings.Contains(config.Name, Param) {
		Param = "&"
	}
	var connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&loc=%s",
		url.QueryEscape(config.User),
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		config.Name,"Asia%2FShanghai")

	log.Infof("Connect to db: %s", connStr)
	x, err := xorm.NewEngine("mysql", connStr)
	if err != nil {
		return nil,err
	}
	log.Info("Connect to db ok.")
	x.SetMapper(core.GonicMapper{})
	log.Infof("start to sync tables ...")
	if err = x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
		return nil, fmt.Errorf("sync tables err: %v",err)
	}
	x.ShowSQL(true)
	return x, nil
}

