package main

import (
	"fmt"
	"kylix/stdlib"
)

type TUser struct {
Id int64
Email string
Name string
}

func (self *TUser) ToRow() map[string]interface{} {
	return map[string]interface{}{
		"Id": self.Id,
		"email": self.Email,
		"Name": self.Name,
	}
}

func (self *TUser) FromRow(row map[string]interface{}) {
	if v, ok := row["Id"].(int64); ok {
		self.Id = v
	}
	if v, ok := row["email"].(string); ok {
		self.Email = v
	}
	if v, ok := row["Name"].(string); ok {
		self.Name = v
	}
}

type TUserRepository struct {
}

func (self *TUserRepository) FindAll(orm *stdlib.ORM) []*TUser {
	rows, err := orm.FindAll("users")
	if err != nil { return nil }
	out := make([]*TUser, 0, len(rows))
	for _, row := range rows {
		e := &TUser{}
		e.FromRow(row)
		out = append(out, e)
	}
	return out
}

func (self *TUserRepository) FindById(orm *stdlib.ORM, id int64) *TUser {
	row, err := orm.Find("users", id)
	if err != nil || row == nil { return nil }
	e := &TUser{}
	e.FromRow(row)
	return e
}

func (self *TUserRepository) Save(orm *stdlib.ORM, e *TUser) int64 {
	if e == nil { return 0 }
	if e.Id == 0 {
		id, _ := orm.Insert("users", e.ToRow())
		e.Id = id
		return id
	}
	orm.Update("users", map[string]interface{}{"Id": e.Id}, e.ToRow())
	return e.Id
}

func (self *TUserRepository) DeleteById(orm *stdlib.ORM, id int64) int64 {
	n, _ := orm.Delete("users", map[string]interface{}{"Id": id})
	return n
}

func (self *TUserRepository) ByEmail(orm *stdlib.ORM, email string) *TUser {
	row, err := orm.Query("SELECT * FROM users WHERE email = ?", email)
	if err != nil || row == nil { return nil }
	e := &TUser{}
	e.FromRow(row)
	return e
}

func (self *TUserRepository) All(orm *stdlib.ORM) []*TUser {
	rows, err := orm.QueryAll("SELECT * FROM users")
	if err != nil { return nil }
	out := make([]*TUser, 0, len(rows))
	for _, row := range rows {
		e := &TUser{}
		e.FromRow(row)
		out = append(out, e)
	}
	return out
}

func main() {
//line example47_orm_annotations.klx:27
fmt.Println("KylixBoot ORM annotations OK")
}
