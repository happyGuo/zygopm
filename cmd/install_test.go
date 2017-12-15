package cmd

import (
	"testing"
	"zygopm/module/repo"
	"zygopm/module/setting"

	"io/ioutil"
	"path/filepath"
)

func TestInstall(t *testing.T) {

	installer := repo.NewInstaller()
	//设置installer的参数
	installer.Force = true
	installer.Vendor = "C:/zygopm"
	//installer.Home = c.GlobalString("home")
	//setting.SetGopmConfFIleName("test")
	//setting.SetGopmLockFIleName("test")
	Install(installer)

}

func TestTempDir(t *testing.T) {
	tempDir, err := ioutil.TempDir(setting.Tmp, "gopm-vendor")
	f := filepath.FromSlash("./vendor")
	t.Log(tempDir, err, f)
}
