package orm_test

import (
	"testing"

	"../orm"
)

func TestOrm(t *testing.T) {
	orm := orm.Create(orm.DefaultConfig())
	defer orm.Close()
}
