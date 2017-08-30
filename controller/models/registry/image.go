package registry

import (
	"time"
	"github.com/go-xorm/xorm"
	"github.com/satori/go.uuid"
)

type Image struct {
	Id 		string 	`xorm:"pk"`
	Repository 	string 	`xorm:"INDEX UNIQUE NOT NULL"`
	Tag 	string 	`xorm:"NOT NULL"`
	Size 	int64
	Digest 	string 	`xorm:"NOT NULL"`
	IsPrivate bool

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64

}


func (me *Image) BeforeInsert() {
	me.Id = uuid.NewV4().String()
	me.CreatedUnix = time.Now().Unix()
	me.UpdatedUnix = me.CreatedUnix
}

func (me *Image) BeforeUpdate() {
	me.UpdatedUnix = time.Now().Unix()
}

func (me *Image) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
	case "created_unix":
		me.Created = time.Unix(me.CreatedUnix, 0).Local()
	case "updated_unix":
		me.Updated = time.Unix(me.UpdatedUnix, 0)
	}
}
