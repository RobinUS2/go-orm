package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Orm struct {
	*gorm.DB
}

// New instance
func Create(conf *Conf) *Orm {
	var err error
	// @todo escape?
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", conf.Username, conf.Password, conf.Hostname, conf.Port, conf.Database)
	if len(conf.ConnectionString) > 0 {
		connectionStr = conf.ConnectionString
	}
	db, err := gorm.Open(conf.Dialect, connectionStr)
	if err != nil {
		panic(fmt.Sprintf("Unable to open database: %s", err))
	}

	o := &Orm{
		db,
	}

	return o
}
