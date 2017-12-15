package cmd

import (
	"zygopm/module/cache"
	cfg "zygopm/module/conf"
	"zygopm/module/msg"
	"zygopm/module/repo"
)

//依赖配置文件更新
func Update(installer *repo.Installer) {
	cache.SystemLock()

	conf, err := cfg.ConfigFromToml()

	if err != nil {
		msg.Die(err.Error())
	}
	// 不存在缓存则更新
	if err := installer.Checkout(conf); err != nil {
		msg.Die("更新失败，请检查配置文件: %s", err)
	}
	//checkout 到需要的版本
	//pin值是 git rev-parse HEAD
	if err := repo.SetReference(conf, installer.ResolveTest); err != nil {
		msg.Die("切换版本失败，请检查配置文件: %s", err)
	}

	e := installer.Export(conf)
	if e != nil {
		msg.Die("不能导出到: %s", err)
	}

	// 写锁文件
	hash, err := cfg.Hash()
	if err != nil {
		msg.Die("不能生成配置的hash值")
	}
	lock, err := cfg.NewLockfile(conf.Imports, hash)
	if err != nil {
		msg.Die("生成配置结构失败: %s", err)
	}

	ee := lock.WriteFile()

	if ee != nil {
		msg.Die("生成配文件失败: %s", e)
	}
	return
}
