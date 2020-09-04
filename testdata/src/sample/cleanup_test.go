package sample

import "testing"

func Test_Cleanup1(t *testing.T) { // want "Test_Cleanup1 should use t.Cleanup"
	teardown := setup("Test_Cleanup1")
	defer teardown()

	t.Parallel()

	t.Run("Cleanup1_Sub1", func(t *testing.T) {
		call("Cleanup1_Sub1")
		t.Parallel()
	})

	t.Run("Cleanup1_Sub2", func(t *testing.T) {
		call("Cleanup1_Sub2")
	})
}

func Test_Cleanup2(t *testing.T) { // OK
	teardown := setup("Test_Cleanup2")
	defer teardown()

	t.Run("Cleanup2_Sub1", func(t *testing.T) {
		call("Cleanup2_Sub1")
	})

	t.Run("Cleanup2_Sub2", func(t *testing.T) {
		call("Cleanup2_Sub2")
	})
}
