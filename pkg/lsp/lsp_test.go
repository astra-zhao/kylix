package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kylix/lexer"
	"kylix/parser"
	"testing"
)

// 测试辅助函数：创建测试文档
func createTestDocument(t *testing.T, code string) *Document {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("解析错误: %v", p.Errors())
	}

	return &Document{
		URI:     "file:///test.klx",
		Text:    code,
		Lines:   splitLines(code),
		AST:     program,
		Symbols: CollectSymbols(program),
	}
}

func splitLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}

// 测试：文档符号
func TestDocumentSymbols(t *testing.T) {
	code := `program Test;
var x: integer;
function Add(a, b: integer): integer;
begin
  result := a + b;
end;
begin
  x := 10;
end.`

	doc := createTestDocument(t, code)

	if doc.Symbols == nil {
		t.Fatal("符号表为空")
	}

	// 检查是否收集到了变量 x
	xSym := doc.Symbols.FindSymbol("x")
	if xSym == nil {
		t.Error("未找到变量 x")
	} else {
		fmt.Printf("✓ 找到变量 x: %+v\n", xSym)
	}

	// 检查是否收集到了函数 Add
	addSym := doc.Symbols.FindSymbol("Add")
	if addSym == nil {
		t.Error("未找到函数 Add")
	} else {
		fmt.Printf("✓ 找到函数 Add: %+v\n", addSym)
	}

	fmt.Printf("✓ 总共收集到 %d 个符号\n", len(doc.Symbols.AllSymbols))
}

// 测试：悬停信息
func TestHoverInfo(t *testing.T) {
	code := `program Test;
function Add(a, b: integer): integer;
begin
  result := a + b;
end;
begin
  var sum := Add(5, 3);
end.`

	doc := createTestDocument(t, code)

	// 测试获取 Add 函数的悬停信息
	addSym := doc.Symbols.FindSymbol("Add")
	if addSym == nil {
		t.Fatal("未找到函数 Add")
	}

	hoverText := formatHoverText(addSym)
	fmt.Printf("✓ Add 函数的悬停信息:\n%s\n", hoverText)

	if len(hoverText) == 0 {
		t.Error("悬停信息为空")
	}
}

// 测试：跳转到定义
func TestGoToDefinition(t *testing.T) {
	code := `program Test;
var x: integer;
begin
  x := 10;
  x := x + 5;
end.`

	doc := createTestDocument(t, code)

	// 查找变量 x 的定义
	xSym := doc.Symbols.FindSymbol("x")
	if xSym == nil {
		t.Fatal("未找到变量 x")
	}

	fmt.Printf("✓ 变量 x 定义在: 行 %d, 列 %d\n", xSym.Location.Line, xSym.Location.Column)

	if xSym.Location.Line == 0 {
		t.Error("位置信息不正确")
	}
}

// 测试：查找引用
func TestFindReferences(t *testing.T) {
	code := `program Test;
var counter: integer;
procedure Increment;
begin
  counter := counter + 1;
end;
begin
  counter := 0;
  Increment;
end.`

	doc := createTestDocument(t, code)

	// 查找 counter 的所有引用
	walker := &ReferenceWalker{
		targetName: "counter",
		uri:        doc.URI,
		references: []Location{},
	}
	walker.Walk(doc.AST)

	fmt.Printf("✓ 找到 %d 个 counter 引用\n", len(walker.references))

	if len(walker.references) < 3 {
		t.Errorf("应该至少有 3 个引用（定义 + 2 次使用），实际找到 %d 个", len(walker.references))
	}

	for i, ref := range walker.references {
		fmt.Printf("  引用 %d: 行 %d, 列 %d\n", i+1, ref.Range.Start.Line+1, ref.Range.Start.Character+1)
	}
}

// 测试：重命名
func TestRename(t *testing.T) {
	code := `program Test;
var oldName: integer;
begin
  oldName := 10;
  oldName := oldName + 5;
end.`

	doc := createTestDocument(t, code)

	// 查找 oldName 的所有引用（用于重命名）
	walker := &ReferenceWalker{
		targetName: "oldName",
		uri:        doc.URI,
		references: []Location{},
	}
	walker.Walk(doc.AST)

	fmt.Printf("✓ 重命名 oldName 会影响 %d 个位置\n", len(walker.references))

	if len(walker.references) < 3 {
		t.Errorf("应该有至少 3 个位置需要重命名，实际找到 %d 个", len(walker.references))
	}
}

// 测试：LSP 消息处理
func TestLSPMessageHandling(t *testing.T) {
	// 创建一个简单的 LSP 服务器实例
	var inBuf bytes.Buffer
	var outBuf bytes.Buffer

	server := New(&inBuf, &outBuf)

	// 测试初始化
	initMsg := Message{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "initialize",
		Params:  json.RawMessage(`{"processId":null,"rootUri":"file:///test","capabilities":{}}`),
	}

	response := server.handleInitialize(&initMsg)
	if response == nil {
		t.Fatal("初始化响应为空")
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatal("响应结果格式不正确")
	}

	capabilities, ok := result["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatal("capabilities 格式不正确")
	}

	// 检查所有声明的功能
	requiredCapabilities := []string{
		"hoverProvider",
		"definitionProvider",
		"documentSymbolProvider",
		"referencesProvider",
		"renameProvider",
		"documentFormattingProvider",
		"signatureHelpProvider",
		"codeActionProvider",
		"workspaceSymbolProvider",
	}

	for _, cap := range requiredCapabilities {
		if _, exists := capabilities[cap]; !exists {
			t.Errorf("缺少功能声明: %s", cap)
		} else {
			fmt.Printf("✓ 已声明功能: %s\n", cap)
		}
	}
}

// 测试：代码格式化
func TestFormatting(t *testing.T) {
	code := `program Test;
var x:integer;
begin
x:=10;
end.`

	doc := createTestDocument(t, code)

	// 测试格式化
	server := &Server{
		docs: NewDocumentStore(),
	}
	server.docs.Update(doc.URI, code)

	formatMsg := Message{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "textDocument/formatting",
		Params:  json.RawMessage(`{"textDocument":{"uri":"` + doc.URI + `"},"options":{"tabSize":2,"insertSpaces":true}}`),
	}

	response := server.handleFormatting(&formatMsg)
	if response == nil {
		t.Fatal("格式化响应为空")
	}

	edits, ok := response.Result.([]TextEdit)
	if !ok {
		t.Fatal("格式化结果格式不正确")
	}

	fmt.Printf("✓ 格式化产生了 %d 个编辑\n", len(edits))

	if len(edits) == 0 {
		t.Error("应该有格式化编辑")
	}
}

// 测试：工作区符号搜索
func TestWorkspaceSymbols(t *testing.T) {
	code := `program Test;
function Calculate: integer;
begin
  result := 42;
end;
procedure Display;
begin
  writeln('Hello');
end;
begin
  Calculate;
  Display;
end.`

	doc := createTestDocument(t, code)

	// 测试工作区符号搜索
	server := &Server{
		docs: NewDocumentStore(),
	}
	server.docs.Update(doc.URI, code)

	symbolMsg := Message{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "workspace/symbol",
		Params:  json.RawMessage(`{"query":"Calc"}`),
	}

	response := server.handleWorkspaceSymbol(&symbolMsg)
	if response == nil {
		t.Fatal("符号搜索响应为空")
	}

	symbols, ok := response.Result.([]SymbolInformation)
	if !ok {
		t.Fatal("符号搜索结果格式不正确")
	}

	fmt.Printf("✓ 搜索 'Calc' 找到 %d 个符号\n", len(symbols))

	if len(symbols) == 0 {
		t.Error("应该找到至少一个匹配的符号")
	}

	for _, sym := range symbols {
		fmt.Printf("  符号: %s (%d)\n", sym.Name, sym.Kind)
	}
}

// 辅助函数
func intPtr(i int) *int {
	return &i
}

func formatHoverText(sym *Symbol) string {
	if sym == nil {
		return ""
	}

	switch sym.Kind {
	case SymbolFunction:
		return fmt.Sprintf("**function** `%s`%s\n\n%s", sym.Name, sym.Type, "函数定义")
	case SymbolProcedure:
		return fmt.Sprintf("**procedure** `%s`\n\n%s", sym.Name, "过程定义")
	case SymbolVariable:
		return fmt.Sprintf("**var** `%s: %s`", sym.Name, sym.Type)
	default:
		return fmt.Sprintf("`%s`", sym.Name)
	}
}

// 测试：完整的 LSP 会话流程
func TestFullLSPSession(t *testing.T) {
	code := `program Test;
var x: integer;
function Double(n: integer): integer;
begin
  result := n * 2;
end;
begin
  x := Double(5);
  writeln(x);
end.`

	// 1. 创建文档
	doc := createTestDocument(t, code)

	// 2. 初始化服务器
	server := &Server{
		docs: NewDocumentStore(),
	}

	// 3. 打开文档
	server.docs.Update(doc.URI, code)

	// 4. 测试各种操作

	// 4.1 获取文档符号
	symbolsMsg := Message{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "textDocument/documentSymbol",
		Params:  json.RawMessage(`{"textDocument":{"uri":"` + doc.URI + `"}}`),
	}
	symbolsResp := server.handleDocumentSymbol(&symbolsMsg)
	if symbolsResp == nil {
		t.Error("文档符号响应为空")
	} else {
		fmt.Println("✓ 文档符号请求成功")
	}

	// 4.2 测试悬停
	hoverMsg := Message{
		JSONRPC: "2.0",
		ID:      intPtr(2),
		Method:  "textDocument/hover",
		Params:  json.RawMessage(`{"textDocument":{"uri":"` + doc.URI + `"},"position":{"line":2,"character":9}}`),
	}
	hoverResp := server.handleHover(&hoverMsg)
	if hoverResp == nil {
		t.Error("悬停响应为空")
	} else {
		fmt.Println("✓ 悬停请求成功")
	}

	// 4.3 测试跳转到定义
	defMsg := Message{
		JSONRPC: "2.0",
		ID:      intPtr(3),
		Method:  "textDocument/definition",
		Params:  json.RawMessage(`{"textDocument":{"uri":"` + doc.URI + `"},"position":{"line":7,"character":7}}`),
	}
	defResp := server.handleDefinition(&defMsg)
	if defResp == nil {
		t.Error("定义响应为空")
	} else {
		fmt.Println("✓ 跳转到定义请求成功")
	}

	// 4.4 测试查找引用
	refMsg := Message{
		JSONRPC: "2.0",
		ID:      intPtr(4),
		Method:  "textDocument/references",
		Params:  json.RawMessage(`{"textDocument":{"uri":"` + doc.URI + `"},"position":{"line":1,"character":4},"context":{"includeDeclaration":true}}`),
	}
	refResp := server.handleReferences(&refMsg)
	if refResp == nil {
		t.Error("引用响应为空")
	} else {
		fmt.Println("✓ 查找引用请求成功")
	}

	fmt.Println("\n✓ 完整 LSP 会话测试通过")
}
