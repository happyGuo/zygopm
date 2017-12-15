package path

import (
	"path/filepath"
	"testing"
)

func TestMoveToInstallDir(t *testing.T) {
	td, err := filepath.Abs("../testdata")
	td2, err := filepath.Abs("../testdata1")
	if err != nil {
		t.Fatal(err)
	}
	MoveToInstallDir(td, td2)
}
