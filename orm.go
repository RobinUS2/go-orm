package orm

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/sqlite" // doesn't work well with cross compilation
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Orm struct {
	Conf *Conf

	// Unexposed backend apis
	db          *gorm.DB
	dbCloneLock sync.RWMutex

	// Private
	registeredModels map[string]interface{}
}

// Raw backend
func (orm *Orm) RawBackend() *gorm.DB {
	return orm.db
}

// Error handler
func (orm *Orm) Error() error {
	return orm.db.Error
}

// Value handler
func (orm *Orm) Value() interface{} {
	return orm.db.Value
}

// Open connection
func (orm *Orm) Open() {
	var err error
	// @todo escape?
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", orm.Conf.Username, orm.Conf.Password, orm.Conf.Hostname, orm.Conf.Port, orm.Conf.Database)
	if len(orm.Conf.ConnectionString) > 0 {
		connectionStr = orm.Conf.ConnectionString
	}
	orm.db, err = gorm.Open(orm.Conf.Dialect, connectionStr)
	if err != nil {
		panic(fmt.Sprintf("Unable to open database: %s", err))
	}

	// Logging
	if orm.Conf.DebugLogging {
		orm.db.LogMode(true)
		log.Println("Opened ORM")
	}
}

// select table
func (orm *Orm) Table(table string) *Orm {
	clone := orm.clone()
	clone.db = clone.db.Table(table)
	return clone
}

// Setup tables automatically based on models
func (orm *Orm) AutoMigrate(values ...interface{}) *Orm {
	clone := orm.clone()
	clone.db.AutoMigrate(values...)
	return clone
}

// First record
func (orm *Orm) First(out interface{}, where ...interface{}) *Orm {
	clone := orm.clone()
	clone.db.First(out, where...)
	return clone
}

// Last record
func (orm *Orm) Last(out interface{}, where ...interface{}) *Orm {
	clone := orm.clone()
	clone.db.Last(out, where...)
	return clone
}

// Find
func (orm *Orm) Find(out interface{}, where ...interface{}) *Orm {
	clone := orm.clone()
	clone.db.Find(out, where...)
	return clone
}

// Where
func (orm *Orm) Where(query interface{}, args ...interface{}) *Orm {
	clone := orm.clone()
	clone.db.Where(query, args...)
	return clone
}

// Is this a new record? true = yes (no primary key)
func (orm *Orm) IsNewRecord(value interface{}) bool {
	return orm.db.NewRecord(value)
}

// Create a new record
func (orm *Orm) Create(value interface{}) *Orm {
	clone := orm.clone()
	clone.db.Create(value)
	return clone
}

// Delete a record
func (orm *Orm) Delete(value interface{}) *Orm {
	if value == nil {
		panic("Can not delete nil value")
	}
	clone := orm.clone()
	clone.db.Limit(1).Delete(value)
	return clone
}

func (orm *Orm) Model(value interface{}) *Orm {
	clone := orm.clone()
	clone.db = clone.db.Model(value)
	return clone
}

// Count
func (orm *Orm) Count(model interface{}, value interface{}) *Orm {
	clone := orm.clone()
	clone.db.Value = model
	clone.db.Count(value)
	return clone
}

// Close
func (orm *Orm) Close() {
	orm.db.Close()
	if orm.Conf.DebugLogging {
		log.Println("Closed ORM")
	}
}

// Has table
func (orm *Orm) HasTable(model interface{}) bool {
	return orm.db.HasTable(model)
}

// Create table
func (orm *Orm) CreateTable(model interface{}) *Orm {
	orm.db.CreateTable(model)
	return orm
}

// Clone ORM
func (orm *Orm) clone() *Orm {
	orm.dbCloneLock.Lock()
	newDb := orm.db.New()
	orm.dbCloneLock.Unlock()

	clone := Orm{
		Conf:             orm.Conf,
		db:               newDb,
		registeredModels: orm.registeredModels,
	}

	return &clone
}

// SQL rows
func (orm *Orm) Rows() (*sql.Rows, error) {
	return orm.db.Rows()
}

// SQL row
func (orm *Orm) Row() *sql.Row {
	return orm.db.Row()
}

// Save
func (orm *Orm) Save(elm interface{}) *Orm {
	clone := orm.clone()
	clone.db.Save(elm)
	return clone
}

// Drop table
func (orm *Orm) DropTable(values ...interface{}) *Orm {
	if orm.Conf.SafeModeEnabled {
		log.Println("Unable to drop table, safe mode enabled")
		return orm
	}
	orm.db.DropTable(values...)
	return orm
}

// Get model
func (orm *Orm) GetModel(name string) interface{} {
	return orm.registeredModels[name]
}

// Register model
func (orm *Orm) RegisterModel(model interface{}) {
	modelI := model.(ModelI)
	name := modelI.GetName()
	if orm.registeredModels[name] != nil {
		panic(fmt.Sprintf("Model %s already registered", name))
	}
	orm.registeredModels[name] = &model
}

// By ID
func (orm *Orm) FetchById(out interface{}, id uint) *Orm {
	return orm.First(out, id)
}

// List models
func (orm *Orm) RegisteredModels() map[string]interface{} {
	return orm.registeredModels
}

// New instance
func Create(conf *Conf) *Orm {
	orm := &Orm{
		Conf:             conf,
		registeredModels: make(map[string]interface{}),
	}
	if orm.Conf.AutoOpen {
		orm.Open()
	}
	return orm
}
