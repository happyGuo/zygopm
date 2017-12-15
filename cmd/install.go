package cmd

import (
	"zygopm/module/cache"
	cfg "zygopm/module/conf"
	"zygopm/module/msg"

	"zygopm/module/repo"
)

// 基于配置文件安装依赖包.
func Install(installer *repo.Installer) {
	cache.SystemLock()
	conf, err := cfg.ConfigFromToml()

	if err != nil {
		msg.Die(err.Error())
	}

	if !cfg.HasLock() {
		msg.Info("锁文件不存在，开始执行update")
		Update(installer)
		return
	}
	// 读取锁文件
	lock, err := cfg.ReadLockFile()
	if err != nil {
		//msg.Die("不能加载锁文件")
		msg.Warn("锁文件不存在,以配置文件执行安装")
		Update(installer)
		return
	}

	//不验证配置文件是否改动 todo
	hash, err := cfg.Hash()
	if err != nil {
		msg.Die(err.Error())
	}
	if hash != lock.Hash {
		msg.Warn("配置文件已经修改,将以配置来更新安装包...")
		Update(installer)
		return
	}

	newConf, err := installer.Install(lock, conf)
	if err != nil {
		msg.Die("安装失败: %s", err)
	}

	msg.Info("设置版本")

	if err := repo.SetReference(newConf); err != nil {
		msg.Die("设置版本失败", err)
	}

	err = installer.Export(newConf)
	if err != nil {
		msg.Die("不能导出到: %s", err)
	}
}
