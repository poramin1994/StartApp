package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type TmpFileList struct {
	Id          int64  `orm:"auto"`
	User        *User  `orm:"rel(fk)"`
	Path        string `orm:"null;size(255)"`
	Extension   string `orm:"null;size(32)"`
	ExpiredDate string `orm:"null;size(255)"`
	Delete      bool   `orm:"default(false)"`

	Created time.Time `orm:"auto_now_add;type(datetime)" json:"created"`
	Updated time.Time `orm:"auto_now;type(datetime)" json:"updated"`
}

func init() {
	orm.RegisterModel(new(TmpFileList))
}

// AddTmpFileList insert a new TmpFileList into database and returns
// last inserted Id on success.
func AddTmpFileList(m *TmpFileList) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetTmpFileListById retrieves TmpFileList by Id. Returns error if
// Id doesn't exist
func GetTmpFileListById(id int64) (v *TmpFileList, err error) {
	o := orm.NewOrm()
	v = &TmpFileList{Id: id}
	if err = o.QueryTable(new(TmpFileList)).Filter("Id", id).RelatedSel().One(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllTmpFileList retrieves all TmpFileList matches certain condition. Returns empty list if
// no records exist
func GetAllTmpFileList(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(TmpFileList))
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Filter(k, v)
	}
	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + v
				} else if order[i] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + v
				} else if order[0] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error: unused 'order' fields")
		}
	}

	var l []TmpFileList
	qs = qs.OrderBy(sortFields...).RelatedSel()
	if _, err = qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(fname).Interface()
				}
				ml = append(ml, m)
			}
		}
		return ml, nil
	}
	return nil, err
}

// UpdateTmpFileList updates TmpFileList by Id and returns error if
// the record to be updated doesn't exist
func UpdateTmpFileListById(m *TmpFileList) (err error) {
	o := orm.NewOrm()
	v := TmpFileList{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteTmpFileList deletes TmpFileList by Id and returns error if
// the record to be deleted doesn't exist
func DeleteTmpFileList(id int64) (err error) {
	o := orm.NewOrm()
	v := TmpFileList{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&TmpFileList{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
func GetToDeleteFileUpload() (v []*TmpFileList, err error) {
	o := orm.NewOrm()
	if _, err = o.QueryTable(new(TmpFileList)).Filter("Delete", false).Filter("ExpiredDate__lte", time.Now()).RelatedSel().All(&v); err == nil {
		return v, nil
	}
	return nil, err
}
