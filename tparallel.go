package tparallel

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const doc = "tparallel is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "tparallel",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		buildssa.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	testTyp := analysisutil.TypeOf(pass, "testing", "*T")
	if testTyp == nil {
		return nil, errors.New("analyzer does not find *testing.T type")
	}

	parallel := analysisutil.MethodOf(testTyp, "Parallel")
	if parallel == nil {
		return nil, errors.New("analyzer does not find (testing.T).Parallel method")
	}

	ssa, _ := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	if ssa == nil {
		return nil, nil
	}
	for _, f := range ssa.SrcFuncs {
		fmt.Println(f.Name())
		if !strings.HasPrefix(f.Name(), "Test") {
			continue
		}
		for _, block := range f.Blocks {
			for i, instr := range block.Instrs {
				called, ok := analysisutil.CalledFrom(block, i, testTyp, parallel)
				if ok && !called {
					pass.Reportf(instr.Pos(), "NG")
				}
			}
		}
	}

	// inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// nodeFilter := []ast.Node{
	// 	(*ast.FuncDecl)(nil),
	// }

	// inspect.Preorder(nodeFilter, func(n ast.Node) {
	// 	switch n := n.(type) {
	// 	case *ast.FuncDecl:
	// 		if strings.HasPrefix(n.Name.String(), "Test") {
	// 			// isParallelTop := false
	// 			// for _, stmt := range n.Body.List {
	// 			// 	// called, _ := stmt.(*ast.DeferStmt)
	// 			// 	// if called == nil {
	// 			// 	// 	continue
	// 			// 	// }

	// 			// 	// fmt.Println(called.Call.Fun)
	// 			// }
	// 		}
	// 	}
	// })

	return nil, nil
}
