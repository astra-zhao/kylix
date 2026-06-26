package generator

import "testing"

func TestGenerateBootServiceAndInject(t *testing.T) {
	input := `
program BootDI;
uses boot;

[Service]
type
  TUserService = class
    function Greeting(): String;
    begin
      result := 'Hello';
    end;
  end;

[Controller('/di')]
type
  TUserController = class
    [Inject]
    UserService: TUserService;

    [Get('/hello')]
    function Hello(req: TRequest): TResponse;
    begin
      result := BootText(200, self.UserService.Greeting());
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `__kylix_svc_TUserService := &TUserService{}`)
	assertContains(t, out, `stdlib.BootRegisterInstance("TUserService", __kylix_svc_TUserService)`)
	assertContains(t, out, `stdlib.BootRegisterInstance("UserService", __kylix_svc_TUserService)`)
	assertContains(t, out, `__kylix_ctrl_TUserController := &TUserController{}`)
	assertContains(t, out, `__kylix_ctrl_TUserController.UserService = __kylix_svc_TUserService`)
	assertContains(t, out, `stdlib.BootGET("/di/hello"`)
}

func TestGenerateBootComponentAndServiceInjection(t *testing.T) {
	input := `
program BootDI;
uses boot;

[Component]
type
  TConfig = class
    function Name(): String;
    begin result := 'cfg'; end;
  end;

[Service]
type
  TUserService = class
    [Inject]
    Config: TConfig;

    function Greeting(): String;
    begin result := self.Config.Name(); end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `__kylix_svc_TConfig := &TConfig{}`)
	assertContains(t, out, `__kylix_svc_TUserService := &TUserService{}`)
	assertContains(t, out, `__kylix_svc_TUserService.Config = __kylix_svc_TConfig`)
}

func TestGenerateBootInjectUnknownDependencySkipped(t *testing.T) {
	input := `
program BootDI;
uses boot;

[Controller('/di')]
type
  TUserController = class
    [Inject]
    Missing: TMissingService;

    [Get('/hello')]
    function Hello(req: TRequest): TResponse;
    begin
      result := BootText(200, 'OK');
    end;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertNotContains(t, out, `.Missing =`)
	assertContains(t, out, `stdlib.BootGET("/di/hello"`)
}
