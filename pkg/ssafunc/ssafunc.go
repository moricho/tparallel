package ssafunc

import (
	"go/types"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/ssa"
)

func IsDeferCalled(f *ssa.Function) bool {
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

func IsCalled(f *ssa.Function, fn *types.Func) bool {
	block := f.Blocks[0]
	for _, instr := range block.Instrs {
		called := analysisutil.Called(instr, nil, fn)
		if called {
			return true
		}
	}
	return false
}
