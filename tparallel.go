package tparallel

import (
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

	trun := analysisutil.MethodOf(testTyp, "Run")
	for _, f := range ssaanalyzer.SrcFuncs {
		if strings.HasPrefix(f.Name(), "Test") && f.Parent() == (*ssa.Function)(nil) {
			testMap[f] = []*ssa.Function{}
			for _, block := range f.Blocks {
				for _, instr := range block.Instrs {
					called := analysisutil.Called(instr, nil, trun)
					if called {
						testMap[f] = appendTestMap(testMap[f], instr)
					}
				}
			}
		}
	}

	return testMap
}

func appendTestMap(subtests []*ssa.Function, instr ssa.Instruction) []*ssa.Function {
	call, ok := instr.(ssa.CallInstruction)
	if !ok {
		return subtests
	}

	ssaCall := call.Value()
	for _, arg := range ssaCall.Call.Args {
		switch arg := arg.(type) {
		case *ssa.Function:
			subtests = append(subtests, arg)
		}
	}
	return subtests
}
