package models

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Prefix struct {
	Id       int64     `orm:"auto"`
	PrefixTh string    `orm:"null;size(125)"`
	PrefixEn string    `orm:"null;size(125)"`
	Created  time.Time `orm:"auto_now_add;type(datetime)" json:"created"`
	Updated  time.Time `orm:"auto_now;type(datetime)" json:"updated"`
}

func init() {
	orm.RegisterModel(new(Prefix))
}

// AddPrefix insert a new Prefix into database and returns
// last inserted Id on success.
func AddPrefix(m *Prefix) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetPrefixById retrieves Prefix by Id. Returns error if
// Id doesn't exist
func GetPrefixById(id int64) (v *Prefix, err error) {
	o := orm.NewOrm()
	v = &Prefix{Id: id}
	if err = o.QueryTable(new(Prefix)).Filter("Id", id).RelatedSel().One(v); err == nil {
		return v, nil
	}
	return nil, err
}

// UpdatePrefix updates Prefix by Id and returns error if
// the record to be updated doesn't exist
func UpdatePrefixById(o orm.Ormer, m *Prefix) (err error) {
	if o == nil {
		o = orm.NewOrm()
	}
	v := Prefix{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeletePrefix deletes Prefix by Id and returns error if
// the record to be deleted doesn't exist
func DeletePrefix(id int64) (err error) {
	o := orm.NewOrm()
	v := Prefix{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&Prefix{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}

func GetPrefixByPrefixTh(name string) (v *Prefix, err error) {
	o := orm.NewOrm()
	v = &Prefix{}
	if err = o.QueryTable(new(Prefix)).Filter("PrefixTh", name).RelatedSel().One(v); err == nil {
		return v, nil
	}
	return nil, err
}

func GetPrefixlList() (v []*Prefix, err error) {
	o := orm.NewOrm()
	if _, err = o.QueryTable(new(Prefix)).RelatedSel().All(&v); err == nil {
		return v, nil
	}
	return nil, err
}