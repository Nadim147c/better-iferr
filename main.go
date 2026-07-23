package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/tools/go/packages"
)

func main() {
	log.SetFlags(0)

	// current cursor position
	var pos int
	// current filename
	var filename string = "main.go"

	pflag.IntVarP(&pos, "position", "p", pos, "Cursor position in buffer")
	pflag.StringVarP(&filename, "filename", "f", filename, "Current file name")

	pflag.Parse()

	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		log.Fatal(err)
	}

	var buf []byte
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeNamedPipe != 0 {
		buf, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		buf, err = os.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
	}

	snippet, err := generateErrReturn(absFilePath, buf, pos)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println(snippet)
}

func generateErrReturn(filePath string, fileContent []byte, byteOffset int) (string, error) {
	fset := token.NewFileSet()

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax,
		Fset: fset,
		Overlay: map[string][]byte{
			filePath: fileContent,
		},
	}

	pkgs, err := packages.Load(cfg, filepath.Dir(filePath))
	if err != nil || len(pkgs) == 0 {
		return "", fmt.Errorf("failed to load package: %v", err)
	}

	pkg := pkgs[0]

	var targetFile *ast.File
	for _, f := range pkg.Syntax {
		tokenFile := fset.File(f.Pos())
		if tokenFile != nil && tokenFile.Name() == filePath {
			targetFile = f
			break
		}
	}
	if targetFile == nil {
		return "", fmt.Errorf("file not found in syntax tree")
	}

	// Convert byte offset to token.Pos
	tokenFile := fset.File(targetFile.Pos())
	targetPos := tokenFile.Pos(byteOffset)

	var targetFuncType *ast.FuncType
	ast.Inspect(targetFile, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		if n.Pos() <= targetPos && targetPos <= n.End() {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				targetFuncType = fn.Type
			case *ast.FuncLit:
				targetFuncType = fn.Type
			}
			return true // Keep digging deeper for nested anonymous functions
		}
		return false
	})

	if targetFuncType == nil || targetFuncType.Results == nil {
		return "", fmt.Errorf("no enclosing function with return values found at position: %v", byteOffset)
	}

	var returns []string
	lastIndex := len(targetFuncType.Results.List) - 1
	for i, field := range targetFuncType.Results.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}

		tv, ok := pkg.TypesInfo.Types[field.Type]
		if !ok {
			return "", fmt.Errorf("type info not found for return field")
		}

		if i == lastIndex {
			if _, ok := tv.Type.Underlying().(*types.Interface); !ok {
				return "", fmt.Errorf("last return type is not error")
			}
			returns = append(returns, "err")
			continue
		}

		zeroVal := formatZeroValue(tv.Type)
		for i := 0; i < count; i++ {
			returns = append(returns, zeroVal)
		}
	}

	// 5. Build the "if err != nil" block
	var sb strings.Builder
	sb.WriteString("if err != nil {\n  return ")
	for i, r := range returns {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(r)
	}
	sb.WriteString("\n}")

	return sb.String(), nil
}

// formatZeroValue evaluates the underlying type to decide how to render the zero-value expression
func formatZeroValue(t types.Type) string {
	named, isNamed := t.(*types.Named)
	underlying := t.Underlying()
	switch u := underlying.(type) {
	case *types.Basic:
		infoFlags := u.Info()
		switch {
		case infoFlags&types.IsBoolean != 0:
			return "false"
		case infoFlags&types.IsNumeric != 0:
			return "0"
		case infoFlags&types.IsString != 0:
			return `""`
		}

	case *types.Interface, *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature:
		return "nil"

	case *types.Array:
		if isNamed {
			return named.Obj().Name() + "{}"
		}
		return fmt.Sprintf("[%d]%s{}", u.Len(), u.Elem())

	case *types.Struct:
		if isNamed {
			return named.Obj().Name() + "{}"
		}
		qualifier := func(p *types.Package) string { return p.Name() }
		return types.TypeString(t, qualifier) + "{}"
	}

	return "nil"
}
