package orm_test

import (
	"testing"
	"github.com/RobinUS2/go-orm"
)

func TestOrm(t *testing.T) {
	o := orm.Create(orm.DefaultConfig())
	defer o.Close()
}
