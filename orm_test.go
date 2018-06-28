package orm_test

import (
	"testing"
	"../go-orm"
)

func TestOrm(t *testing.T) {
	o := orm.Create(orm.DefaultConfig())
	defer o.Close()
}
