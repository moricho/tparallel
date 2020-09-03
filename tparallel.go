package tparallel

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ssa"
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
	ssaanalyzer := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	testTyp := analysisutil.TypeOf(pass, "testing", "*T")
	if testTyp == nil {
		// skip checking
		return nil, nil
	}

	parallel := analysisutil.MethodOf(testTyp, "Parallel")

	testMap := getTestMap(ssaanalyzer, testTyp)
	for top, subs := range testMap {
		isParallelTop, isPararellSub := false, false
		for _, block := range top.Blocks {
			for _, instr := range block.Instrs {
				called := analysisutil.Called(instr, nil, parallel)
				if called {
					isParallelTop = true
					break
				}
			}
		}

		for _, sub := range subs {
			for _, block := range sub.Blocks {
				for _, instr := range block.Instrs {
					called := analysisutil.Called(instr, nil, parallel)
					if called {
						isPararellSub = true
						break
					}
				}
			}
		}

		if isParallelTop == isPararellSub {
			continue
		} else if isPararellSub {
			pass.Reportf(top.Pos(), "%s should call t.Parallel() on the top level", top.Name())
		} else if isParallelTop {
			pass.Reportf(top.Pos(), "%s's sub tests should call t.Parallel()", top.Name())
		}
	}

	return nil, nil
}

func getTestMap(ssaanalyzer *buildssa.SSA, testTyp types.Type) map[*ssa.Function][]*ssa.Function {
	testMap := map[*ssa.Function][]*ssa.Function{}

	for _, f := range ssaanalyzer.SrcFuncs {
		if strings.HasPrefix(f.Name(), "Test") && f.Parent() == (*ssa.Function)(nil) {
			testMap[f] = []*ssa.Function{}
		}
	}

	for _, f := range ssaanalyzer.SrcFuncs {
		p := f.Parent()
		if _, ok := testMap[p]; !ok {
			continue
		}

		if len(f.Params) == 1 && types.Identical(testTyp, f.Params[0].Type()) {
			testMap[p] = append(testMap[p], f)
		}
	}

	return testMap
}

func getTestMap2(ssaanalyzer *buildssa.SSA, testTyp types.Type) map[*ssa.Function][]*ssa.Function {
	testMap := map[*ssa.Function][]*ssa.Function{}

	trun := analysisutil.MethodOf(testTyp, "Run")
	for _, f := range ssaanalyzer.SrcFuncs {
		if strings.HasPrefix(f.Name(), "Test") && f.Parent() == (*ssa.Function)(nil) {
			testMap[f] = []*ssa.Function{}
			for _, block := range f.Blocks {
				for _, instr := range block.Instrs {
					called := analysisutil.Called(instr, nil, trun)
					if called {
						fmt.Println("ここでinstrの中から、t.Run()の引数である無名関数を取得したい")
						for _, v := range instr.Operands(nil) {
							fmt.Printf("%+v\n", v)
						}
					}
				}
			}
		}
	}

	return testMap
}
