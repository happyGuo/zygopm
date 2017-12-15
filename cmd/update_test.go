package cmd

import (
	"testing"
	"zygopm/module/repo"
)

func TestUpdate(t *testing.T) {

	installer := repo.NewInstaller()
	//设置installer的参数
	installer.Force = true
	installer.Vendor = "D:/zygopm"
	//installer.Home = c.GlobalString("home")
	Update(installer)

}
