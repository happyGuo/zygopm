package repo

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/urfave/cli"
	"zygopm/module/cache"
	cfg "zygopm/module/conf"
	"zygopm/module/msg"
	"zygopm/module/util"
	"zygopm/module/setting"
)

// 安装对象
type Installer struct {
	// 强制安装
	Force bool
	// 安装目录
	Vendor string
	// 跟踪远程包.
	Updated *UpdateTracker
}

// 初始化
func NewInstaller() *Installer {
	i := &Installer{}
	i.Updated = NewUpdateTracker()
	return i
}

// VendorPath
func (i *Installer) VendorPath() string {
	if i.Vendor != "" {
		return i.Vendor
	}

	return filepath.FromSlash("./vendor")
}

// 依赖lock文件更新
func (i *Installer) Install(lock *cfg.Lockfile, conf *cfg.Config) (*cfg.Config, error) {

	newConf := &cfg.Config{}
	newConf.Name = conf.Name

	newConf.Imports = make(cfg.Dependencies, len(lock.Imports))
	for k, v := range lock.Imports {
		newConf.Imports[k] = cfg.DependencyFromLock(v)
	}

	newConf.DevImports = make(cfg.Dependencies, len(lock.DevImports))
	for k, v := range lock.DevImports {
		newConf.DevImports[k] = cfg.DependencyFromLock(v)
	}

	if len(newConf.Imports) == 0 && len(newConf.DevImports) == 0 {
		msg.Info("没有依赖需要安装")
		return newConf, nil
	}

	msg.Info("在本地缓存开始查找...")

	err := LazyConcurrentUpdate(newConf.Imports, i, newConf)
	if err != nil {
		return newConf, err
	}
	err = LazyConcurrentUpdate(newConf.DevImports, i, newConf)

	return newConf, err
}

//检出下载包到目录
func (i *Installer) Checkout(conf *cfg.Config) error {

	newDeps := []*cfg.Dependency{}
	for _, dep := range conf.Imports {
		//缓存找到
		key, err := cache.Key(dep.Remote())
		if err != nil {
			newDeps = append(newDeps, dep)
			continue
		}
		//有版本管理的包目录
		destPath := filepath.Join(cache.Location(), "src", key)
		if _, err := dep.GetRepo(destPath); err != nil {
			newDeps = append(newDeps, dep)
			continue
		}

		if dep.Reference != "" && VcsVersion(dep) != nil {
			msg.Warn("本地包的版本库内不存在此版本 %s %s", dep.Name, dep.Reference)
			newDeps = append(newDeps, dep)
			continue
		}
		msg.Info("在本地找到可用的版本 %s %s ", dep.Name, dep.Reference)
	}

	//需要下载的包
	if len(newDeps) > 0 {
		msg.Info("开始下载需要更新的包...")
		if err := ConcurrentUpdate(newDeps, i, conf); err != nil {
			return err
		}
	}
	return nil
}


// 从cache内导出到 i.Vendor  目录
func (i *Installer) Export(conf *cfg.Config) error {
	tempDir, err := ioutil.TempDir(setting.Tmp, "gopm-vendor")
	if err != nil {
		return err
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			msg.Err(err.Error())
		}
	}()

	vp := filepath.Join(tempDir, "src")
	err = os.MkdirAll(vp, 0755)

	msg.Info("开始导出...")
	done := make(chan struct{}, concurrentWorkers)
	in := make(chan *cfg.Dependency, concurrentWorkers)
	var wg sync.WaitGroup
	var lock sync.Mutex
	var returnErr error

	for ii := 0; ii < concurrentWorkers; ii++ {
		go func(ch <-chan *cfg.Dependency) {
			for {
				select {
				case dep := <-ch:
					loc := dep.Remote()
					key, err := cache.Key(loc)
					if err != nil {
						msg.Die(err.Error())
					}
					cache.Lock(key)

					cdir := filepath.Join(cache.Location(), "src", key)
					repo, err := dep.GetRepo(cdir)
					if err != nil {
						msg.Die(err.Error())
					}
					msg.Info("--> 正在导出包 %s", dep.Name)
					if err := repo.ExportDir(filepath.Join(vp, filepath.ToSlash(dep.Name))); err != nil {
						msg.Err("导出包失败 %s: %s\n", dep.Name, err)
						// Capture the error while making sure the concurrent
						// operations don't step on each other.
						lock.Lock()
						if returnErr == nil {
							returnErr = err
						} else {
							returnErr = cli.NewMultiError(returnErr, err)
						}
						lock.Unlock()
					}
					cache.Unlock(key)
					wg.Done()
				case <-done:
					return
				}
			}
		}(in)
	}

	for _, dep := range conf.Imports {

		err = os.MkdirAll(filepath.Join(vp, filepath.ToSlash(dep.Name)), 0755)
		if err != nil {
			lock.Lock()
			if returnErr == nil {
				returnErr = err
			} else {
				returnErr = cli.NewMultiError(returnErr, err)
			}
			lock.Unlock()
		}
		wg.Add(1)
		in <- dep

	}

	wg.Wait()

	// Close goroutines setting the version
	for ii := 0; ii < concurrentWorkers; ii++ {
		done <- struct{}{}
	}

	if returnErr != nil {
		return returnErr
	}

	err = util.MoveToInstallDir(vp, i.VendorPath())

	return err

}

//只更新不在缓存的
func LazyConcurrentUpdate(deps []*cfg.Dependency, i *Installer, c *cfg.Config) error {

	//需要新下载的包
	newDeps := []*cfg.Dependency{}
	for _, dep := range deps {

		//缓存找到
		key, err := cache.Key(dep.Remote())
		if err != nil {
			newDeps = append(newDeps, dep)
			continue
		}
		destPath := filepath.Join(cache.Location(), "src", key)

		//版本管理下载到这个目录
		repo, err := dep.GetRepo(destPath)
		if err != nil {
			newDeps = append(newDeps, dep)
			continue
		}
		ver, err := repo.Version()
		if err != nil {
			newDeps = append(newDeps, dep)
			continue
		}
		if dep.Reference != "" {
			ci, err := repo.CommitInfo(dep.Reference)
			if err == nil && ci.Commit == dep.Reference {
				msg.Info("--> 在本地发现所需的版本 %s %s!", dep.Name, dep.Reference)
				continue
			}
		}
		msg.Debug("--> 包 %s 要更新到需要的版本(%s != %s)", dep.Name, ver, dep.Reference)
		newDeps = append(newDeps, dep)
	}
	if len(newDeps) > 0 {
		return ConcurrentUpdate(newDeps, i, c)
	}

	return nil
}

//并行更新下载
func ConcurrentUpdate(deps []*cfg.Dependency, i *Installer, c *cfg.Config) error {
	msg.Info("开始下载本地不存在的包...")
	done := make(chan struct{}, concurrentWorkers)
	in := make(chan *cfg.Dependency, concurrentWorkers)
	var wg sync.WaitGroup
	var lock sync.Mutex
	var returnErr error

	for ii := 0; ii < concurrentWorkers; ii++ {
		go func(ch <-chan *cfg.Dependency) {
			for {
				select {
				case dep := <-ch:
					loc := dep.Remote()
					key, err := cache.Key(loc)
					if err != nil {
						msg.Die(err.Error())
					}
					cache.Lock(key)
					if err := VcsUpdate(dep, i.Force, i.Updated); err != nil {
						msg.Err("更新失败 %s: %s\n", dep.Name, err)

						lock.Lock()
						if returnErr == nil {
							returnErr = err
						} else {
							returnErr = cli.NewMultiError(returnErr, err)
						}
						lock.Unlock()
					}
					cache.Unlock(key)
					wg.Done()
				case <-done:
					return
				}
			}
		}(in)
	}

	for _, dep := range deps {
		wg.Add(1)
		in <- dep

	}
	wg.Wait()

	for ii := 0; ii < concurrentWorkers; ii++ {
		done <- struct{}{}
	}

	return returnErr
}