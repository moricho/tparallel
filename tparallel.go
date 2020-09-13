package tparallel

import (
	"go/types"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

const doc = "tparallel detects inappropriate usage of t.Parallel() method in your Go test codes."

// Analyzer analyzes Go test codes whether they use t.Parallel() appropriately
// by using SSA (Single Static Assignment)
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

	obj := analysisutil.ObjectOf(pass, "testing", "T")
	if obj == nil {
		// skip checking
		return nil, nil
	}
	testTyp, testPkg := obj.Type(), obj.Pkg()

	p, _, _ := types.LookupFieldOrMethod(testTyp, true, testPkg, "Parallel")
	parallel, _ := p.(*types.Func)
	c, _, _ := types.LookupFieldOrMethod(testTyp, true, testPkg, "Cleanup")
	cleanup, _ := c.(*types.Func)

	testMap := getTestMap(ssaanalyzer, testTyp) // ex. {Test1: [TestSub1, TestSub2], Test2: [TestSub1, TestSub2, TestSub3], ...}
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
			pass.Reportf(top.Pos(), "%s should call t.Parallel on the top level as well as its subtests", top.Name())
		} else if isParallelTop {
			pass.Reportf(top.Pos(), "%s's subtests should call t.Parallel", top.Name())
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

func isCalled(f *ssa.Function, typ *types.Func) bool {
	block := f.Blocks[0]
	for _, instr := range block.Instrs {
		called := analysisutil.Called(instr, nil, typ)
		if called {
			return true
		}
	}
	return false
}
