// Command api prints exported API shapes without documentation or private fields.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func main() {
	cmd := exec.Command("go", "list", "-export", "-deps", "-json", ".")
	cmd.Stderr = os.Stderr
	data, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	exports := map[string]string{}
	var target string
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var entry struct{ ImportPath, Export string }
		if decodeErr := decoder.Decode(&entry); decodeErr == io.EOF {
			break
		} else if decodeErr != nil {
			panic(decodeErr)
		}
		exports[entry.ImportPath] = entry.Export
		target = entry.ImportPath
	}
	lookup := func(path string) (io.ReadCloser, error) { return os.Open(exports[path]) }
	pkg, err := importer.ForCompiler(token.NewFileSet(), "gc", lookup).Import(target)
	if err != nil {
		panic(err)
	}
	shapes := map[string]string{}
	for _, name := range pkg.Scope().Names() {
		object := pkg.Scope().Lookup(name)
		if !object.Exported() {
			continue
		}
		shape := types.ObjectString(object, qualifier)
		if constant, ok := object.(*types.Const); ok {
			shape += " = " + constant.Val().ExactString()
		}
		if typ, ok := object.(*types.TypeName); ok {
			shape = fmt.Sprintf("type %s alias=%t underlying %s", name, typ.IsAlias(), publicShape(typ.Type().Underlying()))
			if named, ok := typ.Type().(*types.Named); ok {
				var methods []string
				for i := 0; i < named.NumMethods(); i++ {
					method := named.Method(i)
					if method.Exported() {
						methods = append(methods, types.ObjectString(method, qualifier))
					}
				}
				sort.Strings(methods)
				shape += "\n" + strings.Join(methods, "\n")
			}
		}
		shapes[name] = shape
	}
	if err := json.NewEncoder(os.Stdout).Encode(shapes); err != nil {
		panic(err)
	}
}

func qualifier(pkg *types.Package) string { return pkg.Path() }

func publicShape(typ types.Type) string {
	if structure, ok := typ.(*types.Struct); ok {
		var fields []string
		for i := 0; i < structure.NumFields(); i++ {
			field := structure.Field(i)
			if field.Exported() {
				fields = append(fields, fmt.Sprintf("%s embedded=%t tag=%q",
					types.ObjectString(field, qualifier), field.Embedded(), structure.Tag(i)))
			}
		}
		return "struct{" + strings.Join(fields, ";") + "}"
	}
	return types.TypeString(typ, qualifier)
}
