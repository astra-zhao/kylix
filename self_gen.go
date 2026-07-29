package main

import (
  "fmt"
  "strings"
  "strconv"
  "os"
)

const (
  tkIllegal TTokenType = iota
  tkEOF
  tkComment
  tkIdent
  tkInt
  tkFloat
  tkString
  tkStringInterp
  tkChar
  tkAssign
  tkPlus
  tkMinus
  tkBang
  tkAsterisk
  tkSlash
  tkMod
  tkDiv
  tkLT
  tkGT
  tkEQ
  tkNotEQ
  tkLTEQ
  tkGTEQ
  tkComma
  tkSemicolon
  tkColon
  tkDot
  tkDotDot
  tkArrow
  tkFatArrow
  tkAssignOp
  tkLParen
  tkRParen
  tkLBrace
  tkRBrace
  tkLBracket
  tkRBracket
  tkProgram
  tkUnit
  tkUses
  tkVar
  tkConst
  tkType
  tkBegin
  tkEnd
  tkFunction
  tkProcedure
  tkIf
  tkThen
  tkElse
  tkWhile
  tkDo
  tkFor
  tkTo
  tkDownTo
  tkRepeat
  tkUntil
  tkCase
  tkOf
  tkWith
  tkTry
  tkExcept
  tkFinally
  tkRaise
  tkOn
  tkClass
  tkInterface
  tkObject
  tkRecord
  tkArray
  tkSet
  tkPacked
  tkFile
  tkMap
  tkVariant
  tkInherits
  tkImplements
  tkImplementation
  tkPublic
  tkPrivate
  tkProtected
  tkPublished
  tkProperty
  tkRead
  tkWrite
  tkDefault
  tkStored
  tkVirtual
  tkOverride
  tkAbstract
  tkStatic
  tkDynamic
  tkExternal
  tkForward
  tkInline
  tkResult
  tkSelf
  tkNil
  tkTrue
  tkFalse
  tkAnd
  tkOr
  tkNot
  tkXor
  tkIn
  tkIs
  tkAs
  tkNew
  tkDelete
  tkBreak
  tkContinue
  tkExit
  tkReturn
  tkAsync
  tkAwait
  tkConstructor
  tkDestructor
  tkInherited
  tkMatch
  tkWhen
  tkImport
  tkExport
  tkModule
)

type TTokenType int

type TToken struct {
TokenType TTokenType  
Literal string  
Line int64  
Column int64  
}

type TDiagnostic struct {
FileName string  
Line int64  
Column int64  
Level string  
Message string  
Source string  
}

func (self *TDiagnostic) IsValid() bool {
  return true
}

type TErrorList struct {
Errors []*TDiagnostic  
}

func (self *TErrorList) Count() int64 {
var result int64  
result = int64(len(self.Errors))  
  return result
}

func (self *TErrorList) Add(fileName string, line int64, col int64, level string, msg string) {
var d *TDiagnostic  
d = &TDiagnostic{}  
d.FileName = fileName  
d.Line = line  
d.Column = col  
d.Level = level  
d.Message = msg  
d.Source = ""  
self.Errors = append(self.Errors, d)  
  _ = d
}

func (self *TErrorList) AddError(fileName string, line int64, col int64, msg string) {
self.Add(fileName, line, col, "error", msg)  
}

func (self *TErrorList) AddWarning(fileName string, line int64, col int64, msg string) {
self.Add(fileName, line, col, "warning", msg)  
}

func (self *TErrorList) HasErrors() bool {
var result bool  
var i int64  
result = false  
for i = 0; i <= (int64(len(self.Errors)) - 1); i  ++ {
if (self.Errors[i].Level == "error")     {
result = true      
      return result
    }
  }
  _ = i
  return result
}

func (self *TErrorList) ToString() string {
var result string  
var i int64  
var s string  
s = ""  
for i = 0; i <= (int64(len(self.Errors)) - 1); i  ++ {
s = ((s + self.Errors[i].FileName) + "(")    
s = ((s + fmt.Sprintf("%d", self.Errors[i].Line)) + ":")    
s = ((s + fmt.Sprintf("%d", self.Errors[i].Column)) + "): ")    
s = ((s + self.Errors[i].Level) + ": ")    
s = (s + self.Errors[i].Message)    
if (i < (int64(len(self.Errors)) - 1))     {
s = (s + "\n")      
    }
  }
result = s  
  _ = i
  _ = s
  return result
}

func (self *TErrorList) IsValid() bool {
  return true
}

type TSourcePosition struct {
FileName string  
Line int64  
Column int64  
Offset int64  
}

type TNode struct {
}

func (self *TNode) IsValid() bool {
  return true
}

type TStatement struct {
  TNode
}

func (self *TStatement) IsValid() bool {
  return true
}

type TExpression struct {
  TNode
}

func (self *TExpression) IsValid() bool {
  return true
}

type TParameter struct {
  TNode
Token TToken  
Name string  
ParamType interface{}  
}

func (self *TParameter) IsValid() bool {
  return true
}

type TTypeParameter struct {
  TNode
Token TToken  
Name string  
Constraint interface{}  
}

func (self *TTypeParameter) IsValid() bool {
  return true
}

type TCaseBranch struct {
  TNode
Values []interface{}  
Body interface{}  
}

func (self *TCaseBranch) IsValid() bool {
  return true
}

type TMatchBranch struct {
  TNode
Pattern interface{}  
AdditionalPatterns []interface{}  
When interface{}  
Body interface{}  
}

func (self *TMatchBranch) IsValid() bool {
  return true
}

type TVariantCase struct {
  TNode
Name string  
CaseType interface{}  
}

func (self *TVariantCase) IsValid() bool {
  return true
}

type TProgram struct {
  TNode
Name string  
NameToken TToken  
UnitName string  
IsUnit bool  
of interface{}  
String interface{}  
Declarations []interface{}  
Statements []interface{}  
}

func (self *TProgram) IsValid() bool {
  return true
}

type TVarDecl struct {
  TStatement
Token TToken  
Names []string  
VarType interface{}  
Value interface{}  
Inferred bool  
}

func (self *TVarDecl) IsValid() bool {
  return true
}

type TConstDecl struct {
  TStatement
Token TToken  
Name string  
ConstType interface{}  
Value interface{}  
}

func (self *TConstDecl) IsValid() bool {
  return true
}

type TTypeDecl struct {
  TStatement
Token TToken  
Name string  
DeclType interface{}  
}

func (self *TTypeDecl) IsValid() bool {
  return true
}

type TFunctionDecl struct {
  TStatement
Token TToken  
Name string  
TypeParams []*TTypeParameter  
Parameters []*TParameter  
ReturnType interface{}  
ReturnTypes []interface{}  
LocalDecls []interface{}  
Body interface{}  
IsAsync bool  
IsExport bool  
}

func (self *TFunctionDecl) IsValid() bool {
  return true
}

type TClassDecl struct {
  TStatement
Token TToken  
Name string  
TypeParams []*TTypeParameter  
Parent string  
Interfaces []string  
Fields []*TVarDecl  
Methods []*TFunctionDecl  
Properties []interface{}  
Visibility TTokenType  
}

func (self *TClassDecl) IsValid() bool {
  return true
}

type TInterfaceDecl struct {
  TStatement
Token TToken  
Name string  
Parents []string  
Methods []*TFunctionDecl  
}

func (self *TInterfaceDecl) IsValid() bool {
  return true
}

type TPropertyDecl struct {
  TStatement
Token TToken  
Name string  
PropType interface{}  
Getter string  
Setter string  
Default interface{}  
}

func (self *TPropertyDecl) IsValid() bool {
  return true
}

type TBlockStatement struct {
  TStatement
Token TToken  
Statements []interface{}  
}

func (self *TBlockStatement) IsValid() bool {
  return true
}

type TAssignmentStatement struct {
  TStatement
Token TToken  
Name interface{}  
Value interface{}  
}

func (self *TAssignmentStatement) IsValid() bool {
  return true
}

type TReturnStatement struct {
  TStatement
Token TToken  
ReturnValue interface{}  
}

func (self *TReturnStatement) IsValid() bool {
  return true
}

type TExpressionStatement struct {
  TStatement
Token TToken  
Expression interface{}  
}

func (self *TExpressionStatement) IsValid() bool {
  return true
}

type TIfStatement struct {
  TStatement
Token TToken  
Condition interface{}  
Consequence *TBlockStatement  
Alternative *TBlockStatement  
}

func (self *TIfStatement) IsValid() bool {
  return true
}

type TWhileStatement struct {
  TStatement
Token TToken  
Condition interface{}  
Body *TBlockStatement  
}

func (self *TWhileStatement) IsValid() bool {
  return true
}

type TForStatement struct {
  TStatement
Token TToken  
Variable string  
From interface{}  
To interface{}  
DownTo bool  
Body *TBlockStatement  
}

func (self *TForStatement) IsValid() bool {
  return true
}

type TForEachStatement struct {
  TStatement
Token TToken  
Variable string  
Iterable interface{}  
Body *TBlockStatement  
}

func (self *TForEachStatement) IsValid() bool {
  return true
}

type TRepeatStatement struct {
  TStatement
Token TToken  
Body *TBlockStatement  
Condition interface{}  
}

func (self *TRepeatStatement) IsValid() bool {
  return true
}

type TCaseStatement struct {
  TStatement
Token TToken  
Expression interface{}  
Branches []*TCaseBranch  
ElseBranch *TBlockStatement  
}

func (self *TCaseStatement) IsValid() bool {
  return true
}

type TMatchStatement struct {
  TStatement
Token TToken  
Expression interface{}  
Branches []*TMatchBranch  
}

func (self *TMatchStatement) IsValid() bool {
  return true
}

type TOnClause struct {
  TStatement
Token TToken  
Variable string  
OnType interface{}  
Body *TBlockStatement  
}

func (self *TOnClause) IsValid() bool {
  return true
}

type TTryStatement struct {
  TStatement
Token TToken  
Body *TBlockStatement  
OnClauses []*TOnClause  
ExceptBlock *TBlockStatement  
FinallyBlock *TBlockStatement  
}

func (self *TTryStatement) IsValid() bool {
  return true
}

type TRaiseStatement struct {
  TStatement
Token TToken  
Exception interface{}  
}

func (self *TRaiseStatement) IsValid() bool {
  return true
}

type TBreakStatement struct {
  TStatement
Token TToken  
}

func (self *TBreakStatement) IsValid() bool {
  return true
}

type TContinueStatement struct {
  TStatement
Token TToken  
}

func (self *TContinueStatement) IsValid() bool {
  return true
}

type TInheritedStatement struct {
  TStatement
Token TToken  
Expr interface{}  
}

func (self *TInheritedStatement) IsValid() bool {
  return true
}

type TIdentifier struct {
  TExpression
Token TToken  
Value string  
}

func (self *TIdentifier) IsValid() bool {
  return true
}

type TIntegerLiteral struct {
  TExpression
Token TToken  
Value int64  
}

func (self *TIntegerLiteral) IsValid() bool {
  return true
}

type TFloatLiteral struct {
  TExpression
Token TToken  
Value float64  
}

func (self *TFloatLiteral) IsValid() bool {
  return true
}

type TStringLiteral struct {
  TExpression
Token TToken  
Value string  
}

func (self *TStringLiteral) IsValid() bool {
  return true
}

type TStringInterpolation struct {
  TExpression
Parts []interface{}  
}

func (self *TStringInterpolation) IsValid() bool {
  return true
}

type TBooleanLiteral struct {
  TExpression
Token TToken  
Value bool  
}

func (self *TBooleanLiteral) IsValid() bool {
  return true
}

type TNilLiteral struct {
  TExpression
Token TToken  
}

func (self *TNilLiteral) IsValid() bool {
  return true
}

type TArrayLiteral struct {
  TExpression
Token TToken  
Elements []interface{}  
}

func (self *TArrayLiteral) IsValid() bool {
  return true
}

type TTupleLiteral struct {
  TExpression
Token TToken  
Elements []interface{}  
}

func (self *TTupleLiteral) IsValid() bool {
  return true
}

type TPrefixExpression struct {
  TExpression
Token TToken  
Operator string  
Right interface{}  
}

func (self *TPrefixExpression) IsValid() bool {
  return true
}

type TInfixExpression struct {
  TExpression
Token TToken  
Left interface{}  
Operator string  
Right interface{}  
}

func (self *TInfixExpression) IsValid() bool {
  return true
}

type TCallExpression struct {
  TExpression
Token TToken  
Func interface{}  
Arguments []interface{}  
}

func (self *TCallExpression) IsValid() bool {
  return true
}

type TMemberExpression struct {
  TExpression
Token TToken  
Obj interface{}  
Member string  
}

func (self *TMemberExpression) IsValid() bool {
  return true
}

type TIndexExpression struct {
  TExpression
Token TToken  
Left interface{}  
Index interface{}  
}

func (self *TIndexExpression) IsValid() bool {
  return true
}

type TSliceExpression struct {
  TExpression
Token TToken  
Left interface{}  
Low interface{}  
High interface{}  
}

func (self *TSliceExpression) IsValid() bool {
  return true
}

type TLambdaExpression struct {
  TExpression
Token TToken  
Parameters []*TParameter  
Body interface{}  
}

func (self *TLambdaExpression) IsValid() bool {
  return true
}

type TAwaitExpression struct {
  TExpression
Token TToken  
Expression interface{}  
}

func (self *TAwaitExpression) IsValid() bool {
  return true
}

type TTypeCastExpression struct {
  TExpression
Token TToken  
Expression interface{}  
TargetType interface{}  
}

func (self *TTypeCastExpression) IsValid() bool {
  return true
}

type TIsExpression struct {
  TExpression
Token TToken  
Expression interface{}  
TargetType interface{}  
}

func (self *TIsExpression) IsValid() bool {
  return true
}

type TRecordType struct {
  TExpression
Fields []*TVarDecl  
}

func (self *TRecordType) IsValid() bool {
  return true
}

type TArrayType struct {
  TExpression
ElementType interface{}  
Size interface{}  
Dynamic bool  
}

func (self *TArrayType) IsValid() bool {
  return true
}

type TMapType struct {
  TExpression
KeyType interface{}  
ValueType interface{}  
}

func (self *TMapType) IsValid() bool {
  return true
}

type TVariantType struct {
  TExpression
Cases []*TVariantCase  
}

func (self *TVariantType) IsValid() bool {
  return true
}

type TEnumType struct {
  TExpression
Names []string  
}

func (self *TEnumType) IsValid() bool {
  return true
}

type TGenericType struct {
  TExpression
Base string  
TypeParams []interface{}  
}

func (self *TGenericType) IsValid() bool {
  return true
}

type TLexer struct {
Input string  
Position int64  
ReadPosition int64  
Ch string  
Line int64  
Column int64  
}

func (self *TLexer) IsValid() bool {
  return true
}

type TParser struct {
Lex *TLexer  
Errors []string  
CurToken TToken  
PeekToken TToken  
}

func (self *TParser) IsValid() bool {
  return true
}

type TGenerator struct {
Output string  
Indent int64  
Variables map[string]string  
InFunction bool  
InReturnFunc bool  
Imports map[string]bool  
NeedsException bool  
ExceptionTypes map[string]bool  
MultiReturn bool  
MultiReturnN int64  
ClassTypes string  
ClassIsBase map[string]bool  
ClassFields map[string][]string  
NeedFmt bool  
NeedStrings bool  
NeedStrconv bool  
NeedOS bool  
NeedMath bool  
NeedRand bool  
NeedStdlib bool  
UserFuncs string  
}

func (self *TGenerator) Create() {
self.Output = ""  
self.Indent = 0  
self.ClassTypes = ""  
self.UserFuncs = ""  
self.InFunction = false  
self.InReturnFunc = false  
self.NeedsException = false  
self.MultiReturn = false  
self.MultiReturnN = 0  
self.NeedFmt = false  
self.NeedStrings = false  
self.NeedStrconv = false  
self.NeedOS = false  
self.NeedMath = false  
self.NeedRand = false  
self.NeedStdlib = false  
}

func (self *TGenerator) Write(s string) {
self.Output = (self.Output + s)  
}

func (self *TGenerator) WriteLine(s string) {
var i int64  
for i = 0; i <= (self.Indent - 1); i  ++ {
self.Output = (self.Output + "  ")    
  }
self.Output = ((self.Output + s) + "\n")  
  _ = i
}

func (self *TGenerator) WriteEscapedGoString(s string) {
var j int64  
var n int64  
var ch string  
var nx string  
n = int64(len(s))  
j = 0  
for (j < n)   {
ch = s[j:(j + 1)]    
if (ch == "\\")     {
if ((j + 1) < n)       {
nx = s[(j + 1):(j + 2)]        
if (((nx == "n") || (nx == "t")) || (nx == "r"))         {
self.Write("\\")          
self.Write(nx)          
j = (j + 2)          
        } else {
self.Write(("\\" + "\\"))          
j = (j + 1)          
        }
      } else {
self.Write(("\\" + "\\"))        
j = (j + 1)        
      }
    } else {
if (ch == "\"")       {
self.Write(("\\" + "\""))        
j = (j + 1)        
      } else {
self.Write(ch)        
j = (j + 1)        
      }
    }
  }
}

func (self *TGenerator) IncreaseIndent() {
self.Indent = (self.Indent + 1)  
}

func (self *TGenerator) DecreaseIndent() {
self.Indent = (self.Indent - 1)  
}

func (self *TGenerator) Generate(prog *TProgram) string {
var result string  
self.CollectClassTypes(prog)  
self.ScanImports(prog)  
self.ScanForException(prog)  
self.GenerateTypes(prog)  
self.GenerateGlobals(prog)  
self.GenerateFunctions(prog)  
if (prog.IsUnit == false)   {
if (int64(len(prog.Statements)) > 0)     {
self.WriteLine("func main() {")      
self.IncreaseIndent      ()
self.GenerateStatements(prog.Statements)      
self.DecreaseIndent      ()
self.WriteLine("}")      
    }
  }
self.GenerateExceptionTypes  ()
self.CollectImports  ()
if prog.IsUnit   {
result = ((((("package main // unit: " + prog.UnitName) + "\n") + "\n") + self.BuildImportBlock()) + self.Output)    
  } else {
result = (((("package main" + "\n") + "\n") + self.BuildImportBlock()) + self.Output)    
  }
  return result
}

func (self *TGenerator) GenerateMulti(progs []*TProgram) string {
var result string  
var pi int64  
var prog *TProgram  
for pi = 0; pi <= (int64(len(progs)) - 1); pi  ++ {
prog = progs[pi]    
self.CollectClassTypes(prog)    
self.ScanImports(prog)    
self.ScanForException(prog)    
  }
for pi = 0; pi <= (int64(len(progs)) - 1); pi  ++ {
prog = progs[pi]    
self.GenerateTypes(prog)    
  }
for pi = 0; pi <= (int64(len(progs)) - 1); pi  ++ {
prog = progs[pi]    
self.GenerateGlobals(prog)    
  }
for pi = 0; pi <= (int64(len(progs)) - 1); pi  ++ {
prog = progs[pi]    
self.GenerateFunctions(prog)    
  }
var hasMain bool  
hasMain = false  
for pi = 0; pi <= (int64(len(progs)) - 1); pi  ++ {
prog = progs[pi]    
if ((prog.IsUnit == false) && (int64(len(prog.Statements)) > 0))     {
self.WriteLine("func main() {")      
self.IncreaseIndent      ()
self.GenerateStatements(prog.Statements)      
self.DecreaseIndent      ()
self.WriteLine("}")      
hasMain = true      
    }
  }
if (hasMain == false)   {
self.WriteLine("func main() {}")    
  }
self.GenerateExceptionTypes  ()
self.CollectImports  ()
result = (((("package main" + "\n") + "\n") + self.BuildImportBlock()) + self.Output)  
  _ = pi
  _ = prog
  return result
}

func (self *TGenerator) BuildImportBlock() string {
var result string  
var s string  
s = ("import (" + "\n")  
if self.NeedFmt   {
s = ((s + "  \"fmt\"") + "\n")    
  }
if self.NeedStrings   {
s = ((s + "  \"strings\"") + "\n")    
  }
if self.NeedStrconv   {
s = ((s + "  \"strconv\"") + "\n")    
  }
if self.NeedOS   {
s = ((s + "  \"os\"") + "\n")    
  }
if self.NeedMath   {
s = ((s + "  \"math\"") + "\n")    
  }
if self.NeedRand   {
s = ((s + "  \"math/rand\"") + "\n")    
  }
if self.NeedStdlib   {
s = ((s + "  \"kylix/stdlib\"") + "\n")    
  }
s = (((s + ")") + "\n") + "\n")  
result = s  
  _ = s
  return result
}

func (self *TGenerator) StrContains(haystack string, needle string) bool {
var result bool  
var i int64  
var nl int64  
var found bool  
found = false  
nl = int64(len(needle))  
if (nl == 0)   {
found = true    
  } else {
if (int64(len(haystack)) >= nl)     {
i = 0      
for ((found == false) && (i <= (int64(len(haystack)) - nl)))       {
if (haystack[i:(i + nl)] == needle)         {
found = true          
        } else {
i = (i + 1)          
        }
      }
    }
  }
result = found  
  _ = found
  return result
}

func (self *TGenerator) CollectImports() {
if self.StrContains(self.Output, ("fmt" + "."))   {
self.NeedFmt = true    
  }
if self.StrContains(self.Output, ("strings" + "."))   {
self.NeedStrings = true    
  }
if self.StrContains(self.Output, ("strconv" + "."))   {
self.NeedStrconv = true    
  }
if self.StrContains(self.Output, ("os" + ".Args"))   {
self.NeedOS = true    
  }
if self.StrContains(self.Output, ("math" + "."))   {
self.NeedMath = true    
  }
if self.StrContains(self.Output, ("rand" + "."))   {
self.NeedRand = true    
  }
}

func (self *TGenerator) CollectClassTypes(prog *TProgram) {
var i int64  
var decl interface{}  
var cd *TClassDecl  
var td *TTypeDecl  
for i = 0; i <= (int64(len(prog.Declarations)) - 1); i  ++ {
decl = prog.Declarations[i]    
if func() bool { _, ok := decl.(*TClassDecl); return ok }()     {
cd = decl.(*TClassDecl)      
self.ClassTypes = (((self.ClassTypes + ",") + cd.Name) + ",")      
    } else {
if func() bool { _, ok := decl.(*TTypeDecl); return ok }()       {
td = decl.(*TTypeDecl)        
if func() bool { _, ok := td.DeclType.(*TClassDecl); return ok }()         {
cd = td.DeclType.(*TClassDecl)          
self.ClassTypes = (((self.ClassTypes + ",") + td.Name) + ",")          
        }
      } else {
if func() bool { _, ok := decl.(*TFunctionDecl); return ok }()         {
var fd *TFunctionDecl          
fd = decl.(*TFunctionDecl)          
if (fd.Name != "")           {
self.UserFuncs = (((self.UserFuncs + ",") + fd.Name) + ",")            
          }
        }
      }
    }
  }
  _ = i
  _ = decl
  _ = cd
  _ = td
}

func (self *TGenerator) ScanImports(prog *TProgram) {
}

func (self *TGenerator) ScanForException(prog *TProgram) {
}

func (self *TGenerator) GenerateTypes(prog *TProgram) {
var i int64  
var decl interface{}  
var td *TTypeDecl  
var cd *TClassDecl  
var id *TInterfaceDecl  
var et *TEnumType  
for i = 0; i <= (int64(len(prog.Declarations)) - 1); i  ++ {
decl = prog.Declarations[i]    
if func() bool { _, ok := decl.(*TTypeDecl); return ok }()     {
td = decl.(*TTypeDecl)      
self.GenerateTypeDecl(td)      
    } else {
if func() bool { _, ok := decl.(*TClassDecl); return ok }()       {
cd = decl.(*TClassDecl)        
self.GenerateClassDecl(cd)        
      } else {
if func() bool { _, ok := decl.(*TInterfaceDecl); return ok }()         {
id = decl.(*TInterfaceDecl)          
self.GenerateInterfaceDecl(id)          
        }
      }
    }
  }
  _ = i
  _ = decl
  _ = td
  _ = cd
  _ = id
  _ = et
}

func (self *TGenerator) GenerateTypeDecl(td *TTypeDecl) {
var et *TEnumType  
var rt *TRecordType  
var innerCd *TClassDecl  
var innerId *TInterfaceDecl  
var i int64  
if func() bool { _, ok := td.DeclType.(*TEnumType); return ok }()   {
et = td.DeclType.(*TEnumType)    
self.GenerateEnumType(td.Name, et)    
  } else {
if func() bool { _, ok := td.DeclType.(*TRecordType); return ok }()     {
rt = td.DeclType.(*TRecordType)      
self.GenerateRecordType(td.Name, rt)      
    } else {
if func() bool { _, ok := td.DeclType.(*TClassDecl); return ok }()       {
innerCd = td.DeclType.(*TClassDecl)        
innerCd.Name = td.Name        
self.GenerateClassDecl(innerCd)        
      } else {
if func() bool { _, ok := td.DeclType.(*TInterfaceDecl); return ok }()         {
innerId = td.DeclType.(*TInterfaceDecl)          
innerId.Name = td.Name          
self.GenerateInterfaceDecl(innerId)          
        } else {
self.Write((("type " + td.Name) + " "))          
self.GenerateTypeExpression(td.DeclType)          
self.WriteLine("")          
self.WriteLine("")          
        }
      }
    }
  }
  _ = et
  _ = rt
  _ = innerCd
  _ = innerId
  _ = i
}

func (self *TGenerator) GenerateEnumType(name string, enum *TEnumType) {
var i int64  
if (int64(len(enum.Names)) == 0)   {
    return
  }
self.WriteLine("const (")  
self.IncreaseIndent  ()
for i = 0; i <= (int64(len(enum.Names)) - 1); i  ++ {
if (i == 0)     {
self.WriteLine((((enum.Names[i] + " ") + name) + " = iota"))      
    } else {
self.WriteLine(enum.Names[i])      
    }
  }
self.DecreaseIndent  ()
self.WriteLine(")")  
self.WriteLine("")  
self.WriteLine((("type " + name) + " int"))  
self.WriteLine("")  
  _ = i
}

func (self *TGenerator) GenerateRecordType(name string, rec *TRecordType) {
var i int64  
var j int64  
var field *TVarDecl  
self.WriteLine((("type " + name) + " struct {"))  
self.IncreaseIndent  ()
for i = 0; i <= (int64(len(rec.Fields)) - 1); i  ++ {
field = rec.Fields[i]    
for j = 0; j <= (int64(len(field.Names)) - 1); j    ++ {
self.Write((field.Names[j] + " "))      
if (field.VarType != nil)       {
self.GenerateTypeExpression(field.VarType)        
      } else {
self.Write("interface{}")        
      }
self.WriteLine("")      
    }
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
self.WriteLine("")  
  _ = i
  _ = j
  _ = field
}

func (self *TGenerator) GenerateInlineRecordType(rec *TRecordType) {
var i int64  
var j int64  
var field *TVarDecl  
self.Write("struct {")  
for i = 0; i <= (int64(len(rec.Fields)) - 1); i  ++ {
field = rec.Fields[i]    
for j = 0; j <= (int64(len(field.Names)) - 1); j    ++ {
self.Write((field.Names[j] + " "))      
if (field.VarType != nil)       {
self.GenerateTypeExpression(field.VarType)        
      } else {
self.Write("interface{}")        
      }
self.Write("; ")      
    }
  }
self.Write("}")  
  _ = i
  _ = j
  _ = field
}

func (self *TGenerator) GenerateClassDecl(cd *TClassDecl) {
var i int64  
var j int64  
var field *TVarDecl  
var method *TFunctionDecl  
var prop *TPropertyDecl  
self.Write(("type " + cd.Name))  
self.GenerateTypeParams(cd.TypeParams)  
self.WriteLine(" struct {")  
self.IncreaseIndent  ()
if (cd.Parent != "")   {
self.WriteLine(cd.Parent)    
  }
for i = 0; i <= (int64(len(cd.Fields)) - 1); i  ++ {
field = cd.Fields[i]    
for j = 0; j <= (int64(len(field.Names)) - 1); j    ++ {
self.Write((field.Names[j] + " "))      
if (field.VarType != nil)       {
self.GenerateTypeExpression(field.VarType)        
      } else {
self.Write("interface{}")        
      }
self.WriteLine("")      
    }
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
self.WriteLine("")  
for i = 0; i <= (int64(len(cd.Methods)) - 1); i  ++ {
method = cd.Methods[i]    
self.GenerateClassMethod(cd.Name, cd.TypeParams, method)    
  }
for i = 0; i <= (int64(len(cd.Properties)) - 1); i  ++ {
prop = cd.Properties[i].(*TPropertyDecl)    
self.GeneratePropertyAccessors(cd.Name, prop)    
  }
self.Write(("func (self *" + cd.Name))  
if (int64(len(cd.TypeParams)) > 0)   {
self.Write("[")    
for i = 0; i <= (int64(len(cd.TypeParams)) - 1); i    ++ {
if (i > 0)       {
self.Write(", ")        
      }
self.Write(cd.TypeParams[i].Name)      
    }
self.Write("]")    
  }
self.WriteLine(") IsValid() bool {")  
self.IncreaseIndent  ()
self.WriteLine("return true")  
self.DecreaseIndent  ()
self.WriteLine("}")  
self.WriteLine("")  
  _ = i
  _ = j
  _ = field
  _ = method
  _ = prop
}

func (self *TGenerator) GenerateClassMethod(className string, typeParams []*TTypeParameter, method *TFunctionDecl) {
var hasReturn bool  
var i int64  
var stmt interface{}  
var body *TBlockStatement  
var local interface{}  
var localVd *TVarDecl  
var localCd *TConstDecl  
hasReturn = ((method.ReturnType != nil) || (int64(len(method.ReturnTypes)) > 0))  
self.Write(("func (self *" + className))  
if (int64(len(typeParams)) > 0)   {
self.Write("[")    
for i = 0; i <= (int64(len(typeParams)) - 1); i    ++ {
if (i > 0)       {
self.Write(", ")        
      }
self.Write(typeParams[i].Name)      
    }
self.Write("]")    
  }
self.Write((") " + method.Name))  
self.GenerateFunctionSignature(method)  
self.WriteLine(" {")  
self.IncreaseIndent  ()
if hasReturn   {
if (method.ReturnType != nil)     {
self.Write("var result ")      
self.GenerateTypeExpression(method.ReturnType)      
self.WriteLine("")      
    }
  }
for i = 0; i <= (int64(len(method.LocalDecls)) - 1); i  ++ {
local = method.LocalDecls[i]    
if func() bool { _, ok := local.(*TVarDecl); return ok }()     {
localVd = local.(*TVarDecl)      
self.GenerateLocalVarDecl(localVd)      
    } else {
if func() bool { _, ok := local.(*TConstDecl); return ok }()       {
localCd = local.(*TConstDecl)        
self.GenerateLocalConstDecl(localCd)        
      }
    }
  }
if (method.Body != nil)   {
body = method.Body.(*TBlockStatement)    
self.InFunction = true    
self.InReturnFunc = hasReturn    
for i = 0; i <= (int64(len(body.Statements)) - 1); i    ++ {
stmt = body.Statements[i]      
self.GenerateStatement(stmt)      
    }
self.InFunction = false    
self.InReturnFunc = false    
  }
for i = 0; i <= (int64(len(method.LocalDecls)) - 1); i  ++ {
local = method.LocalDecls[i]    
if func() bool { _, ok := local.(*TVarDecl); return ok }()     {
localVd = local.(*TVarDecl)      
if (int64(len(localVd.Names)) == 1)       {
self.WriteLine(("_ = " + localVd.Names[0]))        
      }
    }
  }
if hasReturn   {
self.WriteLine("return result")    
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
self.WriteLine("")  
  _ = hasReturn
  _ = i
  _ = stmt
  _ = body
  _ = local
  _ = localVd
  _ = localCd
}

func (self *TGenerator) GeneratePropertyAccessors(className string, prop *TPropertyDecl) {
if (prop.Getter != "")   {
self.Write((((("func (self *" + className) + ") ") + prop.Name) + "() "))    
if (prop.PropType != nil)     {
self.GenerateTypeExpression(prop.PropType)      
    } else {
self.Write("interface{}")      
    }
self.WriteLine(" {")    
self.IncreaseIndent    ()
self.WriteLine(("return self." + prop.Getter))    
self.DecreaseIndent    ()
self.WriteLine("}")    
self.WriteLine("")    
  }
if (prop.Setter != "")   {
self.Write((((("func (self *" + className) + ") Set") + prop.Name) + "(v "))    
if (prop.PropType != nil)     {
self.GenerateTypeExpression(prop.PropType)      
    } else {
self.Write("interface{}")      
    }
self.WriteLine(") {")    
self.IncreaseIndent    ()
self.WriteLine((("self." + prop.Setter) + " = v"))    
self.DecreaseIndent    ()
self.WriteLine("}")    
self.WriteLine("")    
  }
}

func (self *TGenerator) GenerateInterfaceDecl(id *TInterfaceDecl) {
var i int64  
var parent string  
var method *TFunctionDecl  
self.WriteLine((("type " + id.Name) + " interface {"))  
self.IncreaseIndent  ()
for i = 0; i <= (int64(len(id.Parents)) - 1); i  ++ {
parent = id.Parents[i]    
self.WriteLine(parent)    
  }
for i = 0; i <= (int64(len(id.Methods)) - 1); i  ++ {
method = id.Methods[i]    
self.Write(method.Name)    
self.GenerateFunctionSignature(method)    
self.WriteLine("")    
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
self.WriteLine("")  
  _ = i
  _ = parent
  _ = method
}

func (self *TGenerator) GenerateExceptionTypes() {
if self.NeedsException   {
self.WriteLine("type Exception struct {")    
self.IncreaseIndent    ()
self.WriteLine("Message string")    
self.DecreaseIndent    ()
self.WriteLine("}")    
self.WriteLine("")    
  }
}

func (self *TGenerator) GenerateGlobals(prog *TProgram) {
var i int64  
var decl interface{}  
var vd *TVarDecl  
var cd *TConstDecl  
for i = 0; i <= (int64(len(prog.Declarations)) - 1); i  ++ {
decl = prog.Declarations[i]    
if func() bool { _, ok := decl.(*TVarDecl); return ok }()     {
vd = decl.(*TVarDecl)      
self.GenerateGlobalVarDecl(vd)      
    } else {
if func() bool { _, ok := decl.(*TConstDecl); return ok }()       {
cd = decl.(*TConstDecl)        
self.GenerateConstDecl(cd)        
      }
    }
  }
  _ = i
  _ = decl
  _ = vd
  _ = cd
}

func (self *TGenerator) GenerateGlobalVarDecl(vd *TVarDecl) {
var i int64  
for i = 0; i <= (int64(len(vd.Names)) - 1); i  ++ {
self.Write(("var " + vd.Names[i]))    
if (vd.VarType != nil)     {
self.Write(" ")      
self.GenerateTypeExpression(vd.VarType)      
    }
if (vd.Value != nil)     {
self.Write(" = ")      
self.GenerateExpression(vd.Value)      
    } else {
if func() bool { _, ok := vd.VarType.(*TMapType); return ok }()       {
self.Write(" = ")        
self.GenerateTypeExpression(vd.VarType)        
self.Write("{}")        
      }
    }
self.WriteLine("")    
  }
  _ = i
}

func (self *TGenerator) GenerateConstDecl(cd *TConstDecl) {
self.Write(("const " + cd.Name))  
if (cd.ConstType != nil)   {
self.Write(" ")    
self.GenerateTypeExpression(cd.ConstType)    
  }
if (cd.Value != nil)   {
self.Write(" = ")    
self.GenerateExpression(cd.Value)    
  }
self.WriteLine("")  
}

func (self *TGenerator) GenerateLocalVarDecl(vd *TVarDecl) {
var i int64  
for i = 0; i <= (int64(len(vd.Names)) - 1); i  ++ {
self.Write((("var " + vd.Names[i]) + " "))    
if (vd.VarType != nil)     {
self.GenerateTypeExpression(vd.VarType)      
    } else {
self.Write("interface{}")      
    }
self.WriteLine("")    
  }
  _ = i
}

func (self *TGenerator) GenerateLocalConstDecl(cd *TConstDecl) {
self.Write(("const " + cd.Name))  
if (cd.ConstType != nil)   {
self.Write(" ")    
self.GenerateTypeExpression(cd.ConstType)    
  }
self.Write(" = ")  
if (cd.Value != nil)   {
self.GenerateExpression(cd.Value)    
  }
self.WriteLine("")  
}

func (self *TGenerator) GenerateFunctions(prog *TProgram) {
var i int64  
var decl interface{}  
var fd *TFunctionDecl  
for i = 0; i <= (int64(len(prog.Declarations)) - 1); i  ++ {
decl = prog.Declarations[i]    
if func() bool { _, ok := decl.(*TFunctionDecl); return ok }()     {
fd = decl.(*TFunctionDecl)      
self.GenerateFunctionDecl(fd)      
    }
  }
  _ = i
  _ = decl
  _ = fd
}

func (self *TGenerator) GenerateFunctionDecl(fd *TFunctionDecl) {
var hasReturn bool  
var i int64  
var stmt interface{}  
var body *TBlockStatement  
var dotPos int64  
var className string  
var methodName string  
var fullName string  
if (fd.Body == nil)   {
    return
  }
hasReturn = ((fd.ReturnType != nil) || (int64(len(fd.ReturnTypes)) > 0))  
self.MultiReturn = (int64(len(fd.ReturnTypes)) > 1)  
fullName = fd.Name  
dotPos = 0  
for i = 0; i <= (int64(len(fullName)) - 1); i  ++ {
if (fullName[i:(i + 1)] == ".")     {
dotPos = i      
      break
    }
  }
if (dotPos > 0)   {
className = fullName[0:dotPos]    
methodName = fullName[(dotPos + 1):int64(len(fullName))]    
self.Write(((("func (self *" + className) + ") ") + methodName))    
  } else {
self.Write(("func " + fd.Name))    
  }
self.GenerateFunctionSignature(fd)  
self.WriteLine(" {")  
self.IncreaseIndent  ()
if hasReturn   {
if (fd.ReturnType != nil)     {
self.Write("var result ")      
self.GenerateTypeExpression(fd.ReturnType)      
self.WriteLine("")      
    }
  }
for i = 0; i <= (int64(len(fd.LocalDecls)) - 1); i  ++ {
var local interface{}    
var localVd *TVarDecl    
var localCd *TConstDecl    
local = fd.LocalDecls[i]    
if func() bool { _, ok := local.(*TVarDecl); return ok }()     {
localVd = local.(*TVarDecl)      
self.GenerateLocalVarDecl(localVd)      
    } else {
if func() bool { _, ok := local.(*TConstDecl); return ok }()       {
localCd = local.(*TConstDecl)        
self.GenerateLocalConstDecl(localCd)        
      }
    }
  }
if (fd.Body != nil)   {
body = fd.Body.(*TBlockStatement)    
self.InFunction = true    
self.InReturnFunc = hasReturn    
for i = 0; i <= (int64(len(body.Statements)) - 1); i    ++ {
stmt = body.Statements[i]      
self.GenerateStatement(stmt)      
    }
self.InFunction = false    
self.InReturnFunc = false    
  }
for i = 0; i <= (int64(len(fd.LocalDecls)) - 1); i  ++ {
var local2 interface{}    
var localVd2 *TVarDecl    
local2 = fd.LocalDecls[i]    
if func() bool { _, ok := local2.(*TVarDecl); return ok }()     {
localVd2 = local2.(*TVarDecl)      
if (int64(len(localVd2.Names)) == 1)       {
self.WriteLine(("_ = " + localVd2.Names[0]))        
      }
    }
  }
if hasReturn   {
if (self.MultiReturn == false)     {
self.WriteLine("return result")      
    }
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
self.WriteLine("")  
  _ = hasReturn
  _ = i
  _ = stmt
  _ = body
  _ = dotPos
  _ = className
  _ = methodName
  _ = fullName
}

func (self *TGenerator) GenerateFunctionSignature(fd *TFunctionDecl) {
var i int64  
var param *TParameter  
self.Write("(")  
for i = 0; i <= (int64(len(fd.Parameters)) - 1); i  ++ {
if (i > 0)     {
self.Write(", ")      
    }
param = fd.Parameters[i]    
self.Write((param.Name + " "))    
if (param.ParamType != nil)     {
self.GenerateTypeExpression(param.ParamType)      
    } else {
self.Write("interface{}")      
    }
  }
self.Write(")")  
if (fd.ReturnType != nil)   {
self.Write(" ")    
self.GenerateTypeExpression(fd.ReturnType)    
  } else {
if (int64(len(fd.ReturnTypes)) > 0)     {
self.Write(" (")      
for i = 0; i <= (int64(len(fd.ReturnTypes)) - 1); i      ++ {
if (i > 0)         {
self.Write(", ")          
        }
self.GenerateTypeExpression(fd.ReturnTypes[i])        
      }
self.Write(")")      
    }
  }
  _ = i
  _ = param
}

func (self *TGenerator) GenerateTypeParams(params []*TTypeParameter) {
var i int64  
var tp *TTypeParameter  
if (int64(len(params)) == 0)   {
    return
  }
self.Write("[")  
for i = 0; i <= (int64(len(params)) - 1); i  ++ {
if (i > 0)     {
self.Write(", ")      
    }
tp = params[i]    
self.Write((tp.Name + " any"))    
  }
self.Write("]")  
  _ = i
  _ = tp
}

func (self *TGenerator) GenerateStatements(stmts []interface{}) {
var i int64  
for i = 0; i <= (int64(len(stmts)) - 1); i  ++ {
self.GenerateStatement(stmts[i])    
  }
  _ = i
}

func (self *TGenerator) GenerateStatement(stmt interface{}) {
var vd *TVarDecl  
var asgn *TAssignmentStatement  
var es *TExpressionStatement  
var ifStmt *TIfStatement  
var whileStmt *TWhileStatement  
var forStmt *TForStatement  
var forEachStmt *TForEachStatement  
var repeatStmt *TRepeatStatement  
var caseStmt *TCaseStatement  
var matchStmt *TMatchStatement  
var tryStmt *TTryStatement  
var raiseStmt *TRaiseStatement  
var retStmt *TReturnStatement  
var block *TBlockStatement  
var call *TCallExpression  
var ident *TIdentifier  
var member *TMemberExpression  
var i int64  
if func() bool { _, ok := stmt.(*TVarDecl); return ok }()   {
vd = stmt.(*TVarDecl)    
self.GenerateVarDecl(vd)    
  } else {
if func() bool { _, ok := stmt.(*TAssignmentStatement); return ok }()     {
asgn = stmt.(*TAssignmentStatement)      
self.GenerateAssignment(asgn)      
    } else {
if func() bool { _, ok := stmt.(*TExpressionStatement); return ok }()       {
es = stmt.(*TExpressionStatement)        
if func() bool { _, ok := es.Expression.(*TIdentifier); return ok }()         {
ident = es.Expression.(*TIdentifier)          
if ((ident.Value == "Exit") || (ident.Value == "exit"))           {
if self.InReturnFunc             {
self.WriteLine("return result")              
            } else {
self.WriteLine("return")              
            }
            return
          }
        }
if func() bool { _, ok := es.Expression.(*TMemberExpression); return ok }()         {
self.GenerateExpression(es.Expression)          
self.WriteLine("()")          
          return
        }
if func() bool { _, ok := es.Expression.(*TCallExpression); return ok }()         {
call = es.Expression.(*TCallExpression)          
if func() bool { _, ok := call.Func.(*TIdentifier); return ok }()           {
ident = call.Func.(*TIdentifier)            
if (ident.Value == "append")             {
self.GenerateExpression(call.Arguments[0])              
self.Write(" = append(")              
for i = 0; i <= (int64(len(call.Arguments)) - 1); i              ++ {
if (i > 0)                 {
self.Write(", ")                  
                }
self.GenerateExpression(call.Arguments[i])                
              }
self.WriteLine(")")              
              return
            }
          }
        }
self.GenerateExpression(es.Expression)        
self.WriteLine("")        
      } else {
if func() bool { _, ok := stmt.(*TIfStatement); return ok }()         {
ifStmt = stmt.(*TIfStatement)          
self.GenerateIfStatement(ifStmt)          
        } else {
if func() bool { _, ok := stmt.(*TWhileStatement); return ok }()           {
whileStmt = stmt.(*TWhileStatement)            
self.GenerateWhileStatement(whileStmt)            
          } else {
if func() bool { _, ok := stmt.(*TForStatement); return ok }()             {
forStmt = stmt.(*TForStatement)              
self.GenerateForStatement(forStmt)              
            } else {
if func() bool { _, ok := stmt.(*TForEachStatement); return ok }()               {
forEachStmt = stmt.(*TForEachStatement)                
self.GenerateForEachStatement(forEachStmt)                
              } else {
if func() bool { _, ok := stmt.(*TRepeatStatement); return ok }()                 {
repeatStmt = stmt.(*TRepeatStatement)                  
self.GenerateRepeatStatement(repeatStmt)                  
                } else {
if func() bool { _, ok := stmt.(*TCaseStatement); return ok }()                   {
caseStmt = stmt.(*TCaseStatement)                    
self.GenerateCaseStatement(caseStmt)                    
                  } else {
if func() bool { _, ok := stmt.(*TMatchStatement); return ok }()                     {
matchStmt = stmt.(*TMatchStatement)                      
self.GenerateMatchStatement(matchStmt)                      
                    } else {
if func() bool { _, ok := stmt.(*TTryStatement); return ok }()                       {
tryStmt = stmt.(*TTryStatement)                        
self.GenerateTryStatement(tryStmt)                        
                      } else {
if func() bool { _, ok := stmt.(*TRaiseStatement); return ok }()                         {
raiseStmt = stmt.(*TRaiseStatement)                          
self.GenerateRaiseStatement(raiseStmt)                          
                        } else {
if func() bool { _, ok := stmt.(*TReturnStatement); return ok }()                           {
retStmt = stmt.(*TReturnStatement)                            
self.GenerateReturnStatement(retStmt)                            
                          } else {
if func() bool { _, ok := stmt.(*TBreakStatement); return ok }()                             {
self.WriteLine("break")                              
                            } else {
if func() bool { _, ok := stmt.(*TContinueStatement); return ok }()                               {
self.WriteLine("continue")                                
                              } else {
if func() bool { _, ok := stmt.(*TBlockStatement); return ok }()                                 {
block = stmt.(*TBlockStatement)                                  
for i = 0; i <= (int64(len(block.Statements)) - 1); i                                  ++ {
self.GenerateStatement(block.Statements[i])                                    
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  _ = vd
  _ = asgn
  _ = es
  _ = ifStmt
  _ = whileStmt
  _ = forStmt
  _ = forEachStmt
  _ = repeatStmt
  _ = caseStmt
  _ = matchStmt
  _ = tryStmt
  _ = raiseStmt
  _ = retStmt
  _ = block
  _ = call
  _ = ident
  _ = member
  _ = i
}

func (self *TGenerator) GenerateVarDecl(vd *TVarDecl) {
var i int64  
for i = 0; i <= (int64(len(vd.Names)) - 1); i  ++ {
if vd.Inferred     {
self.Write((vd.Names[i] + " := "))      
if (vd.Value != nil)       {
self.GenerateExpression(vd.Value)        
      }
    } else {
self.Write(("var " + vd.Names[i]))      
if (vd.VarType != nil)       {
self.Write(" ")        
self.GenerateTypeExpression(vd.VarType)        
      }
if (vd.Value != nil)       {
self.Write(" = ")        
self.GenerateExpression(vd.Value)        
      }
    }
self.WriteLine("")    
  }
  _ = i
}

func (self *TGenerator) GenerateAssignment(asgn *TAssignmentStatement) {
var i int64  
var tuple *TTupleLiteral  
var elem interface{}  
var n *TIdentifier  
if self.MultiReturn   {
if func() bool { _, ok := asgn.Name.(*TIdentifier); return ok }()     {
n = asgn.Name.(*TIdentifier)      
if (n.Value == "result")       {
self.Write("return ")        
if func() bool { _, ok := asgn.Value.(*TTupleLiteral); return ok }()         {
tuple = asgn.Value.(*TTupleLiteral)          
for i = 0; i <= (int64(len(tuple.Elements)) - 1); i          ++ {
if (i > 0)             {
self.Write(", ")              
            }
elem = tuple.Elements[i]            
self.GenerateExpression(elem)            
          }
        } else {
self.GenerateExpression(asgn.Value)          
        }
self.WriteLine("")        
        return
      }
    }
  }
if func() bool { _, ok := asgn.Name.(*TTupleLiteral); return ok }()   {
tuple = asgn.Name.(*TTupleLiteral)    
for i = 0; i <= (int64(len(tuple.Elements)) - 1); i    ++ {
if (i > 0)       {
self.Write(", ")        
      }
elem = tuple.Elements[i]      
self.GenerateExpression(elem)      
    }
self.Write(" := ")    
self.GenerateExpression(asgn.Value)    
self.WriteLine("")    
    return
  }
self.GenerateExpression(asgn.Name)  
self.Write(" = ")  
self.GenerateExpression(asgn.Value)  
self.WriteLine("")  
  _ = i
  _ = tuple
  _ = elem
  _ = n
}

func (self *TGenerator) GenerateIfStatement(stmt *TIfStatement) {
var i int64  
self.Write("if ")  
self.GenerateExpression(stmt.Condition)  
self.WriteLine(" {")  
self.IncreaseIndent  ()
if (stmt.Consequence != nil)   {
for i = 0; i <= (int64(len(stmt.Consequence.Statements)) - 1); i    ++ {
self.GenerateStatement(stmt.Consequence.Statements[i])      
    }
  }
self.DecreaseIndent  ()
if (stmt.Alternative != nil)   {
self.WriteLine("} else {")    
self.IncreaseIndent    ()
for i = 0; i <= (int64(len(stmt.Alternative.Statements)) - 1); i    ++ {
self.GenerateStatement(stmt.Alternative.Statements[i])      
    }
self.DecreaseIndent    ()
  }
self.WriteLine("}")  
  _ = i
}

func (self *TGenerator) GenerateWhileStatement(stmt *TWhileStatement) {
var i int64  
self.Write("for ")  
self.GenerateExpression(stmt.Condition)  
self.WriteLine(" {")  
self.IncreaseIndent  ()
if (stmt.Body != nil)   {
for i = 0; i <= (int64(len(stmt.Body.Statements)) - 1); i    ++ {
self.GenerateStatement(stmt.Body.Statements[i])      
    }
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
  _ = i
}

func (self *TGenerator) GenerateForStatement(stmt *TForStatement) {
var i int64  
self.Write((("for " + stmt.Variable) + " = "))  
self.GenerateExpression(stmt.From)  
self.Write(("; " + stmt.Variable))  
if stmt.DownTo   {
self.Write(" >= ")    
  } else {
self.Write(" <= ")    
  }
self.GenerateExpression(stmt.To)  
self.Write(("; " + stmt.Variable))  
if stmt.DownTo   {
self.WriteLine("-- {")    
  } else {
self.WriteLine("++ {")    
  }
self.IncreaseIndent  ()
if (stmt.Body != nil)   {
for i = 0; i <= (int64(len(stmt.Body.Statements)) - 1); i    ++ {
self.GenerateStatement(stmt.Body.Statements[i])      
    }
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
  _ = i
}

func (self *TGenerator) GenerateForEachStatement(stmt *TForEachStatement) {
var i int64  
self.Write((("for _, " + stmt.Variable) + " := range "))  
self.GenerateExpression(stmt.Iterable)  
self.WriteLine(" {")  
self.IncreaseIndent  ()
if (stmt.Body != nil)   {
for i = 0; i <= (int64(len(stmt.Body.Statements)) - 1); i    ++ {
self.GenerateStatement(stmt.Body.Statements[i])      
    }
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
  _ = i
}

func (self *TGenerator) GenerateRepeatStatement(stmt *TRepeatStatement) {
var i int64  
self.WriteLine("for {")  
self.IncreaseIndent  ()
if (stmt.Body != nil)   {
for i = 0; i <= (int64(len(stmt.Body.Statements)) - 1); i    ++ {
self.GenerateStatement(stmt.Body.Statements[i])      
    }
  }
self.Write("if ")  
self.GenerateExpression(stmt.Condition)  
self.WriteLine(" {")  
self.IncreaseIndent  ()
self.WriteLine("break")  
self.DecreaseIndent  ()
self.WriteLine("}")  
self.DecreaseIndent  ()
self.WriteLine("}")  
  _ = i
}

func (self *TGenerator) GenerateCaseStatement(stmt *TCaseStatement) {
var i int64  
var j int64  
var branch *TCaseBranch  
self.Write("switch ")  
self.GenerateExpression(stmt.Expression)  
self.WriteLine(" {")  
self.IncreaseIndent  ()
for i = 0; i <= (int64(len(stmt.Branches)) - 1); i  ++ {
branch = stmt.Branches[i]    
for j = 0; j <= (int64(len(branch.Values)) - 1); j    ++ {
self.Write("case ")      
self.GenerateExpression(branch.Values[j])      
self.WriteLine(":")      
    }
self.IncreaseIndent    ()
if func() bool { _, ok := branch.Body.(*TBlockStatement); return ok }()     {
var blk2 *TBlockStatement      
blk2 = branch.Body.(*TBlockStatement)      
for j = 0; j <= (int64(len(blk2.Statements)) - 1); j      ++ {
self.GenerateStatement(blk2.Statements[j])        
      }
    }
self.DecreaseIndent    ()
  }
if (stmt.ElseBranch != nil)   {
self.WriteLine("default:")    
self.IncreaseIndent    ()
for i = 0; i <= (int64(len(stmt.ElseBranch.Statements)) - 1); i    ++ {
self.GenerateStatement(stmt.ElseBranch.Statements[i])      
    }
self.DecreaseIndent    ()
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
  _ = i
  _ = j
  _ = branch
}

func (self *TGenerator) GenerateMatchStatement(stmt *TMatchStatement) {
var i int64  
var j int64  
var branch *TMatchBranch  
self.Write("switch ")  
self.GenerateExpression(stmt.Expression)  
self.WriteLine(" {")  
self.IncreaseIndent  ()
for i = 0; i <= (int64(len(stmt.Branches)) - 1); i  ++ {
branch = stmt.Branches[i]    
self.Write("case ")    
self.GenerateExpression(branch.Pattern)    
for j = 0; j <= (int64(len(branch.AdditionalPatterns)) - 1); j    ++ {
self.Write(", ")      
self.GenerateExpression(branch.AdditionalPatterns[j])      
    }
self.WriteLine(":")    
self.IncreaseIndent    ()
if func() bool { _, ok := branch.Body.(*TBlockStatement); return ok }()     {
var blk3 *TBlockStatement      
blk3 = branch.Body.(*TBlockStatement)      
for j = 0; j <= (int64(len(blk3.Statements)) - 1); j      ++ {
self.GenerateStatement(blk3.Statements[j])        
      }
    }
self.DecreaseIndent    ()
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
  _ = i
  _ = j
  _ = branch
}

func (self *TGenerator) GenerateTryStatement(stmt *TTryStatement) {
var i int64  
var k int64  
var onClause *TOnClause  
self.NeedsException = true  
self.WriteLine("func() {")  
self.IncreaseIndent  ()
self.WriteLine("defer func() {")  
self.IncreaseIndent  ()
self.WriteLine("if r := recover(); r != nil {")  
self.IncreaseIndent  ()
if (int64(len(stmt.OnClauses)) > 0)   {
self.WriteLine("switch e := r.(type) {")    
self.IncreaseIndent    ()
for i = 0; i <= (int64(len(stmt.OnClauses)) - 1); i    ++ {
onClause = stmt.OnClauses[i]      
self.Write("case *")      
self.GenerateTypeExpression(onClause.OnType)      
self.WriteLine(":")      
self.IncreaseIndent      ()
self.WriteLine((onClause.Variable + " := e"))      
if (onClause.Body != nil)       {
for k = 0; k <= (int64(len(onClause.Body.Statements)) - 1); k        ++ {
self.GenerateStatement(onClause.Body.Statements[k])          
        }
      }
self.DecreaseIndent      ()
    }
self.DecreaseIndent    ()
self.WriteLine("}")    
  }
if (stmt.ExceptBlock != nil)   {
for i = 0; i <= (int64(len(stmt.ExceptBlock.Statements)) - 1); i    ++ {
self.GenerateStatement(stmt.ExceptBlock.Statements[i])      
    }
  }
self.DecreaseIndent  ()
self.WriteLine("}")  
self.DecreaseIndent  ()
self.WriteLine("}()")  
if (stmt.Body != nil)   {
for i = 0; i <= (int64(len(stmt.Body.Statements)) - 1); i    ++ {
self.GenerateStatement(stmt.Body.Statements[i])      
    }
  }
self.DecreaseIndent  ()
self.WriteLine("}()")  
  _ = i
  _ = k
  _ = onClause
}

func (self *TGenerator) GenerateRaiseStatement(stmt *TRaiseStatement) {
self.Write("panic(")  
if (stmt.Exception != nil)   {
self.GenerateExpression(stmt.Exception)    
  } else {
self.Write("&Exception{Message: \"exception\"}")    
  }
self.WriteLine(")")  
}

func (self *TGenerator) GenerateReturnStatement(stmt *TReturnStatement) {
self.Write("return ")  
if (stmt.ReturnValue != nil)   {
self.GenerateExpression(stmt.ReturnValue)    
  }
self.WriteLine("")  
}

func (self *TGenerator) GenerateExpression(expr interface{}) {
var ident *TIdentifier  
var intLit *TIntegerLiteral  
var floatLit *TFloatLiteral  
var strLit *TStringLiteral  
var boolLit *TBooleanLiteral  
var nilLit *TNilLiteral  
var arrLit *TArrayLiteral  
var prefix *TPrefixExpression  
var infix *TInfixExpression  
var call *TCallExpression  
var member *TMemberExpression  
var index *TIndexExpression  
var slice *TSliceExpression  
var isExpr *TIsExpression  
var asExpr *TTypeCastExpression  
var lam *TLambdaExpression  
var iParam int64  
var param *TParameter  
var elem interface{}  
var i int64  
var op string  
if func() bool { _, ok := expr.(*TIdentifier); return ok }()   {
ident = expr.(*TIdentifier)    
self.Write(self.MapBuiltinFunction(ident.Value))    
  } else {
if func() bool { _, ok := expr.(*TIntegerLiteral); return ok }()     {
intLit = expr.(*TIntegerLiteral)      
self.Write(fmt.Sprintf("%d", intLit.Value))      
    } else {
if func() bool { _, ok := expr.(*TFloatLiteral); return ok }()       {
floatLit = expr.(*TFloatLiteral)        
self.Write("0.0")        
      } else {
if func() bool { _, ok := expr.(*TStringLiteral); return ok }()         {
strLit = expr.(*TStringLiteral)          
self.Write("\"")          
self.WriteEscapedGoString(strLit.Value)          
self.Write("\"")          
        } else {
if func() bool { _, ok := expr.(*TBooleanLiteral); return ok }()           {
boolLit = expr.(*TBooleanLiteral)            
if boolLit.Value             {
self.Write("true")              
            } else {
self.Write("false")              
            }
          } else {
if func() bool { _, ok := expr.(*TNilLiteral); return ok }()             {
self.Write("nil")              
            } else {
if func() bool { _, ok := expr.(*TArrayLiteral); return ok }()               {
arrLit = expr.(*TArrayLiteral)                
if (int64(len(arrLit.Elements)) == 0)                 {
self.Write("nil")                  
                } else {
self.Write("[]interface{}{")                  
for i = 0; i <= (int64(len(arrLit.Elements)) - 1); i                  ++ {
if (i > 0)                     {
self.Write(", ")                      
                    }
self.GenerateExpression(arrLit.Elements[i])                    
                  }
self.Write("}")                  
                }
              } else {
if func() bool { _, ok := expr.(*TGenericType); return ok }()                 {
var genExp *TGenericType                  
genExp = expr.(*TGenericType)                  
self.Write(genExp.Base)                  
if (int64(len(genExp.TypeParams)) > 0)                   {
self.Write("[")                    
for i = 0; i <= (int64(len(genExp.TypeParams)) - 1); i                    ++ {
if (i > 0)                       {
self.Write(", ")                        
                      }
elem = genExp.TypeParams[i]                      
self.GenerateTypeExpression(elem)                      
                    }
self.Write("]")                    
                  }
                } else {
if func() bool { _, ok := expr.(*TPrefixExpression); return ok }()                   {
prefix = expr.(*TPrefixExpression)                    
op = prefix.Operator                    
if (op == "not")                     {
op = "!"                      
                    }
self.Write(("(" + op))                    
self.GenerateExpression(prefix.Right)                    
self.Write(")")                    
                  } else {
if func() bool { _, ok := expr.(*TInfixExpression); return ok }()                     {
infix = expr.(*TInfixExpression)                      
self.Write("(")                      
self.GenerateExpression(infix.Left)                      
self.Write(" ")                      
op = infix.Operator                      
if (op == "and")                       {
op = "&&"                        
                      } else {
if (op == "or")                         {
op = "||"                          
                        } else {
if (op == "xor")                           {
op = "^"                            
                          } else {
if (op == "div")                             {
op = "/"                              
                            } else {
if (op == "mod")                               {
op = "%"                                
                              } else {
if (op == "<>")                                 {
op = "!="                                  
                                } else {
if (op == "=")                                   {
op = "=="                                    
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
self.Write(op)                      
self.Write(" ")                      
self.GenerateExpression(infix.Right)                      
self.Write(")")                      
                    } else {
if func() bool { _, ok := expr.(*TCallExpression); return ok }()                       {
call = expr.(*TCallExpression)                        
self.GenerateCallExpression(call)                        
                      } else {
if func() bool { _, ok := expr.(*TMemberExpression); return ok }()                         {
member = expr.(*TMemberExpression)                          
if (member.Member == "Create")                           {
if func() bool { _, ok := member.Obj.(*TIdentifier); return ok }()                             {
ident = member.Obj.(*TIdentifier)                              
self.Write((("&" + ident.Value) + "{}"))                              
                              return
                            }
self.Write("&")                            
self.GenerateExpression(member.Obj)                            
self.Write("{}")                            
                            return
                          }
self.GenerateExpression(member.Obj)                          
self.Write(("." + member.Member))                          
                        } else {
if func() bool { _, ok := expr.(*TIndexExpression); return ok }()                           {
index = expr.(*TIndexExpression)                            
self.GenerateExpression(index.Left)                            
self.Write("[")                            
self.GenerateExpression(index.Index)                            
self.Write("]")                            
                          } else {
if func() bool { _, ok := expr.(*TSliceExpression); return ok }()                             {
slice = expr.(*TSliceExpression)                              
self.GenerateExpression(slice.Left)                              
self.Write("[")                              
if (slice.Low != nil)                               {
self.GenerateExpression(slice.Low)                                
                              }
self.Write(":")                              
if (slice.High != nil)                               {
self.GenerateExpression(slice.High)                                
                              }
self.Write("]")                              
                            } else {
if func() bool { _, ok := expr.(*TIsExpression); return ok }()                               {
isExpr = expr.(*TIsExpression)                                
self.Write("func() bool { _, ok := ")                                
self.GenerateExpression(isExpr.Expression)                                
self.Write(".(")                                
self.GenerateTypeExpressionForCast(isExpr.TargetType)                                
self.Write("); return ok }()")                                
                              } else {
if func() bool { _, ok := expr.(*TTypeCastExpression); return ok }()                                 {
asExpr = expr.(*TTypeCastExpression)                                  
self.GenerateExpression(asExpr.Expression)                                  
self.Write(".(")                                  
self.GenerateTypeExpressionForCast(asExpr.TargetType)                                  
self.Write(")")                                  
                                } else {
if func() bool { _, ok := expr.(*TLambdaExpression); return ok }()                                   {
lam = expr.(*TLambdaExpression)                                    
self.Write("func(")                                    
for iParam = 0; iParam <= (int64(len(lam.Parameters)) - 1); iParam                                    ++ {
if (iParam > 0)                                       {
self.Write(", ")                                        
                                      }
param = lam.Parameters[iParam]                                      
self.Write(param.Name)                                      
if (param.ParamType != nil)                                       {
self.Write(" ")                                        
self.GenerateTypeExpression(param.ParamType)                                        
                                      }
                                    }
self.Write(") ")                                    
if (lam.Body != nil)                                     {
self.WriteLine("{")                                      
self.IncreaseIndent                                      ()
var lamBlock *TBlockStatement                                      
lamBlock = lam.Body.(*TBlockStatement)                                      
for iParam = 0; iParam <= (int64(len(lamBlock.Statements)) - 1); iParam                                      ++ {
self.GenerateStatement(lamBlock.Statements[iParam])                                        
                                      }
self.DecreaseIndent                                      ()
self.Write("}")                                      
                                    } else {
self.Write("{}")                                      
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  _ = ident
  _ = intLit
  _ = floatLit
  _ = strLit
  _ = boolLit
  _ = nilLit
  _ = arrLit
  _ = prefix
  _ = infix
  _ = call
  _ = member
  _ = index
  _ = slice
  _ = isExpr
  _ = asExpr
  _ = lam
  _ = iParam
  _ = param
  _ = elem
  _ = i
  _ = op
}

func (self *TGenerator) GenerateCallExpression(call *TCallExpression) {
var i int64  
var ident *TIdentifier  
var member *TMemberExpression  
var typeName string  
var fields []string  
if func() bool { _, ok := call.Func.(*TMemberExpression); return ok }()   {
member = call.Func.(*TMemberExpression)    
if (member.Member == "Create")     {
if func() bool { _, ok := member.Obj.(*TIdentifier); return ok }()       {
ident = member.Obj.(*TIdentifier)        
typeName = ident.Value        
self.Write((("&" + typeName) + "{"))        
fields = self.ClassFields[typeName]        
for i = 0; i <= (int64(len(call.Arguments)) - 1); i        ++ {
if (i > 0)           {
self.Write(", ")            
          }
if (i < int64(len(fields)))           {
self.Write((fields[i] + ": "))            
          }
self.GenerateExpression(call.Arguments[i])          
        }
self.Write("}")        
        return
      }
self.Write("&")      
self.GenerateExpression(member.Obj)      
self.Write("{")      
for i = 0; i <= (int64(len(call.Arguments)) - 1); i      ++ {
if (i > 0)         {
self.Write(", ")          
        }
self.GenerateExpression(call.Arguments[i])        
      }
self.Write("}")      
      return
    }
  }
if func() bool { _, ok := call.Func.(*TIdentifier); return ok }()   {
ident = call.Func.(*TIdentifier)    
if (ident.Value == "Length")     {
self.Write("int64(len(")      
self.GenerateExpression(call.Arguments[0])      
self.Write("))")      
      return
    }
if (ident.Value == "IntToStr")     {
self.Write("fmt.Sprintf(\"%d\", ")      
self.GenerateExpression(call.Arguments[0])      
self.Write(")")      
      return
    }
if (ident.Value == "Ord")     {
self.Write("func() int { if len(")      
self.GenerateExpression(call.Arguments[0])      
self.Write(") == 0 { return 0 }; return int(")      
self.GenerateExpression(call.Arguments[0])      
self.Write("[0]) }()")      
      return
    }
if (ident.Value == "StrToInt64")     {
self.Write("func() int64 { v, _ := strconv.ParseInt(")      
self.GenerateExpression(call.Arguments[0])      
self.Write(", 10, 64); return v }()")      
      return
    }
if (ident.Value == "StrToFloat")     {
self.Write("func() float64 { v, _ := strconv.ParseFloat(")      
self.GenerateExpression(call.Arguments[0])      
self.Write(", 64); return v }()")      
      return
    }
if (ident.Value == "ReadFile")     {
self.NeedOS = true      
self.Write("func() string { data, _ := os.ReadFile(")      
self.GenerateExpression(call.Arguments[0])      
self.Write("); return string(data) }()")      
      return
    }
  }
if func() bool { _, ok := call.Func.(*TIdentifier); return ok }()   {
var callee *TIdentifier    
callee = call.Func.(*TIdentifier)    
if ((((((((((((((self.StrContains(self.UserFuncs, (("," + callee.Value) + ",")) == false) && (int64(len(callee.Value)) > 0)) && (callee.Value[0:1] >= "A")) && (callee.Value[0:1] <= "Z")) && (self.MapBuiltinFunction(callee.Value) == callee.Value)) && (callee.Value != "append")) && (callee.Value != "len")) && (callee.Value != "copy")) && (callee.Value != "delete")) && (callee.Value != "insert")) && (callee.Value != "make")) && (callee.Value != "new")) && (callee.Value != "panic")) && (callee.Value != "recover"))     {
self.NeedStdlib = true      
var retT string      
retT = self.StdlibErrorReturnType(callee.Value)      
if (retT != "")       {
self.Write((((("func() " + retT) + " { _v, _ := stdlib.") + callee.Value) + "("))        
      } else {
self.Write((("stdlib." + callee.Value) + "("))        
      }
for i = 0; i <= (int64(len(call.Arguments)) - 1); i      ++ {
if (i > 0)         {
self.Write(", ")          
        }
self.GenerateExpression(call.Arguments[i])        
      }
if (retT != "")       {
self.Write("); return _v }()")        
      } else {
self.Write(")")        
      }
      return
    }
  }
self.GenerateExpression(call.Func)  
self.Write("(")  
for i = 0; i <= (int64(len(call.Arguments)) - 1); i  ++ {
if (i > 0)     {
self.Write(", ")      
    }
self.GenerateExpression(call.Arguments[i])    
  }
self.Write(")")  
  _ = i
  _ = ident
  _ = member
  _ = typeName
  _ = fields
}

func (self *TGenerator) GenerateTypeExpression(expr interface{}) {
var ident *TIdentifier  
var arrType *TArrayType  
var mapType *TMapType  
var genType *TGenericType  
var recType *TRecordType  
var cd *TClassDecl  
if func() bool { _, ok := expr.(*TIdentifier); return ok }()   {
ident = expr.(*TIdentifier)    
if ((ident.Value == "TRequest") || (ident.Value == "BootRequest"))     {
self.Write("*stdlib.BootRequest")      
self.NeedStdlib = true      
    } else {
if ((ident.Value == "TResponse") || (ident.Value == "BootResponse"))       {
self.Write("*stdlib.BootResponse")        
self.NeedStdlib = true        
      } else {
self.Write(self.MapType(ident.Value))        
      }
    }
  } else {
if func() bool { _, ok := expr.(*TArrayType); return ok }()     {
arrType = expr.(*TArrayType)      
self.Write("[]")      
self.GenerateTypeExpression(arrType.ElementType)      
    } else {
if func() bool { _, ok := expr.(*TMapType); return ok }()       {
mapType = expr.(*TMapType)        
self.Write("map[")        
self.GenerateTypeExpression(mapType.KeyType)        
self.Write("]")        
self.GenerateTypeExpression(mapType.ValueType)        
      } else {
if func() bool { _, ok := expr.(*TGenericType); return ok }()         {
genType = expr.(*TGenericType)          
self.Write(genType.Base)          
        } else {
if func() bool { _, ok := expr.(*TRecordType); return ok }()           {
recType = expr.(*TRecordType)            
self.GenerateInlineRecordType(recType)            
          } else {
if func() bool { _, ok := expr.(*TClassDecl); return ok }()             {
cd = expr.(*TClassDecl)              
self.Write(("*" + cd.Name))              
            } else {
self.Write("interface{}")              
            }
          }
        }
      }
    }
  }
  _ = ident
  _ = arrType
  _ = mapType
  _ = genType
  _ = recType
  _ = cd
}

func (self *TGenerator) GenerateTypeExpressionForCast(expr interface{}) {
var ident *TIdentifier  
if func() bool { _, ok := expr.(*TIdentifier); return ok }()   {
ident = expr.(*TIdentifier)    
if ((ident.Value == "TRequest") || (ident.Value == "BootRequest"))     {
self.Write("*stdlib.BootRequest")      
self.NeedStdlib = true      
    } else {
if ((ident.Value == "TResponse") || (ident.Value == "BootResponse"))       {
self.Write("*stdlib.BootResponse")        
self.NeedStdlib = true        
      } else {
if self.ClassIsBase[ident.Value]         {
self.Write(("*" + ident.Value))          
        } else {
self.Write(self.MapType(ident.Value))          
        }
      }
    }
  } else {
self.GenerateTypeExpression(expr)    
  }
  _ = ident
}

func (self *TGenerator) MapType(kylixType string) string {
var result string  
if (kylixType == "Integer")   {
result = "int64"    
  } else {
if (kylixType == "Real")     {
result = "float64"      
    } else {
if (kylixType == "Boolean")       {
result = "bool"        
      } else {
if (kylixType == "String")         {
result = "string"          
        } else {
if (kylixType == "Char")           {
result = "byte"            
          } else {
if (kylixType == "Byte")             {
result = "byte"              
            } else {
if (kylixType == "Word")               {
result = "uint16"                
              } else {
if (kylixType == "Cardinal")                 {
result = "uint32"                  
                } else {
if (kylixType == "LongInt")                   {
result = "int64"                    
                  } else {
if (kylixType == "Double")                     {
result = "float64"                      
                    } else {
if (kylixType == "Extended")                       {
result = "float64"                        
                      } else {
if (((kylixType == "TNode") || (kylixType == "TStatement")) || (kylixType == "TExpression"))                         {
result = "interface{}"                          
                        } else {
if ((kylixType == "TToken") || (kylixType == "TTokenType"))                           {
result = kylixType                            
                          } else {
if self.ClassIsBase[kylixType]                             {
result = "interface{}"                              
                            } else {
if self.StrContains(self.ClassTypes, (("," + kylixType) + ","))                               {
result = ("*" + kylixType)                                
                              } else {
result = kylixType                                
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  return result
}

func (self *TGenerator) MapBuiltinFunction(name string) string {
var result string  
if (name == "Args")   {
result = "os.Args[1:]"    
  } else {
if (name == "WriteLn")     {
result = "fmt.Println"      
    } else {
if (name == "Write")       {
result = "fmt.Print"        
      } else {
if (name == "LowerCase")         {
result = "strings.ToLower"          
        } else {
if (name == "UpperCase")           {
result = "strings.ToUpper"            
          } else {
if (name == "Length")             {
result = "len"              
            } else {
if (name == "Ord")               {
result = "Ord"                
              } else {
if (name == "IntToStr")                 {
result = "IntToStr"                  
                } else {
if (name == "StrToInt64")                   {
result = "StrToInt64"                    
                  } else {
if (name == "StrToFloat")                     {
result = "StrToFloat"                      
                    } else {
if (name == "ReadFile")                       {
result = "ReadFile"                        
                      } else {
if (name == "Exit")                         {
result = "return"                          
                        } else {
if (name == "Break")                           {
result = "break"                            
                          } else {
if (name == "Continue")                             {
result = "continue"                              
                            } else {
result = name                              
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  return result
}

func (self *TGenerator) StdlibErrorReturnType(name string) string {
var result string  
if (name == "ReadFile")   {
result = "string"    
  } else {
if (name == "ParseDate")     {
result = "*stdlib.TDateTime"      
    } else {
if (name == "ParseDateTime")       {
result = "*stdlib.TDateTime"        
      } else {
if (name == "RegexCompile")         {
result = "*stdlib.TRegex"          
        } else {
if (name == "HttpGet")           {
result = "string"            
          } else {
if (name == "HttpPost")             {
result = "string"              
            } else {
if (name == "HttpGetJSON")               {
result = "map[string]interface{}"                
              } else {
if (name == "HttpPut")                 {
result = "string"                  
                } else {
if (name == "HttpDelete")                   {
result = "string"                    
                  } else {
if (name == "HttpPostJSON")                     {
result = "map[string]interface{}"                      
                    } else {
if (name == "HttpDoGet")                       {
result = "*stdlib.THttpResponse"                        
                      } else {
if (name == "HttpDoPost")                         {
result = "*stdlib.THttpResponse"                          
                        } else {
if (name == "WsDial")                           {
result = "*stdlib.TWsConn"                            
                          } else {
if (name == "WsAccept")                             {
result = "*stdlib.TWsConn"                              
                            } else {
if (name == "WsRecv")                               {
result = "string"                                
                              } else {
if (name == "ListDir")                                 {
result = "[]string"                                  
                                } else {
if (name == "ListFiles")                                   {
result = "[]string"                                    
                                  } else {
if (name == "ReadLines")                                     {
result = "[]string"                                      
                                    } else {
if (name == "GetFileSize")                                       {
result = "int64"                                        
                                      } else {
if (name == "FileOpen")                                         {
result = "*stdlib.TTextFile"                                          
                                        } else {
if (name == "JsonDecode")                                           {
result = "interface{}"                                            
                                          } else {
if (name == "JsonDecodeMap")                                             {
result = "map[string]interface{}"                                              
                                            } else {
if (name == "JsonDecodeArray")                                               {
result = "[]interface{}"                                                
                                              } else {
if (name == "JsonReadFile")                                                 {
result = "interface{}"                                                  
                                                } else {
if (name == "TcpDial")                                                   {
result = "*stdlib.TTcpConn"                                                    
                                                  } else {
if (name == "TcpListen")                                                     {
result = "*stdlib.TTcpListener"                                                      
                                                    } else {
if (name == "UdpDial")                                                       {
result = "*stdlib.TUdpConn"                                                        
                                                      } else {
if (name == "DnsLookup")                                                         {
result = "[]string"                                                          
                                                        } else {
if (name == "DnsLookupCNAME")                                                           {
result = "string"                                                            
                                                          } else {
if (name == "AesEncrypt")                                                             {
result = "string"                                                              
                                                            } else {
if (name == "AesDecrypt")                                                               {
result = "string"                                                                
                                                              } else {
if (name == "BCryptHash")                                                                 {
result = "string"                                                                  
                                                                } else {
if (name == "RandomBytes")                                                                   {
result = "string"                                                                    
                                                                  } else {
if (name == "RandomToken")                                                                     {
result = "string"                                                                      
                                                                    } else {
if (name == "Base64Decode")                                                                       {
result = "string"                                                                        
                                                                      } else {
if (name == "HexDecode")                                                                         {
result = "string"                                                                          
                                                                        } else {
if (name == "UrlDecode")                                                                           {
result = "string"                                                                            
                                                                          } else {
if (name == "CsvDecode")                                                                             {
result = "[][]string"                                                                              
                                                                            } else {
if (name == "JsonLinesDecode")                                                                               {
result = "[]map[string]interface{}"                                                                                
                                                                              } else {
if (name == "JwtSign")                                                                                 {
result = "string"                                                                                  
                                                                                } else {
if (name == "JwtVerify")                                                                                   {
result = "map[string]interface{}"                                                                                    
                                                                                  } else {
if (name == "DbOpen")                                                                                     {
result = "*stdlib.Database"                                                                                      
                                                                                    } else {
if (name == "DbOpenSQLite")                                                                                       {
result = "*stdlib.Database"                                                                                        
                                                                                      } else {
if (name == "DbExec")                                                                                         {
result = "int64"                                                                                          
                                                                                        } else {
if (name == "DbQueryRows")                                                                                           {
result = "[]map[string]interface{}"                                                                                            
                                                                                          } else {
if (name == "DbQueryScalar")                                                                                             {
result = "string"                                                                                              
                                                                                            } else {
result = ""                                                                                              
                                                                                            }
                                                                                          }
                                                                                        }
                                                                                      }
                                                                                    }
                                                                                  }
                                                                                }
                                                                              }
                                                                            }
                                                                          }
                                                                        }
                                                                      }
                                                                    }
                                                                  }
                                                                }
                                                              }
                                                            }
                                                          }
                                                        }
                                                      }
                                                    }
                                                  }
                                                }
                                              }
                                            }
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  return result
}

func (self *TGenerator) IsValid() bool {
  return true
}

var Keywords map[string]TTokenType = map[string]TTokenType{}
var KeywordsInitialized bool
const PREC_LOWEST = 1
const PREC_EQUALS = 2
const PREC_LESSGREATER = 3
const PREC_SUM = 4
const PREC_PRODUCT = 5
const PREC_PREFIX = 6
const PREC_CALL = 7
const PREC_INDEX = 8
const PREC_MEMBER = 9
var SourceFile string
var SourceCode string
var Lex *TLexer
var Par *TParser
var Gen *TGenerator
var Prog *TProgram
var GoCode string
var ErrList *TErrorList
var i int64
var AllOk bool
var Files []string
var Programs []*TProgram
func InitKeywords() {
if (KeywordsInitialized == false)   {
Keywords["program"] = tkProgram    
Keywords["unit"] = tkUnit    
Keywords["uses"] = tkUses    
Keywords["var"] = tkVar    
Keywords["const"] = tkConst    
Keywords["type"] = tkType    
Keywords["begin"] = tkBegin    
Keywords["end"] = tkEnd    
Keywords["function"] = tkFunction    
Keywords["procedure"] = tkProcedure    
Keywords["if"] = tkIf    
Keywords["then"] = tkThen    
Keywords["else"] = tkElse    
Keywords["while"] = tkWhile    
Keywords["do"] = tkDo    
Keywords["for"] = tkFor    
Keywords["to"] = tkTo    
Keywords["downto"] = tkDownTo    
Keywords["repeat"] = tkRepeat    
Keywords["until"] = tkUntil    
Keywords["case"] = tkCase    
Keywords["of"] = tkOf    
Keywords["with"] = tkWith    
Keywords["match"] = tkMatch    
Keywords["when"] = tkWhen    
Keywords["try"] = tkTry    
Keywords["except"] = tkExcept    
Keywords["finally"] = tkFinally    
Keywords["raise"] = tkRaise    
Keywords["on"] = tkOn    
Keywords["class"] = tkClass    
Keywords["interface"] = tkInterface    
Keywords["object"] = tkObject    
Keywords["record"] = tkRecord    
Keywords["array"] = tkArray    
Keywords["set"] = tkSet    
Keywords["packed"] = tkPacked    
Keywords["file"] = tkFile    
Keywords["map"] = tkMap    
Keywords["variant"] = tkVariant    
Keywords["inherits"] = tkInherits    
Keywords["implements"] = tkImplements    
Keywords["implementation"] = tkImplementation    
Keywords["public"] = tkPublic    
Keywords["private"] = tkPrivate    
Keywords["protected"] = tkProtected    
Keywords["published"] = tkPublished    
Keywords["property"] = tkProperty    
Keywords["read"] = tkRead    
Keywords["write"] = tkWrite    
Keywords["default"] = tkDefault    
Keywords["stored"] = tkStored    
Keywords["virtual"] = tkVirtual    
Keywords["override"] = tkOverride    
Keywords["abstract"] = tkAbstract    
Keywords["static"] = tkStatic    
Keywords["dynamic"] = tkDynamic    
Keywords["external"] = tkExternal    
Keywords["forward"] = tkForward    
Keywords["inline"] = tkInline    
Keywords["result"] = tkResult    
Keywords["self"] = tkSelf    
Keywords["nil"] = tkNil    
Keywords["true"] = tkTrue    
Keywords["false"] = tkFalse    
Keywords["and"] = tkAnd    
Keywords["or"] = tkOr    
Keywords["not"] = tkNot    
Keywords["xor"] = tkXor    
Keywords["in"] = tkIn    
Keywords["is"] = tkIs    
Keywords["as"] = tkAs    
Keywords["new"] = tkNew    
Keywords["delete"] = tkDelete    
Keywords["break"] = tkBreak    
Keywords["continue"] = tkContinue    
Keywords["exit"] = tkExit    
Keywords["return"] = tkReturn    
Keywords["async"] = tkAsync    
Keywords["await"] = tkAwait    
Keywords["constructor"] = tkConstructor    
Keywords["destructor"] = tkDestructor    
Keywords["inherited"] = tkInherited    
Keywords["import"] = tkImport    
Keywords["export"] = tkExport    
Keywords["module"] = tkModule    
Keywords["mod"] = tkMod    
Keywords["div"] = tkDiv    
KeywordsInitialized = true    
  }
}

func LookupIdent(ident string) TTokenType {
var result TTokenType  
var lower string  
var tok TTokenType  
if (KeywordsInitialized == false)   {
InitKeywords()    
  }
lower = strings.ToLower(ident)  
tok = Keywords[lower]  
if (tok == tkIllegal)   {
result = tkIdent    
  } else {
result = tok    
  }
  _ = lower
  _ = tok
  return result
}

func IsLetter(ch string) bool {
var result bool  
result = false  
if (ch >= "a")   {
if (ch <= "z")     {
result = true      
    }
  }
if (result == false)   {
if (ch >= "A")     {
if (ch <= "Z")       {
result = true        
      }
    }
  }
if (result == false)   {
if (ch == "_")     {
result = true      
    }
  }
  return result
}

func IsDigit(ch string) bool {
var result bool  
result = false  
if (ch >= "0")   {
if (ch <= "9")     {
result = true      
    }
  }
  return result
}

func NewLexer(input string) *TLexer {
var result *TLexer  
var lex *TLexer  
lex = &TLexer{}  
lex.Input = input  
lex.Position = 0  
lex.ReadPosition = 0  
lex.Ch = ""  
lex.Line = 1  
lex.Column = 0  
lex.ReadChar()  
result = lex  
  _ = lex
  return result
}

func (self *TLexer) ReadChar() {
if (self.ReadPosition >= int64(len(self.Input)))   {
self.Ch = ""    
  } else {
self.Ch = self.Input[self.ReadPosition:(self.ReadPosition + 1)]    
  }
self.Position = self.ReadPosition  
self.ReadPosition = (self.ReadPosition + 1)  
self.Column = (self.Column + 1)  
if (func() int { if len(self.Ch) == 0 { return 0 }; return int(self.Ch[0]) }() == 10)   {
self.Line = (self.Line + 1)    
self.Column = 0    
  }
}

func (self *TLexer) PeekChar() string {
var result string  
if (self.ReadPosition >= int64(len(self.Input)))   {
result = ""    
  } else {
result = self.Input[self.ReadPosition:(self.ReadPosition + 1)]    
  }
  return result
}

func (self *TLexer) NewToken(tokenType TTokenType, literal string) TToken {
var result TToken  
var tok TToken  
tok.TokenType = tokenType  
tok.Literal = literal  
tok.Line = self.Line  
tok.Column = self.Column  
result = tok  
  _ = tok
  return result
}

func (self *TLexer) SkipWhitespace() {
var done bool  
done = false  
for (done == false)   {
if (self.Ch == " ")     {
self.ReadChar()      
    } else {
if (func() int { if len(self.Ch) == 0 { return 0 }; return int(self.Ch[0]) }() == 9)       {
self.ReadChar()        
      } else {
if (func() int { if len(self.Ch) == 0 { return 0 }; return int(self.Ch[0]) }() == 10)         {
self.ReadChar()          
        } else {
if (func() int { if len(self.Ch) == 0 { return 0 }; return int(self.Ch[0]) }() == 13)           {
self.ReadChar()            
          } else {
done = true            
          }
        }
      }
    }
  }
  _ = done
}

func (self *TLexer) SkipComments() {
if (self.Ch == "/")   {
if (self.PeekChar() == "/")     {
for ((func() int { if len(self.Ch) == 0 { return 0 }; return int(self.Ch[0]) }() != 10) && (self.Ch != ""))       {
self.ReadChar()        
      }
    }
  }
if (self.Ch == "(")   {
if (self.PeekChar() == "*")     {
self.ReadChar()      
self.ReadChar()      
for (((self.Ch != "*") || (self.PeekChar() != ")")) && (self.Ch != ""))       {
self.ReadChar()        
      }
if (self.Ch == "*")       {
self.ReadChar()        
self.ReadChar()        
      }
    }
  }
}

func (self *TLexer) ReadIdentifier() string {
var result string  
var startPos int64  
startPos = self.Position  
for (IsLetter(self.Ch) || IsDigit(self.Ch))   {
self.ReadChar()    
  }
result = self.Input[startPos:self.Position]  
  _ = startPos
  return result
}

func (self *TLexer) ReadNumber() TToken {
var result TToken  
var startPos int64  
var startLine int64  
var startCol int64  
var tokType TTokenType  
startPos = self.Position  
startLine = self.Line  
startCol = self.Column  
tokType = tkInt  
for IsDigit(self.Ch)   {
self.ReadChar()    
  }
if (self.Ch == ".")   {
if IsDigit(self.PeekChar())     {
tokType = tkFloat      
self.ReadChar()      
for IsDigit(self.Ch)       {
self.ReadChar()        
      }
    }
  }
result.TokenType = tokType  
result.Literal = self.Input[startPos:self.Position]  
result.Line = startLine  
result.Column = startCol  
  _ = startPos
  _ = startLine
  _ = startCol
  _ = tokType
  return result
}

func (self *TLexer) ReadString() string {
var result string  
var startPos int64  
self.ReadChar()  
startPos = self.Position  
for ((self.Ch != "\"") && (self.Ch != ""))   {
self.ReadChar()    
  }
result = self.Input[startPos:self.Position]  
self.ReadChar()  
  _ = startPos
  return result
}

func (self *TLexer) ReadSingleQuotedString() string {
var result string  
var startPos int64  
self.ReadChar()  
startPos = self.Position  
for ((func() int { if len(self.Ch) == 0 { return 0 }; return int(self.Ch[0]) }() != 39) && (self.Ch != ""))   {
self.ReadChar()    
  }
result = self.Input[startPos:self.Position]  
self.ReadChar()  
  _ = startPos
  return result
}

func (self *TLexer) ReadInterpolatedString() string {
var result string  
var startPos int64  
var braceDepth int64  
var done bool  
self.ReadChar()  
startPos = self.Position  
braceDepth = 0  
done = false  
for ((self.Ch != "") && (done == false))   {
if (self.Ch == "{")     {
braceDepth = (braceDepth + 1)      
    } else {
if (self.Ch == "}")       {
braceDepth = (braceDepth - 1)        
      } else {
if ((self.Ch == "\"") && (braceDepth == 0))         {
done = true          
        }
      }
    }
if (done == false)     {
self.ReadChar()      
    }
  }
result = self.Input[startPos:self.Position]  
self.ReadChar()  
  _ = startPos
  _ = braceDepth
  _ = done
  return result
}

func (self *TLexer) NextToken() TToken {
var result TToken  
var tok TToken  
var before string  
var nextCh string  
var skipRead bool  
InitKeywords()  
skipRead = false  
for true   {
self.SkipWhitespace()    
before = self.Ch    
self.SkipComments()    
if (self.Ch == before)     {
      break
    }
  }
self.SkipWhitespace()  
tok.Line = self.Line  
tok.Column = self.Column  
nextCh = self.PeekChar()  
if (self.Ch == "=")   {
if (nextCh == "=")     {
self.ReadChar()      
tok = self.NewToken(tkEQ, "==")      
    } else {
if (nextCh == ">")       {
self.ReadChar()        
tok = self.NewToken(tkFatArrow, "=>")        
      } else {
tok = self.NewToken(tkAssign, "=")        
      }
    }
  } else {
if (self.Ch == "+")     {
tok = self.NewToken(tkPlus, "+")      
    } else {
if (self.Ch == "-")       {
if (nextCh == ">")         {
self.ReadChar()          
tok = self.NewToken(tkArrow, "->")          
        } else {
tok = self.NewToken(tkMinus, "-")          
        }
      } else {
if (self.Ch == "!")         {
if (nextCh == "=")           {
self.ReadChar()            
tok = self.NewToken(tkNotEQ, "!=")            
          } else {
tok = self.NewToken(tkBang, "!")            
          }
        } else {
if (self.Ch == "*")           {
tok = self.NewToken(tkAsterisk, "*")            
          } else {
if (self.Ch == "/")             {
tok = self.NewToken(tkSlash, "/")              
            } else {
if (self.Ch == "<")               {
if (nextCh == "=")                 {
self.ReadChar()                  
tok = self.NewToken(tkLTEQ, "<=")                  
                } else {
if (nextCh == ">")                   {
self.ReadChar()                    
tok = self.NewToken(tkNotEQ, "<>")                    
                  } else {
tok = self.NewToken(tkLT, "<")                    
                  }
                }
              } else {
if (self.Ch == ">")                 {
if (nextCh == "=")                   {
self.ReadChar()                    
tok = self.NewToken(tkGTEQ, ">=")                    
                  } else {
tok = self.NewToken(tkGT, ">")                    
                  }
                } else {
if (self.Ch == ":")                   {
if (nextCh == "=")                     {
self.ReadChar()                      
tok = self.NewToken(tkAssignOp, ":=")                      
                    } else {
tok = self.NewToken(tkColon, ":")                      
                    }
                  } else {
if (self.Ch == ";")                     {
tok = self.NewToken(tkSemicolon, ";")                      
                    } else {
if (self.Ch == ",")                       {
tok = self.NewToken(tkComma, ",")                        
                      } else {
if (self.Ch == ".")                         {
if (nextCh == ".")                           {
self.ReadChar()                            
tok = self.NewToken(tkDotDot, "..")                            
                          } else {
tok = self.NewToken(tkDot, ".")                            
                          }
                        } else {
if (self.Ch == "(")                           {
tok = self.NewToken(tkLParen, "(")                            
                          } else {
if (self.Ch == ")")                             {
tok = self.NewToken(tkRParen, ")")                              
                            } else {
if (self.Ch == "{")                               {
tok = self.NewToken(tkLBrace, "{")                                
                              } else {
if (self.Ch == "}")                                 {
tok = self.NewToken(tkRBrace, "}")                                  
                                } else {
if (self.Ch == "[")                                   {
tok = self.NewToken(tkLBracket, "[")                                    
                                  } else {
if (self.Ch == "]")                                     {
tok = self.NewToken(tkRBracket, "]")                                      
                                    } else {
if (func() int { if len(self.Ch) == 0 { return 0 }; return int(self.Ch[0]) }() == 39)                                       {
tok.TokenType = tkString                                        
tok.Literal = self.ReadSingleQuotedString()                                        
skipRead = true                                        
                                      } else {
if (self.Ch == "\"")                                         {
tok.TokenType = tkString                                          
tok.Literal = self.ReadString()                                          
skipRead = true                                          
                                        } else {
if (self.Ch == "$")                                           {
if (nextCh == "\"")                                             {
self.ReadChar()                                              
tok.TokenType = tkStringInterp                                              
tok.Literal = self.ReadInterpolatedString()                                              
skipRead = true                                              
                                            } else {
tok = self.NewToken(tkIllegal, self.Ch)                                              
                                            }
                                          } else {
if (self.Ch == "")                                             {
tok.Literal = ""                                              
tok.TokenType = tkEOF                                              
                                            } else {
if IsLetter(self.Ch)                                               {
tok.Literal = self.ReadIdentifier()                                                
tok.TokenType = LookupIdent(tok.Literal)                                                
skipRead = true                                                
                                              } else {
if IsDigit(self.Ch)                                                 {
tok = self.ReadNumber()                                                  
skipRead = true                                                  
                                                } else {
tok = self.NewToken(tkIllegal, self.Ch)                                                  
                                                }
                                              }
                                            }
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
if (skipRead == false)   {
self.ReadChar()    
  }
result = tok  
  _ = tok
  _ = before
  _ = nextCh
  _ = skipRead
  return result
}

func NewParser(lex *TLexer) *TParser {
var result *TParser  
var p *TParser  
p = &TParser{}  
p.Lex = lex  
p.NextToken()  
p.NextToken()  
result = p  
  _ = p
  return result
}

func (self *TParser) NextToken() {
self.CurToken = self.PeekToken  
self.PeekToken = self.Lex.NextToken()  
}

func (self *TParser) CurTokenIs(t TTokenType) bool {
var result bool  
result = (self.CurToken.TokenType == t)  
  return result
}

func (self *TParser) PeekTokenIs(t TTokenType) bool {
var result bool  
result = (self.PeekToken.TokenType == t)  
  return result
}

func (self *TParser) ExpectPeek(t TTokenType) bool {
var result bool  
if self.PeekTokenIs(t)   {
self.NextToken()    
result = true    
  } else {
self.PeekError(t)    
result = false    
  }
  return result
}

func (self *TParser) PeekPrecedence() int64 {
var result int64  
result = self.GetPrecedence(self.PeekToken.TokenType)  
  return result
}

func (self *TParser) CurPrecedence() int64 {
var result int64  
result = self.GetPrecedence(self.CurToken.TokenType)  
  return result
}

func (self *TParser) GetPrecedence(tt TTokenType) int64 {
var result int64  
if (tt == tkAssign)   {
result = PREC_EQUALS    
  } else {
if (tt == tkEQ)     {
result = PREC_EQUALS      
    } else {
if (tt == tkNotEQ)       {
result = PREC_EQUALS        
      } else {
if (tt == tkIs)         {
result = PREC_EQUALS          
        } else {
if (tt == tkAs)           {
result = PREC_EQUALS            
          } else {
if (tt == tkLT)             {
result = PREC_LESSGREATER              
            } else {
if (tt == tkLTEQ)               {
result = PREC_LESSGREATER                
              } else {
if (tt == tkGT)                 {
result = PREC_LESSGREATER                  
                } else {
if (tt == tkGTEQ)                   {
result = PREC_LESSGREATER                    
                  } else {
if (tt == tkPlus)                     {
result = PREC_SUM                      
                    } else {
if (tt == tkMinus)                       {
result = PREC_SUM                        
                      } else {
if (tt == tkOr)                         {
result = PREC_SUM                          
                        } else {
if (tt == tkXor)                           {
result = PREC_SUM                            
                          } else {
if (tt == tkAsterisk)                             {
result = PREC_PRODUCT                              
                            } else {
if (tt == tkSlash)                               {
result = PREC_PRODUCT                                
                              } else {
if (tt == tkDiv)                                 {
result = PREC_PRODUCT                                  
                                } else {
if (tt == tkMod)                                   {
result = PREC_PRODUCT                                    
                                  } else {
if (tt == tkAnd)                                     {
result = PREC_PRODUCT                                      
                                    } else {
if (tt == tkLParen)                                       {
result = PREC_CALL                                        
                                      } else {
if (tt == tkLBracket)                                         {
result = PREC_INDEX                                          
                                        } else {
if (tt == tkDot)                                           {
result = PREC_MEMBER                                            
                                          } else {
result = PREC_LOWEST                                            
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  return result
}

func (self *TParser) PeekError(t TTokenType) {
var msg string  
msg = (((((("expected " + fmt.Sprintf("%d", t)) + ", got ") + fmt.Sprintf("%d", self.PeekToken.TokenType)) + " (line ") + fmt.Sprintf("%d", self.PeekToken.Line)) + ")")  
self.Errors = append(self.Errors, msg  )
  _ = msg
}

func (self *TParser) NoPrefixError(t TTokenType) {
var msg string  
msg = (((("no prefix parse function for " + fmt.Sprintf("%d", t)) + " (line ") + fmt.Sprintf("%d", self.CurToken.Line)) + ")")  
self.Errors = append(self.Errors, msg  )
  _ = msg
}

func (self *TParser) SkipAttributes() {
var depth int64  
for self.CurTokenIs(tkLBracket)   {
depth = 1    
self.NextToken()    
for ((depth > 0) && (self.CurTokenIs(tkEOF) == false))     {
if self.CurTokenIs(tkLBracket)       {
depth = (depth + 1)        
      } else {
if self.CurTokenIs(tkRBracket)         {
depth = (depth - 1)          
        }
      }
self.NextToken()      
    }
  }
  _ = depth
}

func (self *TParser) IsIdentOrSoftKeyword() bool {
var result bool  
result = ((((((((((((((((((((((((((self.CurTokenIs(tkIdent) || self.CurTokenIs(tkMatch)) || self.CurTokenIs(tkResult)) || self.CurTokenIs(tkDefault)) || self.CurTokenIs(tkDownTo)) || self.CurTokenIs(tkDynamic)) || self.CurTokenIs(tkTo)) || self.CurTokenIs(tkDo)) || self.CurTokenIs(tkOf)) || self.CurTokenIs(tkIn)) || self.CurTokenIs(tkRead)) || self.CurTokenIs(tkWrite)) || self.CurTokenIs(tkAbstract)) || self.CurTokenIs(tkExternal)) || self.CurTokenIs(tkForward)) || self.CurTokenIs(tkVirtual)) || self.CurTokenIs(tkOverride)) || self.CurTokenIs(tkStatic)) || self.CurTokenIs(tkStored)) || self.CurTokenIs(tkPacked)) || self.CurTokenIs(tkFile)) || self.CurTokenIs(tkNew)) || self.CurTokenIs(tkDelete)) || self.CurTokenIs(tkIs)) || self.CurTokenIs(tkExcept)) || self.CurTokenIs(tkOn)) || self.CurTokenIs(tkWhen))  
  return result
}

func ReadInt64(s string) int64 {
var result int64  
result = func() int64 { v, _ := strconv.ParseInt(s, 10, 64); return v }()  
  return result
}

func ReadFloat64(s string) float64 {
var result float64  
result = func() float64 { v, _ := strconv.ParseFloat(s, 64); return v }()  
  return result
}

func (self *TParser) ParseProgram() *TProgram {
var result *TProgram  
var prog *TProgram  
var iterations int64  
prog = &TProgram{}  
iterations = 0  
if self.CurTokenIs(tkUnit)   {
self.NextToken()    
if self.CurTokenIs(tkIdent)     {
prog.UnitName = self.CurToken.Literal      
prog.IsUnit = true      
prog.Name = self.CurToken.Literal      
prog.NameToken = self.CurToken      
self.NextToken()      
    }
if self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    }
  }
if self.CurTokenIs(tkProgram)   {
self.NextToken()    
if self.CurTokenIs(tkIdent)     {
prog.Name = self.CurToken.Literal      
prog.NameToken = self.CurToken      
self.NextToken()      
    }
if self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    }
  }
if self.CurTokenIs(tkUses)   {
self.NextToken()    
for self.CurTokenIs(tkIdent)     {
self.NextToken()      
if self.CurTokenIs(tkComma)       {
self.NextToken()        
      } else {
        break
      }
    }
if self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    }
  }
for (((self.CurTokenIs(tkEOF) == false) && (self.CurTokenIs(tkEnd) == false)) && (self.CurTokenIs(tkDot) == false))   {
iterations = (iterations + 1)    
if (iterations > 10000)     {
self.Errors = append(self.Errors, "parser exceeded max iterations"      )
      break
    }
self.SkipAttributes()    
if ((self.CurTokenIs(tkEOF) || self.CurTokenIs(tkEnd)) || self.CurTokenIs(tkDot))     {
      break
    }
if self.CurTokenIs(tkVar)     {
self.NextToken()      
for (self.IsIdentOrSoftKeyword() || self.CurTokenIs(tkLParen))       {
prog.Declarations = append(prog.Declarations, self.ParseSingleVarDecl()        )
for self.CurTokenIs(tkSemicolon)         {
self.NextToken()          
        }
      }
    } else {
if self.CurTokenIs(tkConst)       {
self.NextToken()        
for self.CurTokenIs(tkIdent)         {
prog.Declarations = append(prog.Declarations, self.ParseSingleConstDecl()          )
for self.CurTokenIs(tkSemicolon)           {
self.NextToken()            
          }
        }
      } else {
if self.CurTokenIs(tkType)         {
self.NextToken()          
for self.CurTokenIs(tkIdent)           {
prog.Declarations = append(prog.Declarations, self.ParseSingleTypeDecl()            )
for self.CurTokenIs(tkSemicolon)             {
self.NextToken()              
            }
          }
        } else {
if ((self.CurTokenIs(tkFunction) || self.CurTokenIs(tkProcedure)) || self.CurTokenIs(tkAsync))           {
prog.Declarations = append(prog.Declarations, self.ParseFunctionDecl()            )
          } else {
if self.CurTokenIs(tkClass)             {
prog.Declarations = append(prog.Declarations, self.ParseClassDecl()              )
            } else {
if self.CurTokenIs(tkInterface)               {
if prog.IsUnit                 {
self.NextToken()                  
                } else {
prog.Declarations = append(prog.Declarations, self.ParseInterfaceDecl()                  )
                }
              } else {
if self.CurTokenIs(tkImplementation)                 {
self.NextToken()                  
                } else {
if self.CurTokenIs(tkBegin)                   {
var block *TBlockStatement                    
block = self.ParseBlockStatement()                    
if (block != nil)                     {
prog.Statements = append(prog.Statements, block                      )
                    }
                  } else {
var stmt interface{}                    
stmt = self.ParseStatement()                    
if (stmt != nil)                     {
prog.Statements = append(prog.Statements, stmt                      )
                    } else {
self.NextToken()                      
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
for self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    }
  }
if self.CurTokenIs(tkEnd)   {
self.NextToken()    
if self.CurTokenIs(tkDot)     {
self.NextToken()      
    }
  }
result = prog  
  _ = prog
  _ = iterations
  return result
}

func (self *TParser) ParseSingleVarDecl() *TVarDecl {
var result *TVarDecl  
var decl *TVarDecl  
decl = &TVarDecl{}  
decl.Names = nil  
if self.CurTokenIs(tkLParen)   {
self.NextToken()    
for self.IsIdentOrSoftKeyword()     {
decl.Names = append(decl.Names, self.CurToken.Literal      )
self.NextToken()      
if self.CurTokenIs(tkComma)       {
self.NextToken()        
      } else {
        break
      }
    }
if self.CurTokenIs(tkRParen)     {
self.NextToken()      
    }
if self.CurTokenIs(tkAssignOp)     {
decl.Inferred = true      
self.NextToken()      
decl.Value = self.ParseExpression(PREC_LOWEST)      
self.NextToken()      
    }
result = decl    
  }
for self.IsIdentOrSoftKeyword()   {
decl.Names = append(decl.Names, self.CurToken.Literal    )
self.NextToken()    
if self.CurTokenIs(tkComma)     {
self.NextToken()      
    } else {
      break
    }
  }
if self.CurTokenIs(tkAssignOp)   {
decl.Inferred = true    
self.NextToken()    
decl.Value = self.ParseExpression(PREC_LOWEST)    
self.NextToken()    
result = decl    
  }
if self.CurTokenIs(tkColon)   {
self.NextToken()    
decl.VarType = self.ParseTypeExpression()    
  }
if self.CurTokenIs(tkAssign)   {
self.NextToken()    
decl.Value = self.ParseExpression(PREC_LOWEST)    
self.NextToken()    
  }
result = decl  
  _ = decl
  return result
}

func (self *TParser) ParseSingleConstDecl() *TConstDecl {
var result *TConstDecl  
var decl *TConstDecl  
decl = &TConstDecl{}  
if self.CurTokenIs(tkIdent)   {
decl.Token = self.CurToken    
decl.Name = self.CurToken.Literal    
self.NextToken()    
  }
if self.CurTokenIs(tkColon)   {
self.NextToken()    
decl.ConstType = self.ParseTypeExpression()    
  }
if self.CurTokenIs(tkAssign)   {
self.NextToken()    
decl.Value = self.ParseExpression(PREC_LOWEST)    
self.NextToken()    
  }
result = decl  
  _ = decl
  return result
}

func (self *TParser) ParseSingleTypeDecl() *TTypeDecl {
var result *TTypeDecl  
var decl *TTypeDecl  
var typeParams []*TTypeParameter  
decl = &TTypeDecl{}  
typeParams = nil  
if self.CurTokenIs(tkIdent)   {
decl.Token = self.CurToken    
decl.Name = self.CurToken.Literal    
self.NextToken()    
  }
if self.CurTokenIs(tkLT)   {
typeParams = self.ParseTypeParameterList()    
  }
if self.CurTokenIs(tkAssign)   {
self.NextToken()    
if self.CurTokenIs(tkClass)     {
var cd *TClassDecl      
cd = self.ParseClassDecl()      
cd.Name = decl.Name      
cd.TypeParams = typeParams      
decl.DeclType = cd      
    } else {
if self.CurTokenIs(tkInterface)       {
var iface *TInterfaceDecl        
iface = self.ParseInterfaceDecl()        
iface.Name = decl.Name        
decl.DeclType = iface        
      } else {
decl.DeclType = self.ParseTypeExpression()        
      }
    }
  }
result = decl  
  _ = decl
  _ = typeParams
  return result
}

func (self *TParser) ParseFunctionDecl() *TFunctionDecl {
var result *TFunctionDecl  
var decl *TFunctionDecl  
var isProcedure bool  
var hasFuncKeyword bool  
var localDecls []interface{}  
decl = &TFunctionDecl{}  
localDecls = nil  
if self.CurTokenIs(tkAsync)   {
decl.IsAsync = true    
self.NextToken()    
  }
decl.Token = self.CurToken  
isProcedure = self.CurTokenIs(tkProcedure)  
hasFuncKeyword = (((isProcedure == false) && (self.CurTokenIs(tkConstructor) == false)) && (self.CurTokenIs(tkDestructor) == false))  
self.NextToken()  
if (self.CurTokenIs(tkIdent) || self.IsIdentOrSoftKeyword())   {
decl.Name = self.CurToken.Literal    
self.NextToken()    
if self.CurTokenIs(tkDot)     {
self.NextToken()      
if (self.CurTokenIs(tkIdent) || self.IsIdentOrSoftKeyword())       {
decl.Name = ((decl.Name + ".") + self.CurToken.Literal)        
self.NextToken()        
      }
    }
  }
if self.CurTokenIs(tkLT)   {
decl.TypeParams = self.ParseTypeParameterList()    
  }
if self.CurTokenIs(tkLParen)   {
decl.Parameters = self.ParseParameterList()    
  }
if (((isProcedure == false) && hasFuncKeyword) && self.CurTokenIs(tkColon))   {
self.NextToken()    
if self.CurTokenIs(tkLParen)     {
self.NextToken()      
decl.ReturnTypes = nil      
for (self.CurTokenIs(tkRParen) == false)       {
decl.ReturnTypes = append(decl.ReturnTypes, self.ParseTypeExpression()        )
if self.CurTokenIs(tkComma)         {
self.NextToken()          
        }
      }
if self.CurTokenIs(tkRParen)       {
self.NextToken()        
      }
    } else {
decl.ReturnType = self.ParseTypeExpression()      
    }
  }
if self.CurTokenIs(tkSemicolon)   {
self.NextToken()    
  }
if ((((self.CurTokenIs(tkVirtual) || self.CurTokenIs(tkOverride)) || self.CurTokenIs(tkAbstract)) || self.CurTokenIs(tkStatic)) || self.CurTokenIs(tkDynamic))   {
self.NextToken()    
if self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    }
  }
for true   {
if self.CurTokenIs(tkVar)     {
self.NextToken()      
for self.IsIdentOrSoftKeyword()       {
localDecls = append(localDecls, self.ParseSingleVarDecl()        )
for self.CurTokenIs(tkSemicolon)         {
self.NextToken()          
        }
      }
    } else {
if self.CurTokenIs(tkConst)       {
self.NextToken()        
for self.IsIdentOrSoftKeyword()         {
localDecls = append(localDecls, self.ParseSingleConstDecl()          )
for self.CurTokenIs(tkSemicolon)           {
self.NextToken()            
          }
        }
      } else {
        break
      }
    }
  }
if self.CurTokenIs(tkBegin)   {
var block *TBlockStatement    
block = self.ParseBlockStatement()    
decl.Body = block    
  }
decl.LocalDecls = localDecls  
if self.CurTokenIs(tkSemicolon)   {
self.NextToken()    
  }
result = decl  
  _ = decl
  _ = isProcedure
  _ = hasFuncKeyword
  _ = localDecls
  return result
}

func (self *TParser) ParseParameterList() []*TParameter {
var result []*TParameter  
var params []*TParameter  
var param *TParameter  
var names []string  
var i int64  
params = nil  
self.NextToken()  
for ((self.CurTokenIs(tkRParen) == false) && (self.CurTokenIs(tkEOF) == false))   {
names = nil    
for self.CurTokenIs(tkIdent)     {
names = append(names, self.CurToken.Literal      )
self.NextToken()      
if self.CurTokenIs(tkComma)       {
self.NextToken()        
      } else {
        break
      }
    }
var paramType interface{}    
paramType = nil    
if self.CurTokenIs(tkColon)     {
self.NextToken()      
paramType = self.ParseTypeExpression()      
    }
for i = 0; i <= (int64(len(names)) - 1); i    ++ {
param = &TParameter{}      
param.Name = names[i]      
param.ParamType = paramType      
params = append(params, param      )
    }
if self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    } else {
if ((self.CurTokenIs(tkRParen) == false) && (self.CurTokenIs(tkEOF) == false))       {
self.NextToken()        
      }
    }
  }
self.NextToken()  
result = params  
  _ = params
  _ = param
  _ = names
  _ = i
  return result
}

func (self *TParser) ParseTypeParameterList() []*TTypeParameter {
var result []*TTypeParameter  
var params []*TTypeParameter  
var tp *TTypeParameter  
params = nil  
self.NextToken()  
for ((self.CurTokenIs(tkGT) == false) && (self.CurTokenIs(tkEOF) == false))   {
tp = &TTypeParameter{}    
if self.CurTokenIs(tkIdent)     {
tp.Token = self.CurToken      
tp.Name = self.CurToken.Literal      
self.NextToken()      
    }
if self.CurTokenIs(tkColon)     {
self.NextToken()      
tp.Constraint = self.ParseTypeExpression()      
    }
params = append(params, tp    )
if self.CurTokenIs(tkComma)     {
self.NextToken()      
    }
  }
if self.CurTokenIs(tkGT)   {
self.NextToken()    
  }
result = params  
  _ = params
  _ = tp
  return result
}

func (self *TParser) ParseClassDecl() *TClassDecl {
var result *TClassDecl  
var decl *TClassDecl  
decl = &TClassDecl{}  
decl.Visibility = tkPublic  
decl.Fields = nil  
decl.Methods = nil  
decl.Properties = nil  
decl.Interfaces = nil  
self.NextToken()  
if (self.CurTokenIs(tkIdent) && (self.PeekTokenIs(tkColon) == false))   {
decl.Name = self.CurToken.Literal    
self.NextToken()    
  }
if self.CurTokenIs(tkLT)   {
decl.TypeParams = self.ParseTypeParameterList()    
  }
if self.CurTokenIs(tkLParen)   {
self.NextToken()    
if self.CurTokenIs(tkIdent)     {
decl.Parent = self.CurToken.Literal      
self.NextToken()      
    }
if self.CurTokenIs(tkRParen)     {
self.NextToken()      
    }
  } else {
if self.CurTokenIs(tkInherits)     {
self.NextToken()      
if self.CurTokenIs(tkIdent)       {
decl.Parent = self.CurToken.Literal        
self.NextToken()        
      }
    }
  }
if self.CurTokenIs(tkImplements)   {
self.NextToken()    
for self.CurTokenIs(tkIdent)     {
decl.Interfaces = append(decl.Interfaces, self.CurToken.Literal      )
self.NextToken()      
if self.CurTokenIs(tkComma)       {
self.NextToken()        
      } else {
        break
      }
    }
  }
if self.CurTokenIs(tkSemicolon)   {
self.NextToken()    
  }
for ((self.CurTokenIs(tkEnd) == false) && (self.CurTokenIs(tkEOF) == false))   {
self.SkipAttributes()    
if (self.CurTokenIs(tkEnd) || self.CurTokenIs(tkEOF))     {
      break
    }
if ((self.CurTokenIs(tkPublic) || self.CurTokenIs(tkPrivate)) || self.CurTokenIs(tkProtected))     {
decl.Visibility = self.CurToken.TokenType      
self.NextToken()      
    } else {
if self.CurTokenIs(tkVar)       {
self.NextToken()        
for self.IsIdentOrSoftKeyword()         {
decl.Fields = append(decl.Fields, self.ParseSingleVarDecl()          )
for self.CurTokenIs(tkSemicolon)           {
self.NextToken()            
          }
        }
      } else {
if (((self.CurTokenIs(tkFunction) || self.CurTokenIs(tkProcedure)) || self.CurTokenIs(tkConstructor)) || self.CurTokenIs(tkDestructor))         {
decl.Methods = append(decl.Methods, self.ParseFunctionDecl()          )
        } else {
if self.CurTokenIs(tkProperty)           {
decl.Properties = append(decl.Properties, self.ParsePropertyDecl()            )
          } else {
if self.IsIdentOrSoftKeyword()             {
decl.Fields = append(decl.Fields, self.ParseSingleVarDecl()              )
for self.CurTokenIs(tkSemicolon)               {
self.NextToken()                
              }
            } else {
self.NextToken()              
            }
          }
        }
      }
    }
  }
if self.CurTokenIs(tkEnd)   {
self.NextToken()    
  }
if self.CurTokenIs(tkSemicolon)   {
self.NextToken()    
  }
result = decl  
  _ = decl
  return result
}

func (self *TParser) ParseInterfaceDecl() *TInterfaceDecl {
var result *TInterfaceDecl  
var decl *TInterfaceDecl  
decl = &TInterfaceDecl{}  
decl.Parents = nil  
decl.Methods = nil  
self.NextToken()  
if self.CurTokenIs(tkIdent)   {
decl.Name = self.CurToken.Literal    
self.NextToken()    
  }
if self.CurTokenIs(tkLParen)   {
self.NextToken()    
for self.CurTokenIs(tkIdent)     {
decl.Parents = append(decl.Parents, self.CurToken.Literal      )
self.NextToken()      
if self.CurTokenIs(tkComma)       {
self.NextToken()        
      } else {
        break
      }
    }
if self.CurTokenIs(tkRParen)     {
self.NextToken()      
    }
  }
if self.CurTokenIs(tkSemicolon)   {
self.NextToken()    
  }
for ((self.CurTokenIs(tkEnd) == false) && (self.CurTokenIs(tkEOF) == false))   {
if (self.CurTokenIs(tkFunction) || self.CurTokenIs(tkProcedure))     {
decl.Methods = append(decl.Methods, self.ParseFunctionDecl()      )
    } else {
self.NextToken()      
    }
  }
if self.CurTokenIs(tkEnd)   {
self.NextToken()    
  }
if self.CurTokenIs(tkSemicolon)   {
self.NextToken()    
  }
result = decl  
  _ = decl
  return result
}

func (self *TParser) ParsePropertyDecl() *TPropertyDecl {
var result *TPropertyDecl  
var decl *TPropertyDecl  
decl = &TPropertyDecl{}  
self.NextToken()  
if self.CurTokenIs(tkIdent)   {
decl.Name = self.CurToken.Literal    
self.NextToken()    
  }
if self.CurTokenIs(tkColon)   {
self.NextToken()    
decl.PropType = self.ParseTypeExpression()    
  }
for ((self.CurTokenIs(tkRead) || self.CurTokenIs(tkWrite)) || self.CurTokenIs(tkDefault))   {
if self.CurTokenIs(tkRead)     {
self.NextToken()      
if self.CurTokenIs(tkIdent)       {
decl.Getter = self.CurToken.Literal        
self.NextToken()        
      }
    } else {
if self.CurTokenIs(tkWrite)       {
self.NextToken()        
if self.CurTokenIs(tkIdent)         {
decl.Setter = self.CurToken.Literal          
self.NextToken()          
        }
      } else {
if self.CurTokenIs(tkDefault)         {
self.NextToken()          
decl.Default = self.ParseExpression(PREC_LOWEST)          
        }
      }
    }
  }
if self.CurTokenIs(tkSemicolon)   {
self.NextToken()    
  }
result = decl  
  _ = decl
  return result
}

func (self *TParser) ParseStatement() interface{} {
var result interface{}  
if self.CurTokenIs(tkVar)   {
self.NextToken()    
var decl *TVarDecl    
decl = self.ParseSingleVarDecl()    
for self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    }
result = decl    
  } else {
if self.CurTokenIs(tkIf)     {
result = self.ParseIfStatement()      
    } else {
if self.CurTokenIs(tkWhile)       {
result = self.ParseWhileStatement()        
      } else {
if self.CurTokenIs(tkFor)         {
result = self.ParseForStatement()          
        } else {
if self.CurTokenIs(tkRepeat)           {
result = self.ParseRepeatStatement()            
          } else {
if self.CurTokenIs(tkCase)             {
result = self.ParseCaseStatement()              
            } else {
if self.CurTokenIs(tkMatch)               {
if (self.PeekTokenIs(tkAssignOp) || self.PeekTokenIs(tkColon))                 {
result = self.ParseExpressionOrAssignment()                  
                } else {
result = self.ParseMatchStatement()                  
                }
              } else {
if self.CurTokenIs(tkTry)                 {
result = self.ParseTryStatement()                  
                } else {
if self.CurTokenIs(tkRaise)                   {
result = self.ParseRaiseStatement()                    
                  } else {
if self.CurTokenIs(tkInherited)                     {
result = self.ParseInheritedStatement()                      
                    } else {
if self.CurTokenIs(tkBreak)                       {
var brk *TBreakStatement                        
brk = &TBreakStatement{}                        
brk.Token = self.CurToken                        
self.NextToken()                        
result = brk                        
                      } else {
if self.CurTokenIs(tkContinue)                         {
var cont *TContinueStatement                          
cont = &TContinueStatement{}                          
cont.Token = self.CurToken                          
self.NextToken()                          
result = cont                          
                        } else {
if self.CurTokenIs(tkReturn)                           {
result = self.ParseReturnStatement()                            
                          } else {
result = self.ParseExpressionOrAssignment()                            
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  return result
}

func (self *TParser) ParseBlockStatement() *TBlockStatement {
var result *TBlockStatement  
var block *TBlockStatement  
var stmt interface{}  
var iterations int64  
block = &TBlockStatement{}  
block.Token = self.CurToken  
block.Statements = nil  
self.NextToken()  
iterations = 0  
for ((self.CurTokenIs(tkEnd) == false) && (self.CurTokenIs(tkEOF) == false))   {
iterations = (iterations + 1)    
if (iterations > 10000)     {
self.Errors = append(self.Errors, "block parsing exceeded max iterations"      )
      break
    }
stmt = self.ParseStatement()    
if (stmt != nil)     {
block.Statements = append(block.Statements, stmt      )
    } else {
if (self.CurTokenIs(tkSemicolon) == false)       {
self.NextToken()        
      }
    }
for self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    }
  }
if self.CurTokenIs(tkEnd)   {
self.NextToken()    
  }
result = block  
  _ = block
  _ = stmt
  _ = iterations
  return result
}

func (self *TParser) ParseIfStatement() *TIfStatement {
var result *TIfStatement  
var stmt *TIfStatement  
var s interface{}  
stmt = &TIfStatement{}  
stmt.Token = self.CurToken  
self.NextToken()  
stmt.Condition = self.ParseExpression(PREC_LOWEST)  
self.NextToken()  
if self.CurTokenIs(tkThen)   {
self.NextToken()    
  }
if self.CurTokenIs(tkBegin)   {
stmt.Consequence = self.ParseBlockStatement()    
  } else {
s = self.ParseStatement()    
var blk *TBlockStatement    
blk = &TBlockStatement{}    
blk.Statements = nil    
if (s != nil)     {
blk.Statements = append(blk.Statements, s      )
    }
stmt.Consequence = blk    
  }
for self.CurTokenIs(tkSemicolon)   {
self.NextToken()    
  }
if self.CurTokenIs(tkElse)   {
self.NextToken()    
if self.CurTokenIs(tkBegin)     {
stmt.Alternative = self.ParseBlockStatement()      
    } else {
s = self.ParseStatement()      
var blk2 *TBlockStatement      
blk2 = &TBlockStatement{}      
blk2.Statements = nil      
if (s != nil)       {
blk2.Statements = append(blk2.Statements, s        )
      }
stmt.Alternative = blk2      
    }
  }
result = stmt  
  _ = stmt
  _ = s
  return result
}

func (self *TParser) ParseWhileStatement() *TWhileStatement {
var result *TWhileStatement  
var stmt *TWhileStatement  
var s interface{}  
stmt = &TWhileStatement{}  
stmt.Token = self.CurToken  
self.NextToken()  
stmt.Condition = self.ParseExpression(PREC_LOWEST)  
self.NextToken()  
if self.CurTokenIs(tkDo)   {
self.NextToken()    
  }
if self.CurTokenIs(tkBegin)   {
stmt.Body = self.ParseBlockStatement()    
  } else {
s = self.ParseStatement()    
var blk *TBlockStatement    
blk = &TBlockStatement{}    
blk.Statements = nil    
if (s != nil)     {
blk.Statements = append(blk.Statements, s      )
    }
stmt.Body = blk    
  }
result = stmt  
  _ = stmt
  _ = s
  return result
}

func (self *TParser) ParseForStatement() interface{} {
var result interface{}  
var forToken TToken  
var variable string  
forToken = self.CurToken  
self.NextToken()  
variable = ""  
if self.CurTokenIs(tkIdent)   {
variable = self.CurToken.Literal    
self.NextToken()    
  }
if self.CurTokenIs(tkIn)   {
self.NextToken()    
var foreach *TForEachStatement    
foreach = &TForEachStatement{}    
foreach.Token = forToken    
foreach.Variable = variable    
foreach.Iterable = self.ParseExpression(PREC_LOWEST)    
self.NextToken()    
if self.CurTokenIs(tkDo)     {
self.NextToken()      
    }
if self.CurTokenIs(tkBegin)     {
foreach.Body = self.ParseBlockStatement()      
    } else {
var s interface{}      
s = self.ParseStatement()      
var blk *TBlockStatement      
blk = &TBlockStatement{}      
blk.Statements = nil      
if (s != nil)       {
blk.Statements = append(blk.Statements, s        )
      }
foreach.Body = blk      
    }
result = foreach    
    return result
  }
var stmt *TForStatement  
stmt = &TForStatement{}  
stmt.Token = forToken  
stmt.Variable = variable  
if self.CurTokenIs(tkAssignOp)   {
self.NextToken()    
  }
stmt.From = self.ParseExpression(PREC_LOWEST)  
self.NextToken()  
if self.CurTokenIs(tkTo)   {
self.NextToken()    
stmt.DownTo = false    
  } else {
if self.CurTokenIs(tkDownTo)     {
self.NextToken()      
stmt.DownTo = true      
    }
  }
stmt.To = self.ParseExpression(PREC_LOWEST)  
self.NextToken()  
if self.CurTokenIs(tkDo)   {
self.NextToken()    
  }
if self.CurTokenIs(tkBegin)   {
stmt.Body = self.ParseBlockStatement()    
  } else {
var s2 interface{}    
s2 = self.ParseStatement()    
var blk2 *TBlockStatement    
blk2 = &TBlockStatement{}    
blk2.Statements = nil    
if (s2 != nil)     {
blk2.Statements = append(blk2.Statements, s2      )
    }
stmt.Body = blk2    
  }
result = stmt  
  _ = forToken
  _ = variable
  return result
}

func (self *TParser) ParseRepeatStatement() *TRepeatStatement {
var result *TRepeatStatement  
var stmt *TRepeatStatement  
var s interface{}  
stmt = &TRepeatStatement{}  
stmt.Token = self.CurToken  
self.NextToken()  
stmt.Body = &TBlockStatement{}  
stmt.Body.Statements = nil  
for ((self.CurTokenIs(tkUntil) == false) && (self.CurTokenIs(tkEOF) == false))   {
s = self.ParseStatement()    
if (s != nil)     {
stmt.Body.Statements = append(stmt.Body.Statements, s      )
    }
for self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    }
  }
if self.CurTokenIs(tkUntil)   {
self.NextToken()    
stmt.Condition = self.ParseExpression(PREC_LOWEST)    
self.NextToken()    
  }
result = stmt  
  _ = stmt
  _ = s
  return result
}

func (self *TParser) ParseCaseStatement() *TCaseStatement {
var result *TCaseStatement  
var stmt *TCaseStatement  
var branch *TCaseBranch  
var iterations int64  
stmt = &TCaseStatement{}  
stmt.Token = self.CurToken  
stmt.Branches = nil  
self.NextToken()  
stmt.Expression = self.ParseExpression(PREC_LOWEST)  
self.NextToken()  
if self.CurTokenIs(tkOf)   {
self.NextToken()    
  }
iterations = 0  
for ((self.CurTokenIs(tkEnd) == false) && (self.CurTokenIs(tkEOF) == false))   {
iterations = (iterations + 1)    
if (iterations > 10000)     {
      break
    }
branch = &TCaseBranch{}    
branch.Values = nil    
for true     {
var val interface{}      
val = self.ParseExpression(PREC_LOWEST)      
if (val != nil)       {
branch.Values = append(branch.Values, val        )
      } else {
        break
      }
self.NextToken()      
if self.CurTokenIs(tkComma)       {
self.NextToken()        
      } else {
        break
      }
    }
if self.CurTokenIs(tkColon)     {
self.NextToken()      
    }
if self.CurTokenIs(tkBegin)     {
branch.Body = self.ParseBlockStatement()      
    } else {
var s interface{}      
s = self.ParseStatement()      
var blk *TBlockStatement      
blk = &TBlockStatement{}      
blk.Statements = nil      
if (s != nil)       {
blk.Statements = append(blk.Statements, s        )
      }
branch.Body = blk      
    }
stmt.Branches = append(stmt.Branches, branch    )
if self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    }
  }
if self.CurTokenIs(tkEnd)   {
self.NextToken()    
  }
result = stmt  
  _ = stmt
  _ = branch
  _ = iterations
  return result
}

func (self *TParser) ParseMatchStatement() *TMatchStatement {
var result *TMatchStatement  
var stmt *TMatchStatement  
var branch *TMatchBranch  
var isBrace bool  
var endType TTokenType  
stmt = &TMatchStatement{}  
stmt.Token = self.CurToken  
stmt.Branches = nil  
self.NextToken()  
stmt.Expression = self.ParseExpression(PREC_LOWEST)  
self.NextToken()  
if (self.CurTokenIs(tkLBrace) || self.CurTokenIs(tkBegin))   {
isBrace = self.CurTokenIs(tkLBrace)    
self.NextToken()    
if isBrace     {
endType = tkRBrace      
    } else {
endType = tkEnd      
    }
for ((self.CurTokenIs(endType) == false) && (self.CurTokenIs(tkEOF) == false))     {
branch = &TMatchBranch{}      
branch.AdditionalPatterns = nil      
if self.CurTokenIs(tkWhen)       {
self.NextToken()        
branch.When = self.ParseExpression(PREC_LOWEST)        
self.NextToken()        
      } else {
branch.Pattern = self.ParseExpression(PREC_LOWEST)        
self.NextToken()        
for (self.CurTokenIs(tkComma) && (self.PeekTokenIs(tkFatArrow) == false))         {
self.NextToken()          
branch.AdditionalPatterns = append(branch.AdditionalPatterns, self.ParseExpression(PREC_LOWEST)          )
self.NextToken()          
        }
if self.CurTokenIs(tkWhen)         {
self.NextToken()          
branch.When = self.ParseExpression(PREC_LOWEST)          
self.NextToken()          
        }
      }
if (self.CurTokenIs(tkFatArrow) || self.CurTokenIs(tkColon))       {
self.NextToken()        
      }
if self.CurTokenIs(tkBegin)       {
branch.Body = self.ParseBlockStatement()        
      } else {
var s interface{}        
s = self.ParseStatement()        
var blk *TBlockStatement        
blk = &TBlockStatement{}        
blk.Statements = nil        
if (s != nil)         {
blk.Statements = append(blk.Statements, s          )
        }
branch.Body = blk        
      }
stmt.Branches = append(stmt.Branches, branch      )
if (self.CurTokenIs(tkComma) || self.CurTokenIs(tkSemicolon))       {
self.NextToken()        
      }
    }
self.NextToken()    
  }
result = stmt  
  _ = stmt
  _ = branch
  _ = isBrace
  _ = endType
  return result
}

func (self *TParser) ParseTryStatement() *TTryStatement {
var result *TTryStatement  
var stmt *TTryStatement  
var s interface{}  
stmt = &TTryStatement{}  
stmt.Token = self.CurToken  
stmt.OnClauses = nil  
self.NextToken()  
if self.CurTokenIs(tkBegin)   {
stmt.Body = self.ParseBlockStatement()    
  } else {
stmt.Body = &TBlockStatement{}    
stmt.Body.Statements = nil    
for ((((self.CurTokenIs(tkExcept) == false) && (self.CurTokenIs(tkFinally) == false)) && (self.CurTokenIs(tkEnd) == false)) && (self.CurTokenIs(tkEOF) == false))     {
s = self.ParseStatement()      
if (s != nil)       {
stmt.Body.Statements = append(stmt.Body.Statements, s        )
      } else {
if (self.CurTokenIs(tkSemicolon) == false)         {
self.NextToken()          
        }
      }
for self.CurTokenIs(tkSemicolon)       {
self.NextToken()        
      }
    }
  }
if self.CurTokenIs(tkExcept)   {
self.NextToken()    
if self.CurTokenIs(tkBegin)     {
stmt.ExceptBlock = self.ParseBlockStatement()      
    } else {
for (((self.CurTokenIs(tkEnd) == false) && (self.CurTokenIs(tkFinally) == false)) && (self.CurTokenIs(tkEOF) == false))       {
if self.CurTokenIs(tkOn)         {
stmt.OnClauses = append(stmt.OnClauses, self.ParseOnClause()          )
        } else {
if (stmt.ExceptBlock == nil)           {
stmt.ExceptBlock = &TBlockStatement{}            
stmt.ExceptBlock.Statements = nil            
          }
s = self.ParseStatement()          
if (s != nil)           {
stmt.ExceptBlock.Statements = append(stmt.ExceptBlock.Statements, s            )
          } else {
if (self.CurTokenIs(tkSemicolon) == false)             {
self.NextToken()              
            }
          }
        }
for self.CurTokenIs(tkSemicolon)         {
self.NextToken()          
        }
      }
    }
  }
if self.CurTokenIs(tkFinally)   {
self.NextToken()    
if self.CurTokenIs(tkBegin)     {
stmt.FinallyBlock = self.ParseBlockStatement()      
    } else {
stmt.FinallyBlock = &TBlockStatement{}      
stmt.FinallyBlock.Statements = nil      
for ((self.CurTokenIs(tkEnd) == false) && (self.CurTokenIs(tkEOF) == false))       {
s = self.ParseStatement()        
if (s != nil)         {
stmt.FinallyBlock.Statements = append(stmt.FinallyBlock.Statements, s          )
        } else {
if (self.CurTokenIs(tkSemicolon) == false)           {
self.NextToken()            
          }
        }
for self.CurTokenIs(tkSemicolon)         {
self.NextToken()          
        }
      }
    }
  }
if self.CurTokenIs(tkEnd)   {
self.NextToken()    
  }
result = stmt  
  _ = stmt
  _ = s
  return result
}

func (self *TParser) ParseOnClause() *TOnClause {
var result *TOnClause  
var clause *TOnClause  
var s interface{}  
clause = &TOnClause{}  
clause.Token = self.CurToken  
self.NextToken()  
if self.CurTokenIs(tkIdent)   {
clause.Variable = self.CurToken.Literal    
self.NextToken()    
  }
if self.CurTokenIs(tkColon)   {
self.NextToken()    
clause.OnType = self.ParseTypeExpression()    
  }
if self.CurTokenIs(tkDo)   {
self.NextToken()    
  }
if self.CurTokenIs(tkBegin)   {
clause.Body = self.ParseBlockStatement()    
  } else {
s = self.ParseStatement()    
clause.Body = &TBlockStatement{}    
clause.Body.Statements = nil    
if (s != nil)     {
clause.Body.Statements = append(clause.Body.Statements, s      )
    }
  }
result = clause  
  _ = clause
  _ = s
  return result
}

func (self *TParser) ParseRaiseStatement() *TRaiseStatement {
var result *TRaiseStatement  
var stmt *TRaiseStatement  
stmt = &TRaiseStatement{}  
stmt.Token = self.CurToken  
self.NextToken()  
if (self.CurTokenIs(tkSemicolon) == false)   {
stmt.Exception = self.ParseExpression(PREC_LOWEST)    
self.NextToken()    
  }
result = stmt  
  _ = stmt
  return result
}

func (self *TParser) ParseReturnStatement() *TReturnStatement {
var result *TReturnStatement  
var stmt *TReturnStatement  
stmt = &TReturnStatement{}  
stmt.Token = self.CurToken  
self.NextToken()  
if ((self.CurTokenIs(tkSemicolon) == false) && (self.CurTokenIs(tkEnd) == false))   {
stmt.ReturnValue = self.ParseExpression(PREC_LOWEST)    
  }
result = stmt  
  _ = stmt
  return result
}

func (self *TParser) ParseInheritedStatement() *TInheritedStatement {
var result *TInheritedStatement  
var stmt *TInheritedStatement  
stmt = &TInheritedStatement{}  
stmt.Token = self.CurToken  
self.NextToken()  
if (self.CurTokenIs(tkSemicolon) == false)   {
stmt.Expr = self.ParseExpression(PREC_LOWEST)    
  }
for self.PeekTokenIs(tkSemicolon)   {
self.NextToken()    
  }
result = stmt  
  _ = stmt
  return result
}

func (self *TParser) ParseExpressionOrAssignment() interface{} {
var result interface{}  
var firstToken TToken  
var expr interface{}  
firstToken = self.CurToken  
expr = self.ParseExpression(PREC_LOWEST)  
if (expr == nil)   {
result = nil    
    return result
  }
if (self.PeekTokenIs(tkAssign) || self.PeekTokenIs(tkAssignOp))   {
self.NextToken()    
var assignTok TToken    
assignTok = self.CurToken    
self.NextToken()    
var value interface{}    
value = self.ParseExpression(PREC_LOWEST)    
self.NextToken()    
var assign *TAssignmentStatement    
assign = &TAssignmentStatement{}    
assign.Token = assignTok    
assign.Name = expr    
assign.Value = value    
result = assign    
    return result
  }
self.NextToken()  
var exprStmt *TExpressionStatement  
exprStmt = &TExpressionStatement{}  
exprStmt.Token = firstToken  
exprStmt.Expression = expr  
result = exprStmt  
  _ = firstToken
  _ = expr
  return result
}

func (self *TParser) ParseExpression(prec int64) interface{} {
var result interface{}  
var leftExp interface{}  
if self.CurTokenIs(tkIdent)   {
leftExp = self.ParseIdentifier()    
  } else {
if self.CurTokenIs(tkResult)     {
leftExp = self.ParseResultIdent()      
    } else {
if self.CurTokenIs(tkInt)       {
leftExp = self.ParseIntegerLiteral()        
      } else {
if self.CurTokenIs(tkFloat)         {
leftExp = self.ParseFloatLiteral()          
        } else {
if self.CurTokenIs(tkString)           {
leftExp = self.ParseStringLiteral()            
          } else {
if self.CurTokenIs(tkStringInterp)             {
leftExp = self.ParseStringInterpolation()              
            } else {
if self.CurTokenIs(tkChar)               {
leftExp = self.ParseStringLiteral()                
              } else {
if (self.CurTokenIs(tkTrue) || self.CurTokenIs(tkFalse))                 {
leftExp = self.ParseBooleanLiteral()                  
                } else {
if self.CurTokenIs(tkNil)                   {
leftExp = self.ParseNilLiteral()                    
                  } else {
if ((self.CurTokenIs(tkBang) || self.CurTokenIs(tkMinus)) || self.CurTokenIs(tkNot))                     {
leftExp = self.ParsePrefixExpression()                      
                    } else {
if self.CurTokenIs(tkLParen)                       {
leftExp = self.ParseGroupedExpression()                        
                      } else {
if self.CurTokenIs(tkLBracket)                         {
leftExp = self.ParseArrayLiteral()                          
                        } else {
if self.CurTokenIs(tkAwait)                           {
leftExp = self.ParseAwaitExpression()                            
                          } else {
if self.CurTokenIs(tkSelf)                             {
leftExp = self.ParseSelfExpression()                              
                            } else {
if (self.CurTokenIs(tkProcedure) || self.CurTokenIs(tkFunction))                               {
leftExp = self.ParseAnonymousFunction()                                
                              } else {
if self.CurTokenIs(tkMatch)                                 {
leftExp = self.ParseIdentifier()                                  
                                } else {
if self.CurTokenIs(tkExit)                                   {
leftExp = self.ParseIdentifier()                                    
                                  } else {
if self.CurTokenIs(tkReturn)                                     {
leftExp = self.ParseIdentifier()                                      
                                    } else {
if self.CurTokenIs(tkBreak)                                       {
leftExp = self.ParseIdentifier()                                        
                                      } else {
if self.CurTokenIs(tkContinue)                                         {
leftExp = self.ParseIdentifier()                                          
                                        } else {
if self.CurTokenIs(tkDelete)                                           {
leftExp = self.ParseIdentifier()                                            
                                          } else {
if self.CurTokenIs(tkNew)                                             {
leftExp = self.ParseIdentifier()                                              
                                            } else {
if self.CurTokenIs(tkDefault)                                               {
leftExp = self.ParseIdentifier()                                                
                                              } else {
if self.CurTokenIs(tkInherited)                                                 {
leftExp = self.ParseIdentifier()                                                  
                                                } else {
if self.CurTokenIs(tkImport)                                                   {
leftExp = self.ParseIdentifier()                                                    
                                                  } else {
if self.CurTokenIs(tkExport)                                                     {
leftExp = self.ParseIdentifier()                                                      
                                                    } else {
if self.CurTokenIs(tkModule)                                                       {
leftExp = self.ParseIdentifier()                                                        
                                                      } else {
if self.CurTokenIs(tkAbstract)                                                         {
leftExp = self.ParseIdentifier()                                                          
                                                        } else {
if self.CurTokenIs(tkStatic)                                                           {
leftExp = self.ParseIdentifier()                                                            
                                                          } else {
if self.CurTokenIs(tkVirtual)                                                             {
leftExp = self.ParseIdentifier()                                                              
                                                            } else {
if self.CurTokenIs(tkOverride)                                                               {
leftExp = self.ParseIdentifier()                                                                
                                                              } else {
self.NoPrefixError(self.CurToken.TokenType)                                                                
result = nil                                                                
                                                              }
                                                            }
                                                          }
                                                        }
                                                      }
                                                    }
                                                  }
                                                }
                                              }
                                            }
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
for ((self.PeekTokenIs(tkSemicolon) == false) && (prec < self.PeekPrecedence()))   {
if ((((((((((((((self.PeekTokenIs(tkPlus) || self.PeekTokenIs(tkMinus)) || self.PeekTokenIs(tkAsterisk)) || self.PeekTokenIs(tkSlash)) || self.PeekTokenIs(tkDiv)) || self.PeekTokenIs(tkMod)) || self.PeekTokenIs(tkAssign)) || self.PeekTokenIs(tkEQ)) || self.PeekTokenIs(tkNotEQ)) || self.PeekTokenIs(tkLTEQ)) || self.PeekTokenIs(tkGT)) || self.PeekTokenIs(tkGTEQ)) || self.PeekTokenIs(tkAnd)) || self.PeekTokenIs(tkOr)) || self.PeekTokenIs(tkXor))     {
self.NextToken()      
leftExp = self.ParseInfixExpression(leftExp)      
    } else {
if self.PeekTokenIs(tkLT)       {
self.NextToken()        
leftExp = self.ParseLTExpression(leftExp)        
      } else {
if self.PeekTokenIs(tkLParen)         {
self.NextToken()          
leftExp = self.ParseCallExpression(leftExp)          
        } else {
if self.PeekTokenIs(tkLBracket)           {
self.NextToken()            
leftExp = self.ParseIndexExpression(leftExp)            
          } else {
if self.PeekTokenIs(tkDot)             {
self.NextToken()              
leftExp = self.ParseMemberExpression(leftExp)              
            } else {
if self.PeekTokenIs(tkIs)               {
self.NextToken()                
leftExp = self.ParseIsExpression(leftExp)                
              } else {
if self.PeekTokenIs(tkAs)                 {
self.NextToken()                  
leftExp = self.ParseAsExpression(leftExp)                  
                } else {
                  break
                }
              }
            }
          }
        }
      }
    }
  }
result = leftExp  
  _ = leftExp
  return result
}

func (self *TParser) ParseIdentifier() interface{} {
var result interface{}  
var ident *TIdentifier  
ident = &TIdentifier{}  
ident.Token = self.CurToken  
ident.Value = self.CurToken.Literal  
result = ident  
  _ = ident
  return result
}

func (self *TParser) ParseResultIdent() interface{} {
var result interface{}  
var ident *TIdentifier  
ident = &TIdentifier{}  
ident.Token = self.CurToken  
ident.Value = "result"  
result = ident  
  _ = ident
  return result
}

func (self *TParser) ParseIntegerLiteral() interface{} {
var result interface{}  
var lit *TIntegerLiteral  
lit = &TIntegerLiteral{}  
lit.Token = self.CurToken  
lit.Value = ReadInt64(self.CurToken.Literal)  
result = lit  
  _ = lit
  return result
}

func (self *TParser) ParseFloatLiteral() interface{} {
var result interface{}  
var lit *TFloatLiteral  
lit = &TFloatLiteral{}  
lit.Token = self.CurToken  
lit.Value = ReadFloat64(self.CurToken.Literal)  
result = lit  
  _ = lit
  return result
}

func (self *TParser) ParseStringLiteral() interface{} {
var result interface{}  
var lit *TStringLiteral  
lit = &TStringLiteral{}  
lit.Token = self.CurToken  
lit.Value = self.CurToken.Literal  
result = lit  
  _ = lit
  return result
}

func (self *TParser) ParseBooleanLiteral() interface{} {
var result interface{}  
var lit *TBooleanLiteral  
lit = &TBooleanLiteral{}  
lit.Token = self.CurToken  
lit.Value = self.CurTokenIs(tkTrue)  
result = lit  
  _ = lit
  return result
}

func (self *TParser) ParseNilLiteral() interface{} {
var result interface{}  
var lit *TNilLiteral  
lit = &TNilLiteral{}  
lit.Token = self.CurToken  
result = lit  
  _ = lit
  return result
}

func (self *TParser) ParseSelfExpression() interface{} {
var result interface{}  
var ident *TIdentifier  
ident = &TIdentifier{}  
ident.Token = self.CurToken  
ident.Value = "self"  
result = ident  
  _ = ident
  return result
}

func (self *TParser) ParsePrefixExpression() interface{} {
var result interface{}  
var expr *TPrefixExpression  
expr = &TPrefixExpression{}  
expr.Token = self.CurToken  
expr.Operator = self.CurToken.Literal  
self.NextToken()  
expr.Right = self.ParseExpression(PREC_PREFIX)  
result = expr  
  _ = expr
  return result
}

func (self *TParser) ParseInfixExpression(left interface{}) interface{} {
var result interface{}  
var expr *TInfixExpression  
expr = &TInfixExpression{}  
expr.Token = self.CurToken  
expr.Operator = self.CurToken.Literal  
expr.Left = left  
var prec int64  
prec = self.CurPrecedence()  
self.NextToken()  
expr.Right = self.ParseExpression(prec)  
result = expr  
  _ = expr
  return result
}

func (self *TParser) ParseLTExpression(left interface{}) interface{} {
var result interface{}  
var leftIdent *TIdentifier  
var gen *TGenericType  
var tp interface{}  
if func() bool { _, ok := left.(*TIdentifier); return ok }()   {
leftIdent = left.(*TIdentifier)    
if (((int64(len(leftIdent.Value)) > 0) && (leftIdent.Value[0:1] >= "A")) && (leftIdent.Value[0:1] <= "Z"))     {
if (((self.PeekTokenIs(tkIdent) || self.PeekTokenIs(tkArray)) || self.PeekTokenIs(tkMap)) || self.PeekTokenIs(tkRecord))       {
if self.PeekTokenIs(tkIdent)         {
if (((int64(len(self.PeekToken.Literal)) > 0) && (self.PeekToken.Literal[0:1] >= "A")) && (self.PeekToken.Literal[0:1] <= "Z"))           {
gen = &TGenericType{}            
gen.Base = leftIdent.Value            
self.NextToken()            
for ((self.CurTokenIs(tkGT) == false) && (self.CurTokenIs(tkEOF) == false))             {
tp = self.ParseTypeExpression()              
if (tp != nil)               {
gen.TypeParams = append(gen.TypeParams, tp                )
              }
if self.CurTokenIs(tkComma)               {
self.NextToken()                
              } else {
                break
              }
            }
result = gen            
            return result
          }
        }
      }
    }
  }
result = self.ParseInfixExpression(left)  
  _ = leftIdent
  _ = gen
  _ = tp
  return result
}

func (self *TParser) ParseGroupedExpression() interface{} {
var result interface{}  
var openParen TToken  
var exp interface{}  
openParen = self.CurToken  
self.NextToken()  
if (self.CurTokenIs(tkRParen) && self.PeekTokenIs(tkArrow))   {
self.NextToken()    
self.NextToken()    
var lam *TLambdaExpression    
lam = &TLambdaExpression{}    
lam.Token = openParen    
lam.Parameters = nil    
lam.Body = self.ParseExpression(PREC_LOWEST)    
result = lam    
    return result
  }
if (self.CurTokenIs(tkIdent) && self.PeekTokenIs(tkColon))   {
result = self.TryParseLambdaParams(openParen)    
    return result
  }
exp = self.ParseExpression(PREC_LOWEST)  
if self.PeekTokenIs(tkComma)   {
self.NextToken()    
var tuple *TTupleLiteral    
tuple = &TTupleLiteral{}    
tuple.Token = openParen    
tuple.Elements = nil    
tuple.Elements = append(tuple.Elements, exp    )
for self.CurTokenIs(tkComma)     {
self.NextToken()      
tuple.Elements = append(tuple.Elements, self.ParseExpression(PREC_LOWEST)      )
if self.PeekTokenIs(tkComma)       {
self.NextToken()        
      }
    }
if (self.ExpectPeek(tkRParen) == false)     {
result = nil      
      return result
    }
result = tuple    
    return result
  }
if (self.ExpectPeek(tkRParen) == false)   {
result = nil    
    return result
  }
result = exp  
  _ = openParen
  _ = exp
  return result
}

func (self *TParser) TryParseLambdaParams(openParen TToken) interface{} {
var result interface{}  
var params []*TParameter  
var param *TParameter  
params = nil  
for ((self.CurTokenIs(tkRParen) == false) && (self.CurTokenIs(tkEOF) == false))   {
if (self.CurTokenIs(tkIdent) == false)     {
result = nil      
    }
param = &TParameter{}    
param.Token = self.CurToken    
param.Name = self.CurToken.Literal    
self.NextToken()    
if self.CurTokenIs(tkColon)     {
self.NextToken()      
param.ParamType = self.ParseTypeExpression()      
    }
params = append(params, param    )
if self.CurTokenIs(tkSemicolon)     {
self.NextToken()      
    } else {
if self.CurTokenIs(tkComma)       {
self.NextToken()        
      }
    }
  }
if (self.CurTokenIs(tkRParen) == false)   {
result = nil    
  }
self.NextToken()  
if (self.CurTokenIs(tkArrow) == false)   {
result = nil    
  }
self.NextToken()  
var lam *TLambdaExpression  
lam = &TLambdaExpression{}  
lam.Token = openParen  
lam.Parameters = params  
if self.CurTokenIs(tkBegin)   {
lam.Body = self.ParseBlockStatement()    
  } else {
lam.Body = self.ParseExpression(PREC_LOWEST)    
  }
result = lam  
  _ = params
  _ = param
  return result
}

func (self *TParser) ParseCallExpression(fn interface{}) interface{} {
var result interface{}  
var expr *TCallExpression  
expr = &TCallExpression{}  
expr.Token = self.CurToken  
expr.Func = fn  
expr.Arguments = self.ParseExpressionList(tkRParen)  
result = expr  
  _ = expr
  return result
}

func (self *TParser) ParseIndexExpression(left interface{}) interface{} {
var result interface{}  
var tok TToken  
var first interface{}  
tok = self.CurToken  
self.NextToken()  
first = self.ParseExpression(PREC_LOWEST)  
if self.PeekTokenIs(tkColon)   {
self.NextToken()    
self.NextToken()    
var high interface{}    
high = self.ParseExpression(PREC_LOWEST)    
if (self.ExpectPeek(tkRBracket) == false)     {
result = nil      
    }
var slice *TSliceExpression    
slice = &TSliceExpression{}    
slice.Token = tok    
slice.Left = left    
slice.Low = first    
slice.High = high    
result = slice    
    return result
  }
if (self.ExpectPeek(tkRBracket) == false)   {
result = nil    
    return result
  }
var idx *TIndexExpression  
idx = &TIndexExpression{}  
idx.Token = tok  
idx.Left = left  
idx.Index = first  
result = idx  
  _ = tok
  _ = first
  return result
}

func (self *TParser) ParseMemberExpression(left interface{}) interface{} {
var result interface{}  
var dotToken TToken  
dotToken = self.CurToken  
self.NextToken()  
if (self.CurTokenIs(tkIdent) || self.IsIdentOrSoftKeyword())   {
var mem *TMemberExpression    
mem = &TMemberExpression{}    
mem.Token = dotToken    
mem.Obj = left    
mem.Member = self.CurToken.Literal    
result = mem    
    return result
  }
result = left  
  _ = dotToken
  return result
}

func (self *TParser) ParseArrayLiteral() interface{} {
var result interface{}  
var arr *TArrayLiteral  
arr = &TArrayLiteral{}  
arr.Token = self.CurToken  
arr.Elements = nil  
self.NextToken()  
for ((self.CurTokenIs(tkRBracket) == false) && (self.CurTokenIs(tkEOF) == false))   {
var elem interface{}    
elem = self.ParseExpression(PREC_LOWEST)    
if (elem != nil)     {
arr.Elements = append(arr.Elements, elem      )
    }
self.NextToken()    
if self.CurTokenIs(tkComma)     {
self.NextToken()      
    }
  }
result = arr  
  _ = arr
  return result
}

func (self *TParser) ParseAwaitExpression() interface{} {
var result interface{}  
var awaitToken TToken  
var expr *TAwaitExpression  
awaitToken = self.CurToken  
self.NextToken()  
expr = &TAwaitExpression{}  
expr.Token = awaitToken  
expr.Expression = self.ParseExpression(PREC_PREFIX)  
result = expr  
  _ = awaitToken
  _ = expr
  return result
}

func (self *TParser) ParseIsExpression(left interface{}) interface{} {
var result interface{}  
var isToken TToken  
var expr *TIsExpression  
isToken = self.CurToken  
self.NextToken()  
expr = &TIsExpression{}  
expr.Token = isToken  
expr.Expression = left  
expr.TargetType = self.ParseTypeExpression()  
result = expr  
  _ = isToken
  _ = expr
  return result
}

func (self *TParser) ParseAsExpression(left interface{}) interface{} {
var result interface{}  
var asToken TToken  
var expr *TTypeCastExpression  
asToken = self.CurToken  
self.NextToken()  
expr = &TTypeCastExpression{}  
expr.Token = asToken  
expr.Expression = left  
expr.TargetType = self.ParseTypeExpression()  
result = expr  
  _ = asToken
  _ = expr
  return result
}

func (self *TParser) ParseStringInterpolation() interface{} {
var result interface{}  
var interp *TStringInterpolation  
var raw string  
var i int64  
var depth int64  
var currentText string  
var exprStr string  
var j int64  
interp = &TStringInterpolation{}  
interp.Parts = nil  
raw = self.CurToken.Literal  
currentText = ""  
i = 0  
for (i < int64(len(raw)))   {
if (((raw[i:(i + 1)] == "$") && ((i + 1) < int64(len(raw)))) && (raw[(i + 1):(i + 2)] == "{"))     {
if (currentText != "")       {
var lit *TStringLiteral        
lit = &TStringLiteral{}        
lit.Value = currentText        
interp.Parts = append(interp.Parts, lit        )
currentText = ""        
      }
depth = 1      
j = (i + 2)      
for ((j < int64(len(raw))) && (depth > 0))       {
if (raw[j:(j + 1)] == "{")         {
depth = (depth + 1)          
        }
if (raw[j:(j + 1)] == "}")         {
depth = (depth - 1)          
        }
j = (j + 1)        
      }
exprStr = raw[(i + 2):(j - 1)]      
var subLex *TLexer      
subLex = NewLexer(exprStr)      
var subParser *TParser      
subParser = NewParser(subLex)      
var subExpr interface{}      
subExpr = subParser.ParseExpression(PREC_LOWEST)      
if (subExpr != nil)       {
interp.Parts = append(interp.Parts, subExpr        )
      }
i = j      
    } else {
currentText = (currentText + raw[i:(i + 1)])      
i = (i + 1)      
    }
  }
if (currentText != "")   {
var lit2 *TStringLiteral    
lit2 = &TStringLiteral{}    
lit2.Value = currentText    
interp.Parts = append(interp.Parts, lit2    )
  }
result = interp  
  _ = interp
  _ = raw
  _ = i
  _ = depth
  _ = currentText
  _ = exprStr
  _ = j
  return result
}

func (self *TParser) ParseAnonymousFunction() interface{} {
var result interface{}  
var kind TToken  
var params []*TParameter  
var param *TParameter  
kind = self.CurToken  
self.NextToken()  
params = nil  
if self.CurTokenIs(tkLParen)   {
self.NextToken()    
for ((self.CurTokenIs(tkRParen) == false) && (self.CurTokenIs(tkEOF) == false))     {
if self.CurTokenIs(tkIdent)       {
param = &TParameter{}        
param.Token = self.CurToken        
param.Name = self.CurToken.Literal        
self.NextToken()        
if self.CurTokenIs(tkColon)         {
self.NextToken()          
param.ParamType = self.ParseTypeExpression()          
        }
params = append(params, param        )
if self.CurTokenIs(tkSemicolon)         {
self.NextToken()          
        } else {
if self.CurTokenIs(tkComma)           {
self.NextToken()            
          }
        }
      } else {
self.NextToken()        
      }
    }
if self.CurTokenIs(tkRParen)     {
self.NextToken()      
    }
  }
if (kind.TokenType == tkFunction)   {
if self.CurTokenIs(tkColon)     {
self.NextToken()      
self.ParseTypeExpression()      
    }
  }
if self.CurTokenIs(tkSemicolon)   {
self.NextToken()    
  }
for (self.CurTokenIs(tkVar) || self.CurTokenIs(tkConst))   {
if self.CurTokenIs(tkVar)     {
self.NextToken()      
for self.IsIdentOrSoftKeyword()       {
self.ParseSingleVarDecl()        
for self.CurTokenIs(tkSemicolon)         {
self.NextToken()          
        }
      }
    } else {
self.NextToken()      
for self.IsIdentOrSoftKeyword()       {
self.ParseSingleConstDecl()        
for self.CurTokenIs(tkSemicolon)         {
self.NextToken()          
        }
      }
    }
  }
var lam *TLambdaExpression  
lam = &TLambdaExpression{}  
lam.Token = kind  
lam.Parameters = params  
if self.CurTokenIs(tkBegin)   {
self.NextToken()    
var block *TBlockStatement    
block = &TBlockStatement{}    
block.Statements = nil    
var iter int64    
iter = 0    
for ((self.CurTokenIs(tkEnd) == false) && (self.CurTokenIs(tkEOF) == false))     {
iter = (iter + 1)      
if (iter > 1000)       {
        break
      }
var s interface{}      
s = self.ParseStatement()      
if (s != nil)       {
block.Statements = append(block.Statements, s        )
      } else {
if (self.CurTokenIs(tkSemicolon) == false)         {
self.NextToken()          
        }
      }
for self.CurTokenIs(tkSemicolon)       {
self.NextToken()        
      }
    }
lam.Body = block    
  }
result = lam  
  _ = kind
  _ = params
  _ = param
  return result
}

func (self *TParser) ParseExpressionList(endType TTokenType) []interface{} {
var result []interface{}  
var list []interface{}  
list = nil  
if self.PeekTokenIs(endType)   {
self.NextToken()    
result = list    
    return result
  }
self.NextToken()  
list = append(list, self.ParseExpression(PREC_LOWEST)  )
for self.PeekTokenIs(tkComma)   {
self.NextToken()    
self.NextToken()    
list = append(list, self.ParseExpression(PREC_LOWEST)    )
  }
if (self.ExpectPeek(endType) == false)   {
result = nil    
    return result
  }
result = list  
  _ = list
  return result
}

func (self *TParser) ParseTypeExpression() interface{} {
var result interface{}  
if self.CurTokenIs(tkLParen)   {
var enumType *TEnumType    
enumType = self.TryParseEnumType()    
if (enumType != nil)     {
result = enumType      
      return result
    }
  }
if self.CurTokenIs(tkIdent)   {
var name string    
name = self.CurToken.Literal    
self.NextToken()    
if self.CurTokenIs(tkLT)     {
self.NextToken()      
var gen *TGenericType      
gen = &TGenericType{}      
gen.Base = name      
gen.TypeParams = nil      
for ((self.CurTokenIs(tkGT) == false) && (self.CurTokenIs(tkEOF) == false))       {
gen.TypeParams = append(gen.TypeParams, self.ParseTypeExpression()        )
if self.CurTokenIs(tkComma)         {
self.NextToken()          
        }
      }
if self.CurTokenIs(tkGT)       {
self.NextToken()        
      }
result = gen      
      return result
    }
var ident *TIdentifier    
ident = &TIdentifier{}    
ident.Value = name    
result = ident    
    return result
  }
if self.CurTokenIs(tkMap)   {
self.NextToken()    
var mapType *TMapType    
mapType = &TMapType{}    
if self.CurTokenIs(tkLBracket)     {
self.NextToken()      
mapType.KeyType = self.ParseTypeExpression()      
if self.CurTokenIs(tkRBracket)       {
self.NextToken()        
      }
    }
mapType.ValueType = self.ParseTypeExpression()    
result = mapType    
    return result
  }
if self.CurTokenIs(tkVariant)   {
self.NextToken()    
var vtype *TVariantType    
vtype = &TVariantType{}    
vtype.Cases = nil    
for self.IsIdentOrSoftKeyword()     {
var caseNode *TVariantCase      
caseNode = &TVariantCase{}      
caseNode.Name = self.CurToken.Literal      
self.NextToken()      
if self.CurTokenIs(tkColon)       {
self.NextToken()        
caseNode.CaseType = self.ParseTypeExpression()        
      }
vtype.Cases = append(vtype.Cases, caseNode      )
for self.CurTokenIs(tkSemicolon)       {
self.NextToken()        
      }
    }
if self.CurTokenIs(tkEnd)     {
self.NextToken()      
    }
result = vtype    
    return result
  }
if self.CurTokenIs(tkArray)   {
self.NextToken()    
var arrayType *TArrayType    
arrayType = &TArrayType{}    
arrayType.Dynamic = true    
if self.CurTokenIs(tkLBracket)     {
self.NextToken()      
var lowerBound interface{}      
lowerBound = self.ParseExpression(PREC_LOWEST)      
self.NextToken()      
arrayType.Dynamic = false      
if self.CurTokenIs(tkDotDot)       {
self.NextToken()        
var upperBound interface{}        
upperBound = self.ParseExpression(PREC_LOWEST)        
self.NextToken()        
var sub *TInfixExpression        
sub = &TInfixExpression{}        
sub.Left = upperBound        
sub.Operator = "-"        
sub.Right = lowerBound        
var add *TInfixExpression        
add = &TInfixExpression{}        
add.Left = sub        
add.Operator = "+"        
var one *TIntegerLiteral        
one = &TIntegerLiteral{}        
one.Value = 1        
add.Right = one        
arrayType.Size = add        
      } else {
arrayType.Size = lowerBound        
      }
if self.CurTokenIs(tkRBracket)       {
self.NextToken()        
      }
    }
if self.CurTokenIs(tkOf)     {
self.NextToken()      
arrayType.ElementType = self.ParseTypeExpression()      
    }
result = arrayType    
    return result
  }
if self.CurTokenIs(tkRecord)   {
self.NextToken()    
var rec *TRecordType    
rec = &TRecordType{}    
rec.Fields = nil    
var depth int64    
depth = 1    
for ((depth > 0) && (self.CurTokenIs(tkEOF) == false))     {
if self.CurTokenIs(tkRecord)       {
depth = (depth + 1)        
      } else {
if self.CurTokenIs(tkEnd)         {
depth = (depth - 1)          
if (depth == 0)           {
self.NextToken()            
          }
        }
      }
if (depth > 0)       {
if self.CurTokenIs(tkVar)         {
self.NextToken()          
for self.CurTokenIs(tkIdent)           {
var field *TVarDecl            
field = self.ParseSingleVarDecl()            
if (field != nil)             {
rec.Fields = append(rec.Fields, field              )
            }
for self.CurTokenIs(tkSemicolon)             {
self.NextToken()              
            }
          }
        } else {
if self.CurTokenIs(tkIdent)           {
var field2 *TVarDecl            
field2 = self.ParseSingleVarDecl()            
if (field2 != nil)             {
rec.Fields = append(rec.Fields, field2              )
            }
for self.CurTokenIs(tkSemicolon)             {
self.NextToken()              
            }
          } else {
self.NextToken()            
          }
        }
      }
    }
result = rec    
    return result
  }
if (self.CurTokenIs(tkFunction) || self.CurTokenIs(tkProcedure))   {
var funcToken TToken    
funcToken = self.CurToken    
self.NextToken()    
var funcType *TIdentifier    
funcType = &TIdentifier{}    
funcType.Token = funcToken    
funcType.Value = funcToken.Literal    
if self.CurTokenIs(tkLParen)     {
self.NextToken()      
var d int64      
d = 1      
for ((d > 0) && (self.CurTokenIs(tkEOF) == false))       {
if self.CurTokenIs(tkLParen)         {
d = (d + 1)          
        }
if self.CurTokenIs(tkRParen)         {
d = (d - 1)          
        }
if (d > 0)         {
self.NextToken()          
        }
      }
if self.CurTokenIs(tkRParen)       {
self.NextToken()        
      }
    }
if self.CurTokenIs(tkColon)     {
self.NextToken()      
self.ParseTypeExpression()      
    }
result = funcType    
    return result
  }
var fallback *TIdentifier  
fallback = &TIdentifier{}  
fallback.Token = self.CurToken  
fallback.Value = self.CurToken.Literal  
result = fallback  
  return result
}

func (self *TParser) TryParseEnumType() *TEnumType {
var result *TEnumType  
var savedCur TToken  
var savedPeek TToken  
var enum *TEnumType  
savedCur = self.CurToken  
savedPeek = self.PeekToken  
self.NextToken()  
if (self.CurTokenIs(tkIdent) == false)   {
self.CurToken = savedCur    
self.PeekToken = savedPeek    
result = nil    
  }
if ((self.PeekTokenIs(tkComma) == false) && (self.PeekTokenIs(tkRParen) == false))   {
self.CurToken = savedCur    
self.PeekToken = savedPeek    
result = nil    
  }
enum = &TEnumType{}  
enum.Names = nil  
for self.CurTokenIs(tkIdent)   {
enum.Names = append(enum.Names, self.CurToken.Literal    )
self.NextToken()    
if self.CurTokenIs(tkComma)     {
self.NextToken()      
    } else {
      break
    }
  }
if self.CurTokenIs(tkRParen)   {
self.NextToken()    
  }
result = enum  
  _ = savedCur
  _ = savedPeek
  _ = enum
  return result
}

func main() {
ErrList = &TErrorList{}  
Files = nil  
if (int64(len(os.Args[1:])) > 0)   {
for i = 0; i <= (int64(len(os.Args[1:])) - 1); i    ++ {
Files = append(Files, os.Args[1:][i]      )
    }
  } else {
Files = append(Files, "token.klx"    )
Files = append(Files, "error.klx"    )
Files = append(Files, "ast.klx"    )
Files = append(Files, "lexer.klx"    )
Files = append(Files, "parser.klx"    )
Files = append(Files, "generator.klx"    )
  }
Programs = nil  
for i = 0; i <= (int64(len(Files)) - 1); i  ++ {
Programs = append(Programs, nil    )
  }
AllOk = true  
for i = 0; i <= (int64(len(Files)) - 1); i  ++ {
SourceFile = Files[i]    
SourceCode = func() string { data, _ := os.ReadFile(SourceFile); return string(data) }()    
if (SourceCode == "")     {
ErrList.AddError(SourceFile, 0, 0, "cannot read file")      
AllOk = false      
    } else {
Lex = NewLexer(SourceCode)      
Par = NewParser(Lex)      
Programs[i] = Par.ParseProgram()      
if (int64(len(Par.Errors)) > 0)       {
var j int64        
for j = 0; j <= (int64(len(Par.Errors)) - 1); j        ++ {
ErrList.AddError(SourceFile, 0, 0, Par.Errors[j])          
AllOk = false          
        }
      }
    }
  }
if AllOk   {
Gen = &TGenerator{}    
GoCode = Gen.GenerateMulti(Programs)    
fmt.Println(GoCode)    
  } else {
fmt.Println(ErrList.ToString())    
  }
}

