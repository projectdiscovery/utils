package memoize

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemo(t *testing.T) {
	testingFunc := func() (interface{}, error) {
		time.Sleep(10 * time.Second)
		return "b", nil
	}

	m, err := New(WithMaxSize(5))
	require.Nil(t, err)
	start := time.Now()
	_, _, _ = m.Do("test", testingFunc)
	_, _, _ = m.Do("test", testingFunc)
	require.True(t, time.Since(start) < time.Duration(15*time.Second))
}

func TestSrc(t *testing.T) {
	out, err := File(PackageTemplate, "tests/test.go", "test")
	require.Nil(t, err)
	require.True(t, len(out) > 0)
}

func TestParamTypeContextContext(t *testing.T) {
	source := `package example

import "context"

// @memo
func DoSomething(ctx context.Context, key string) (string, error) {
	return "", nil
}
`

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err)

	var params []FuncValue
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Doc == nil {
			return true
		}
		for _, comment := range fn.Doc.List {
			if comment.Text == "// @memo" {
				for idx, param := range fn.Type.Params.List {
					var fv FuncValue
					fv.Index = idx
					for _, name := range param.Names {
						fv.Name = name.String()
					}
					fv.Type = types.ExprString(param.Type)
					params = append(params, fv)
				}
			}
		}
		return false
	})

	require.Len(t, params, 2)
	require.Equal(t, "ctx", params[0].Name)
	require.Equal(t, "context.Context", params[0].Type, "context.Context param type should be a clean string, not fmt.Sprint garbage")
	require.Equal(t, "key", params[1].Name)
	require.Equal(t, "string", params[1].Type)
}
