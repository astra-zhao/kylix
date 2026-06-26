package generator

import "testing"

func TestGenerateValidationMethods(t *testing.T) {
	input := `
program ValidationTest;

type
  TCreateUser = class
    [Required]
    [Email]
    Email: String;

    [MinLen(8)]
    Password: String;

    [Min(18)]
    Age: Integer;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "func (self *TCreateUser) Validate() map[string]string {")
	assertContains(t, out, `errors["Email"] = "is required"`)
	assertContains(t, out, "regexp.MustCompile(")
	assertContains(t, out, "len(self.Password) < 8")
	assertContains(t, out, "self.Age < 18")
	assertContains(t, out, "func (self *TCreateUser) IsValid() bool {")
	assertContains(t, out, "return len(self.Validate()) == 0")
}

func TestGenerateValidationSkipsClassesWithoutAttrs(t *testing.T) {
	input := `
program Plain;

type
  TPlain = class
    Name: String;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertNotContains(t, out, "func (self *TPlain) Validate()")
	assertNotContains(t, out, "func (self *TPlain) IsValid()")
}

func TestGenerateValidationSkipsWhenUserDefinedValidate(t *testing.T) {
	input := `
program UserDefined;

type
  TCustom = class
    [Required]
    Name: String;

    function Validate(): String;
    begin
      result := 'user';
    end;

    function IsValid(): Boolean;
    begin
      result := true;
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertNotContains(t, out, "func (self *TCustom) Validate() map[string]string")
	assertNotContains(t, out, "return len(self.Validate()) == 0")
}
