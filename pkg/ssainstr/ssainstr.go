package ssainstr

import (
	"go/types"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/ssa"
)

func IsCalled(instr ssa.Instruction, fn *types.Func) bool {
	call, ok := instr.(ssa.CallInstruction)
	if !ok {
		return false
	}

	ssaCall := call.Value()
	common := ssaCall.Common()
	if common == nil {
		return false
	}
	val := common.Value

	switch fnval := val.(type) {
	case *ssa.Function:
		for _, block := range fnval.Blocks {
			for _, instr := range block.Instrs {
				if analysisutil.Called(instr, nil, fn) {
					return true
				}
			}
		}
	}

	return false
}

func HasArgs(instr ssa.Instruction, typ types.Type) bool {
	call, ok := instr.(ssa.CallInstruction)
	if !ok {
		return false
	}

	ssaCall := call.Value()
	if ssaCall == nil {
		return false
	}

	for _, arg := range ssaCall.Call.Args {
		if types.Identical(arg.Type(), typ) {
			return true
		}
	}
	return false
}
