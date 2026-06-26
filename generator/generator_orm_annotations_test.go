package generator

import "testing"

func TestGenerateORMEntityMethods(t *testing.T) {
	input := `
program ORMTest;
uses orm;

[Entity('users')]
type
  TUser = class
    [PrimaryKey]
    Id: Integer;

    [Column('email')]
    Email: String;

    Name: String;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "func (self *TUser) ToRow() map[string]interface{}")
	assertContains(t, out, `"Id": self.Id,`)
	assertContains(t, out, `"email": self.Email,`)
	assertContains(t, out, `"Name": self.Name,`)
	assertContains(t, out, "func (self *TUser) FromRow(row map[string]interface{})")
	assertContains(t, out, `if v, ok := row["email"].(string); ok {`)
}

func TestGenerateORMRepositoryMethods(t *testing.T) {
	input := `
program ORMRepo;
uses orm;

[Entity('users')]
type
  TUser = class
    [PrimaryKey]
    Id: Integer;
    Email: String;
  end;

[Repository(TUser)]
type
  TUserRepository = class
    [Query('SELECT * FROM users WHERE email = ?')]
    function ByEmail(email: String): TUser;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "func (self *TUserRepository) FindAll(orm *stdlib.ORM) []*TUser")
	assertContains(t, out, "func (self *TUserRepository) FindById(orm *stdlib.ORM, id int64) *TUser")
	assertContains(t, out, "func (self *TUserRepository) Save(orm *stdlib.ORM, e *TUser) int64")
	assertContains(t, out, "func (self *TUserRepository) DeleteById(orm *stdlib.ORM, id int64) int64")
	assertContains(t, out, `orm.Query("SELECT * FROM users WHERE email = ?", email)`)
}

func TestGenerateORMRepositoryListQuery(t *testing.T) {
	input := `
program ORMList;
uses orm;

[Entity('users')]
type
  TUser = class
    [PrimaryKey]
    Id: Integer;
    Name: String;
  end;

[Repository(TUser)]
type
  TUserRepository = class
    [Query('SELECT * FROM users')]
    function All(): array of TUser;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "func (self *TUserRepository) All(orm *stdlib.ORM) []*TUser")
	assertContains(t, out, `orm.QueryAll("SELECT * FROM users")`)
}

func TestGenerateORMSkipsUserDefinedMethod(t *testing.T) {
	input := `
program ORMUserDef;
uses orm;

[Entity('users')]
type
  TUser = class
    [PrimaryKey]
    Id: Integer;
    Name: String;
  end;

[Repository(TUser)]
type
  TUserRepository = class
    function FindAll(): String;
    begin
      result := 'user override';
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertNotContains(t, out, "func (self *TUserRepository) FindAll(orm *stdlib.ORM)")
	assertContains(t, out, "func (self *TUserRepository) FindAll() string")
}
