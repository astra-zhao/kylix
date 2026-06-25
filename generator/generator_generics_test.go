package generator

import "testing"

func TestGenerateGenericClassMethodReceiver(t *testing.T) {
	input := `
program GenericClass;

class TStack<T>
private
  Items: array[0..9] of T;
  Count: Integer;
public
  procedure Push(item: T);
  begin
    self.Items[self.Count] := item;
    self.Count := self.Count + 1;
  end;

  function Pop(): T;
  begin
    self.Count := self.Count - 1;
    result := self.Items[self.Count];
  end;
end;

begin
  var s := TStack<Integer>.Create();
  s.Push(1);
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "type TStack[T interface{}] struct")
	assertContains(t, out, "func (self *TStack[T]) Push(item T)")
	assertContains(t, out, "func (self *TStack[T]) Pop() T")
	assertNotContains(t, out, "func (self *TStack) Push")
}
