package generator

import "testing"

func TestGenerateBootProcedureRoute(t *testing.T) {
	input := `
program BootProcedure;
uses boot;

[Controller('/proc')]
type
  TProcController = class
    [Get('/hello')]
    procedure Hello(req: TRequest; res: TResponse);
    begin
      res.StatusCode(200);
      res.Send('hello');
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `stdlib.BootGET("/proc/hello", func(req *stdlib.BootRequest) *stdlib.BootResponse {`)
	assertContains(t, out, `res := stdlib.BootText(200, "")`)
	assertContains(t, out, `__kylix_ctrl_TProcController.Hello(req, res)`)
	assertContains(t, out, `return res`)
}

func TestGenerateBootControllerGetRoute(t *testing.T) {
	input := `
program BootAutoWire;
uses boot;

[Controller('/api')]
type
  THelloController = class
    [Get('/hello')]
    function Hello(req: TRequest): TResponse;
    begin
      result := boot.Text(200, 'Hello');
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `"kylix/stdlib"`)
	assertContains(t, out, `func (self *THelloController) Hello(req *stdlib.BootRequest) *stdlib.BootResponse`)
	assertContains(t, out, `__kylix_ctrl_THelloController := &THelloController{}`)
	assertContains(t, out, `stdlib.BootGET("/api/hello", func(req *stdlib.BootRequest) *stdlib.BootResponse {`)
	assertContains(t, out, `return __kylix_ctrl_THelloController.Hello(req)`)
}

func TestGenerateBootRouteMethods(t *testing.T) {
	input := `
program BootMethods;
uses boot;

[Controller('/api')]
type
  TController = class
    [Post('/items')]
    function Create(req: TRequest): TResponse;
    begin result := boot.Text(201, 'created'); end;

    [Put('/items')]
    function Update(req: TRequest): TResponse;
    begin result := boot.Text(200, 'updated'); end;

    [Delete('/items')]
    function Remove(req: TRequest): TResponse;
    begin result := boot.Text(204, 'removed'); end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `stdlib.BootPOST("/api/items"`)
	assertContains(t, out, `stdlib.BootPUT("/api/items"`)
	assertContains(t, out, `stdlib.BootDELETE("/api/items"`)
}

func TestGenerateBootRouteIgnoresUnsupportedSignature(t *testing.T) {
	input := `
program BootUnsupported;
uses boot;

[Controller('/api')]
type
  TController = class
    [Get('/bad')]
    procedure Bad(req: TRequest);
    begin
      WriteLn('bad');
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertNotContains(t, out, `stdlib.BootGET("/api/bad"`)
}

func TestNormalizeBootPath(t *testing.T) {
	cases := []struct {
		base string
		sub  string
		want string
	}{
		{"/api", "/users", "/api/users"},
		{"/api/", "/users", "/api/users"},
		{"api", "users", "/api/users"},
		{"/api", "", "/api"},
		{"", "users", "/users"},
		{"", "", "/"},
	}
	for _, tc := range cases {
		if got := normalizeBootPath(tc.base, tc.sub); got != tc.want {
			t.Fatalf("normalizeBootPath(%q, %q) = %q, want %q", tc.base, tc.sub, got, tc.want)
		}
	}
}
