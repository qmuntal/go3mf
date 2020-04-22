// Copyright 2017 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// gen-accessors generates accessor methods for structs with pointer fields.
//
// It is meant to be used by go-github contributors in conjunction with the
// go generate tool before sending a PR to GitHub.
// Please see the CONTRIBUTING.md file for more information.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

const (
	fileSuffix = "_js.go"
)

var (
	verbose = flag.Bool("v", false, "Print verbose log messages")

	sourceTmpl = template.Must(template.New("source").Parse(source))

	// blacklistStructMethod lists "struct.method" combos to skip.
	blacklistStructMethod = map[string]bool{}
	// blacklistStruct lists structs to skip.
	blacklistStruct = map[string]bool{}
)

func logf(fmt string, args ...interface{}) {
	if *verbose {
		log.Printf(fmt, args...)
	}
}

func toJSName(s string) string {
	if s == "" {
		return ""
	}
	s = strings.Replace(s, "ID", "Id", -1)
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func main() {
	flag.Parse()
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, ".", sourceFilter, 0)
	if err != nil {
		log.Fatal(err)
		return
	}

	for pkgName, pkg := range pkgs {
		t := &templateData{
			filename: pkgName + fileSuffix,
			Year:     2017,
			Package:  pkgName,
			Imports: map[string]string{
				"syscall/js": "syscall/js",
			},
			Structs: make(map[string]*structs),
		}
		for filename, f := range pkg.Files {
			logf("Processing %v...", filename)
			if err := t.processAST(f); err != nil {
				log.Fatal(err)
			}
		}
		if err := t.dump(); err != nil {
			log.Fatal(err)
		}
	}
	logf("Done.")
}

func (t *templateData) processAST(f *ast.File) error {
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			// Skip unexported identifiers.
			if !ts.Name.IsExported() {
				logf("Struct %v is unexported; skipping.", ts.Name)
				continue
			}
			// Check if the struct is blacklisted.
			if blacklistStruct[ts.Name.Name] {
				logf("Struct %v is blacklisted; skipping.", ts.Name)
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			t.addStruct(ts.Name.String())
			for _, field := range st.Fields.List {
				if len(field.Names) == 0 {
					continue
				}
				t.processField(ts.Name.String(), field)
			}
		}
	}
	return nil
}

func (t *templateData) processField(receiverType string, field *ast.Field) {
	fieldName := field.Names[0]
	// Skip unexported identifiers.
	if !fieldName.IsExported() {
		logf("Field %v is unexported; skipping.", fieldName)
		return
	}
	// Check if "struct.method" is blacklisted.
	if key := fmt.Sprintf("%v.Get%v", receiverType, fieldName); blacklistStructMethod[key] {
		logf("Method %v is blacklisted; skipping.", key)
		return
	}

	switch x := field.Type.(type) {
	case *ast.ArrayType:
		t.addArrayType(x, receiverType, fieldName.String())
	case *ast.Ident:
		t.addIdent(x, receiverType, fieldName.String())
	case *ast.MapType:
		t.addMapType(x, receiverType, fieldName.String())
	case *ast.SelectorExpr:
		t.addSelectorExpr(x, receiverType, fieldName.String())
	default:
		logf("processAST: type %s, field %q, unknown %T: %+v", receiverType, fieldName, x, x)
	}
}

func sourceFilter(fi os.FileInfo) bool {
	if fi.Name() != "core.go" {
		return false
	}
	return !strings.HasSuffix(fi.Name(), "_test.go") && !strings.HasSuffix(fi.Name(), fileSuffix)
}

func (t *templateData) dump() error {
	SortedStructs := make(pairList, 0, len(t.Structs))
	for k, v := range t.Structs {
		SortedStructs = append(SortedStructs, pair{
			key:   k,
			value: v.sortVal,
		})
	}
	sort.Sort(SortedStructs)

	t.SortedStructs = make([]*structs, len(t.Structs))
	var c int
	for _, m := range SortedStructs {
		s := t.Structs[m.key]
		t.SortedStructs[c] = s
		sort.Sort(byName(s.Getters))
		c++
	}

	if c == 0 {
		logf("No getters for %v; skipping.", t.filename)
		return nil
	}

	var buf bytes.Buffer
	if err := sourceTmpl.Execute(&buf, t); err != nil {
		return err
	}
	clean, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	logf("Writing %v...", t.filename)
	return ioutil.WriteFile(t.filename, clean, 0644)
}

func newGetter(fieldName, fieldType, zeroValue string, namedStruct bool) *getter {
	return &getter{
		sortVal:     strings.ToLower(fieldName),
		FieldName:   fieldName,
		MethodName:  toJSName(fieldName),
		FieldType:   fieldType,
		ZeroValue:   zeroValue,
		NamedStruct: namedStruct,
		IsSlice:     strings.HasPrefix(fieldType, "[]"),
		IsMap:       strings.HasPrefix(fieldType, "map["),
	}
}

func (t *templateData) addArrayType(x *ast.ArrayType, receiverType, fieldName string) {
	var eltType string
	switch elt := x.Elt.(type) {
	case *ast.StarExpr:
		if value, ok := elt.X.(*ast.Ident); ok {
			eltType = value.String()
		} else {
			logf("addArrayType: type %q, field %q: unknown elt type: %T %+v; skipping.", receiverType, fieldName, elt, elt)
			return
		}
	case *ast.Ident:
		eltType = elt.String()
	default:
		logf("addArrayType: type %q, field %q: unknown elt type: %T %+v; skipping.", receiverType, fieldName, elt, elt)
		return
	}

	s := t.Structs[receiverType]
	s.Getters = append(s.Getters, newGetter(fieldName, "[]"+eltType, "nil", false))
}

func (t *templateData) addStruct(receiverType string) {
	if _, ok := t.Structs[receiverType]; ok {
		return
	}
	t.Structs[receiverType] = &structs{
		sortVal:      strings.ToLower(receiverType),
		ReceiverType: receiverType,
		ReceiverVar:  strings.ToLower(receiverType[:1]),
	}
}

func (t *templateData) addIdent(x *ast.Ident, receiverType, fieldName string) {
	var zeroValue string
	var namedStruct = false
	switch x.String() {
	case "int", "int64":
		zeroValue = "0"
	case "string":
		zeroValue = `""`
	case "bool":
		zeroValue = "false"
	case "Timestamp":
		zeroValue = "Timestamp{}"
	default:
		zeroValue = "nil"
		namedStruct = true
	}

	s := t.Structs[receiverType]
	s.Getters = append(s.Getters, newGetter(fieldName, x.String(), zeroValue, namedStruct))
}

func (t *templateData) addMapType(x *ast.MapType, receiverType, fieldName string) {
	var keyType string
	switch key := x.Key.(type) {
	case *ast.Ident:
		keyType = key.String()
	default:
		logf("addMapType: type %q, field %q: unknown key type: %T %+v; skipping.", receiverType, fieldName, key, key)
		return
	}

	var valueType string
	switch value := x.Value.(type) {
	case *ast.Ident:
		valueType = value.String()
	case *ast.StarExpr:
		if value, ok := value.X.(*ast.Ident); ok {
			valueType = value.String()
		} else {
			logf("addMapType: type %q, field %q: unknown value type: %T %+v; skipping.", receiverType, fieldName, value, value)
			return
		}
	default:
		logf("addMapType: type %q, field %q: unknown value type: %T %+v; skipping.", receiverType, fieldName, value, value)
		return
	}

	fieldType := fmt.Sprintf("map[%v]%v", keyType, valueType)
	zeroValue := fmt.Sprintf("map[%v]%v{}", keyType, valueType)
	s := t.Structs[receiverType]
	s.Getters = append(s.Getters, newGetter(fieldName, fieldType, zeroValue, false))
}

func (t *templateData) addSelectorExpr(x *ast.SelectorExpr, receiverType, fieldName string) {
	if strings.ToLower(fieldName[:1]) == fieldName[:1] { // Non-exported field.
		return
	}

	var xX string
	if xx, ok := x.X.(*ast.Ident); ok {
		xX = xx.String()
	}

	switch xX {
	case "time", "json":
		if xX == "json" {
			t.Imports["encoding/json"] = "encoding/json"
		} else {
			t.Imports[xX] = xX
		}
		fieldType := fmt.Sprintf("%v.%v", xX, x.Sel.Name)
		zeroValue := fmt.Sprintf("%v.%v{}", xX, x.Sel.Name)
		if xX == "time" && x.Sel.Name == "Duration" {
			zeroValue = "0"
		}

		s := t.Structs[receiverType]
		s.Getters = append(s.Getters, newGetter(fieldName, fieldType, zeroValue, false))
	default:
		logf("addSelectorExpr: xX %q, type %q, field %q: unknown x=%+v; skipping.", xX, receiverType, fieldName, x)
	}
}

type templateData struct {
	filename      string
	Year          int
	Package       string
	Imports       map[string]string
	Structs       map[string]*structs
	SortedStructs []*structs
}

type structs struct {
	sortVal      string // Lower-case version of "ReceiverType".
	ReceiverVar  string // The one-letter variable name to match the ReceiverType.
	ReceiverType string
	Getters      []*getter
}

type getter struct {
	sortVal     string // Lower-case version of "ReceiverType.FieldName".
	MethodName  string
	FieldName   string
	FieldType   string
	ZeroValue   string
	NamedStruct bool // Getter for named struct.
	IsSlice     bool
	IsMap       bool
}

type pair struct {
	key   string
	value string
}

type pairList []pair

func (b pairList) Len() int           { return len(b) }
func (b pairList) Less(i, j int) bool { return b[i].value < b[j].value }
func (b pairList) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

type byName []*getter

func (b byName) Len() int           { return len(b) }
func (b byName) Less(i, j int) bool { return b[i].sortVal < b[j].sortVal }
func (b byName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

const source = `// Copyright {{.Year}} The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code generated by gen-accessors; DO NOT EDIT.

package {{.Package}}
{{with .Imports}}
import (
  {{- range . -}}
  "{{.}}"
  {{end -}}
)
{{end}}

var (
  jsNS                    = "GO3MF"
  objectConstructor       = js.Global().Get("Object")
  arrayConstructor        = js.Global().Get("Array")
)

var registeredFuncs []js.Func

func JSRelease() {
  for i := range registeredFuncs {
    registeredFuncs[i].Release()
  }
}

func GetterFunc(fn func() interface{}) js.Func {
  jfn := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
    return fn()
  })
  registeredFuncs = append(registeredFuncs, jfn)
  return jfn
}

{{range .SortedStructs}}
// JSValue implements js.Wrapper interface.
func ({{.ReceiverVar}} *{{.ReceiverType}}) JSValue() js.Value {
  v := objectConstructor.New()
  {{$out := .}}
  {{range .Getters -}}
  {{if .IsSlice}}
  arr := arrayConstructor.New(len({{$out.ReceiverVar}}.{{.FieldName}}))
  for i, v := range {{$out.ReceiverVar}}.{{.FieldName}} {
	arr.SetIndex(i, v)
  }
  {{end}}
  {{if .IsMap}}
  {{end}}
  {{if and (not .IsSlice) (not .IsMap) }}
  v.Set("{{.MethodName}}", GetterFunc(func() interface{} { return {{$out.ReceiverVar}}.{{.FieldName}} }))
  {{end}}
  {{end}}
  return v
}
{{end}}
`
