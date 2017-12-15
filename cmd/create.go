package cmd

import (
	"os"

	"sort"
	"strings"

	"path"
	cfg "zygopm/module/conf"
	"zygopm/module/dependency"
	"zygopm/module/msg"
	"zygopm/module/setting"
	"zygopm/module/util"
)

func Create(base string) {
	//配置文件路径
	p := path.Join(base, setting.GetGopmConfFileName())
	// 生成依赖
	c := genDeps(base)
	msg.Info("写入配置文件 (%s)", p)
	if err := c.WriteFile(p); err != nil {
		msg.Die("写入失败 %s: %s", p, err)
	}

}

//扫描依赖
func genDeps(base string) *cfg.Config {

	buildContext, err := util.GetBuildContext()
	name := buildContext.PackageName(base)
	msg.Info("生成配置文件并检查依赖")
	config := new(cfg.Config)
	config.Name = name
	// 返回依赖解析器
	r, err := dependency.NewResolver(base)

	if err != nil {
		msg.Die("创建依赖分析失败: %s", err)
	}

	sortable, _, err := r.ResolveLocal(false)
	if err != nil {
		msg.Die("分析本地依赖失败: %s", err)
	}

	sort.Strings(sortable)

	vpath := r.VendorDir
	if !strings.HasSuffix(vpath, "/") {
		vpath = vpath + string(os.PathSeparator)
	}

	for _, pa := range sortable {
		n := strings.TrimPrefix(pa, vpath)
		root, _ := util.NormalizeName(n)

		if !config.Imports.Has(root) {
			msg.Info("--> 找到依赖 %s\n", n)
			d := &cfg.Dependency{
				Name: root,
			}

			config.Imports = append(config.Imports, d)
		}
	}
	return config
}
