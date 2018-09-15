package products

import (
	"testing"
)

func TestDBProducts_MainError(t *testing.T) {
	x := MustOpen(":memory:")
	x.testCreateParty()
}

func (x DBProducts) testAddMainErrorData() {
	x.DB.MustExec(`
INSERT INTO
`)
}
