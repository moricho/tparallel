package tparallel

import (
	"go/types"
	"strings"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

const doc = "tparallel detects inappropriate usage of t.Parallel() method in your Go test codes."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "tparallel",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
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

	testCommonTyp := analysisutil.TypeOf(pass, "testing", "common")
	if testCommonTyp == nil {
		// skip checking
		return nil, nil
	}

	parallel := analysisutil.MethodOf(testTyp, "Parallel")
	cleanup := analysisutil.MethodOf(testCommonTyp, "Cleanup")

	testMap := getTestMap(ssaanalyzer, testTyp) // {Test1: [TestSub1, TestSub2], Test2: [TestSub1, TestSub2, TestSub3], ...}
	for top, subs := range testMap {
		isParallelTop := isCalled(top, parallel)

		isPararellSub := false
		for _, sub := range subs {
			isPararellSub = isCalled(sub, parallel)
			if isPararellSub {
				break
			}
		}

		if isDeferCalled(top) {
			useCleanup := isCalled(top, cleanup)
			if isPararellSub && !useCleanup {
				pass.Reportf(top.Pos(), "%s should use t.Cleanup instead of defer", top.Name())
			}
		}

		if isParallelTop == isPararellSub {
			continue
		} else if isPararellSub {
			pass.Reportf(top.Pos(), "%s should call t.Parallel on the top level", top.Name())
		} else if isParallelTop {
			pass.Reportf(top.Pos(), "%s's sub tests should call t.Parallel", top.Name())
		}
	}

	return nil, nil
}

func isDeferCalled(f *ssa.Function) bool {
	for _, block := range f.Blocks {
		for _, instr := range block.Instrs {
			switch instr.(type) {
			case *ssa.Defer:
				return true
			}
		}
	}
	return false
}

// isCalled checks whether a given types.Func is called in a ssa.Function
func isCalled(f *ssa.Function, typ *types.Func) bool {
	for _, block := range f.Blocks {
		for _, instr := range block.Instrs {
			called := analysisutil.Called(instr, nil, typ)
			if called {
				return true
			}
		}
	}
	return false
}

// getTestMap gets a set of a top-level test and its sub-tests
func getTestMap(ssaanalyzer *buildssa.SSA, testTyp types.Type) map[*ssa.Function][]*ssa.Function {
	testMap := map[*ssa.Function][]*ssa.Function{}

	trun := analysisutil.MethodOf(testTyp, "Run")
	for _, f := range ssaanalyzer.SrcFuncs {
		if !strings.HasPrefix(f.Name(), "Test") || !(f.Parent() == (*ssa.Function)(nil)) {
			continue
		}
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

	return testMap
}

// appendTestMap converts ssa.Instruction to ssa.Function and append it to a given sub-test slice
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
		case *ssa.MakeClosure:
			fn, _ := arg.Fn.(*ssa.Function)
			subtests = append(subtests, fn)
		}
	}

	return subtests
}
