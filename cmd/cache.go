package cmd

import (
	"os"

	"zygopm/module/cache"
	"zygopm/module/msg"
)

//清除本地缓存
func CacheClear() {
	l := cache.Location()

	err := os.RemoveAll(l)
	if err != nil {
		msg.Die("清除缓存失败: %s", err)
	}

	cache.SetupReset()
	cache.Setup()

	msg.Info("缓存已清除")
}
