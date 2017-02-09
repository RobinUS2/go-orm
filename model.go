package orm

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"reflect"
	"strconv"
	"time"
)

type Model struct {
	gorm.Model
	Name         string                                        `gorm:"-" json:"-"`
	StructType   reflect.Type                                  `gorm:"-" json:"-"`
	DoSpecialize func(data map[string]interface{}) interface{} `gorm:"-" json:"-"`
	DoClone      func() interface{}                            `gorm:"-" json:"-"`
	DoValidate   func(data map[string]interface{}) error       `gorm:"-" json:"-"`
}

type ModelI interface {
	Save(orm *Orm)
	Clone() interface{}
	GetModel() ModelI
	GetName() string
	GetStructType() reflect.Type
	SpecializeRow(orm *Orm, data map[string]interface{}) interface{} // convert data map to actual struct with fields populated
	Validate(data map[string]interface{}) error                      // validate user input

	// Query methods
	First(orm *Orm, where ...interface{}) *Orm
	Count(orm *Orm) *Orm
	Find(orm *Orm, query interface{}, args ...interface{}) *Orm
	Delete(orm *Orm, value interface{}) *Orm
	Create(orm *Orm, changes map[string]interface{}) *Orm
	Update(orm *Orm, value interface{}, changes map[string]interface{}) *Orm
}

func (model Model) Save(orm *Orm) {
	clone := orm.clone()
	clone.db.Model(model.Clone()).Save(model)
}

func (model Model) ParseID(val interface{}) uint {
	if val == nil {
		return 0
	}
	if f, ok := val.(int64); ok {
		return uint(f)
	}
	u, _ := strconv.ParseUint(fmt.Sprintf("%s", val.([]byte)), 10, 64)
	return uint(u)
}

func (model Model) ParseTime(val interface{}) *time.Time {
	if val == nil {
		return nil
	}
	if f, ok := val.(time.Time); ok {
		return &f
	}
	if f, ok := val.(*time.Time); ok {
		return f
	}
	return nil
}

func (model Model) ParseString(val interface{}) string {
	if val == nil {
		return ""
	}
	if f, ok := val.(string); ok {
		return f
	}
	if f, ok := val.([]byte); ok {
		return string(f)
	}
	return ""
}

func (model Model) SpecializeRow(orm *Orm, data map[string]interface{}) interface{} {
	return model.DoSpecialize(data)
}

func (model Model) Validate(data map[string]interface{}) error {
	return model.DoValidate(data)
}

func (model Model) Clone() interface{} {
	return model.DoClone()
}

func (model Model) GetModel() ModelI {
	return model
}

func (model Model) GetName() string {
	return model.Name
}

func (model Model) GetStructType() reflect.Type {
	return model.StructType
}

func (model Model) CloneInterface() interface{} {
	i := reflect.New(model.GetStructType()).Interface()
	return i
}

func (model Model) Delete(orm *Orm, value interface{}) *Orm {
	clone := orm.clone()
	clone.Delete(value)
	return clone
}

func (model Model) Update(orm *Orm, value interface{}, changes map[string]interface{}) *Orm {
	clone := orm.clone()

	// Validate
	validationErr := model.Validate(changes)
	if validationErr != nil {
		clone.db.Value = nil
		clone.db.Error = validationErr
		return clone
	}

	// Update
	clone.db.Model(value).Limit(1).Updates(changes)
	clone.db.Value = value
	clone.db.Error = clone.Error()
	return clone
}

func (model Model) Create(orm *Orm, changes map[string]interface{}) *Orm {
	clone := orm.clone()

	// Validate
	validationErr := model.Validate(changes)
	if validationErr != nil {
		clone.db.Value = nil
		clone.db.Error = validationErr
		return clone
	}

	// Use json lib to conveniently fill our struct
	value := model.CloneInterface()
	jb, me := json.Marshal(changes)
	if me != nil {
		clone.db.Value = nil
		clone.db.Error = me
		return clone
	}
	um := json.Unmarshal(jb, value)
	if um != nil {
		clone.db.Value = nil
		clone.db.Error = um
		return clone
	}

	// Create
	clone.db.Model(model).Create(value)
	clone.db.Value = value
	clone.db.Error = clone.Error()
	return clone
}

func (model Model) First(orm *Orm, where ...interface{}) *Orm {
	res := model.CloneInterface()
	clone := orm.clone()
	clone.First(res, where...)
	if clone.IsNewRecord(res) {
		res = nil
	}
	clone.db.Value = res
	clone.db.Error = clone.Error()
	log.Printf("%v %v", res, clone.Value())
	return clone
}

func (model Model) Count(orm *Orm) *Orm {
	var res uint64
	orm = orm.clone().Count(model.CloneInterface(), &res)
	orm.db.Value = res
	orm.db.Error = orm.Error()
	return orm
}

func (model Model) Find(orm *Orm, query interface{}, args ...interface{}) *Orm {
	var modelRows []interface{} = make([]interface{}, 0)
	var rows *sql.Rows
	var err error
	// needs to be directly on db to work
	clone := orm.clone()
	rows, err = clone.db.Model(model.Clone()).Where(query, args...).Rows()
	if err != nil {
		log.Printf("%#v", err)
	}
	// @todo handle error
	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	//	log.Printf("%v", columns)
	defer rows.Close()
	for rows.Next() {
		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		//log.Printf("%v %v", valuePtrs, scanErr)
		tmp_struct := make(map[string]interface{})

		for i, col := range columns {
			tmp_struct[col] = values[i]
		}
		//log.Printf("tmp struct %v", tmp_struct)

		// update row res
		rowRes := model.SpecializeRow(clone, tmp_struct)
		modelRows = append(modelRows, rowRes)

	}
	clone.db.Value = modelRows // @todo
	clone.db.Error = err
	//log.Printf("%v %v", rows, err)
	return clone
}
