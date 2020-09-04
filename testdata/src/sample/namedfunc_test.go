package sample

import "testing"

func testFn1(t *testing.T) {
	t.Parallel()
	call("Named1_Sub")
}

func testFn2(t *testing.T) {
	call("Named2_Sub")
}

func testFn3(t *testing.T) {
	t.Parallel()
	call("Named3_Sub")
}

func Test_Named1(t *testing.T) { // want "Test_Named1 should call t.Parallel on the top level"
	teardown := setup("Test_Named1")
	t.Cleanup(teardown)

	fn := testFn1

	t.Run("Named1_Sub1", fn)

	t.Run("Named1_Sub2", fn)
}

func Test_Named2(t *testing.T) { // want "Test_Named2's sub tests should call t.Parallel"
	teardown := setup("Test_Named1")
	t.Cleanup(teardown)
	t.Parallel()

	fn := testFn2

	t.Run("Named2_Sub1", fn)

	t.Run("Named2_Sub2", fn)
}

func Test_Named3(t *testing.T) { // OK
	teardown := setup("Test_Named3")
	t.Cleanup(teardown)
	t.Parallel()

	fn := testFn3

	t.Run("Named3_Sub1", fn)

	t.Run("Named3_Sub2", fn)
}
