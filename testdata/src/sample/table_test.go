package sample

import (
	"testing"
)

func Test_Table1(t *testing.T) { // want "Test_Table1 should call t.Parallel on the top level" "Test_Table1 should use t.Cleanup"
	teardown := setup("Test_Func1")
	defer teardown()

	tests := []struct {
		name string
	}{
		{
			name: "Table1_Sub1",
		},
		{
			name: "Table1_Sub2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			call(tt.name)
		})
	}
}
