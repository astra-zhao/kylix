package generator

import "testing"

func TestGenerateBootAuthenticatedRoute(t *testing.T) {
	input := `
program SecureRoute;
uses boot;

[Controller('/admin')]
type
  TAdminController = class
    [Get('/dashboard')]
    [Authenticated]
    function Dashboard(req: TRequest): TResponse;
    begin
      result := BootText(200, 'dashboard');
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `stdlib.BootGET("/admin/dashboard", func(req *stdlib.BootRequest) *stdlib.BootResponse {`)
	assertContains(t, out, `if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }`)
	assertContains(t, out, `return __kylix_ctrl_TAdminController.Dashboard(req)`)
}

func TestGenerateBootRoleRoute(t *testing.T) {
	input := `
program SecureRoute;
uses boot;

[Controller('/admin')]
type
  TAdminController = class
    [Get('/users')]
    [Role('admin')]
    function ListUsers(req: TRequest): TResponse;
    begin
      result := BootText(200, 'users');
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }`)
	assertContains(t, out, `if __r := stdlib.BootEnforceRole(req, "admin"); __r != nil { return __r }`)
}

func TestGenerateBootRouteWithoutSecurity(t *testing.T) {
	input := `
program PlainRoute;
uses boot;

[Controller('/api')]
type
  TPlain = class
    [Get('/hi')]
    function Hi(req: TRequest): TResponse;
    begin
      result := BootText(200, 'hi');
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertNotContains(t, out, `BootEnforceAuth`)
	assertNotContains(t, out, `BootEnforceRole`)
}
