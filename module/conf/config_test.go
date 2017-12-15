package conf

import (
	"os"
	"testing"
)

func TestReadLockFile(t *testing.T) {
	c, _ := ReadLockFile()
	t.Log(c)
}

func TestHash(t *testing.T) {
	Hash()
}
func TestGopath(t *testing.T) {
	s := os.Getenv("GOPATH")
	t.Log(s)
}

func TestConfigFromToml(t *testing.T) {
	c, _ := ConfigFromToml()
	t.Log(c)
}
