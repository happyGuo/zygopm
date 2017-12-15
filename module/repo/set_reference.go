package repo

import (
	"sync"

	"github.com/urfave/cli"
	"zygopm/module/cache"
	cfg "zygopm/module/conf"
	"zygopm/module/msg"
)

//将缓存内的git仓库checkout 到指定的版本
func SetReference(conf *cfg.Config, resolveTest bool) error {

	if len(conf.Imports) == 0 && len(conf.DevImports) == 0 {
		msg.Info("没有需要导入的依赖")
		return nil
	}

	done := make(chan struct{}, concurrentWorkers)
	in := make(chan *cfg.Dependency, concurrentWorkers)
	var wg sync.WaitGroup
	var lock sync.Mutex
	var returnErr error

	for i := 0; i < concurrentWorkers; i++ {
		go func(ch <-chan *cfg.Dependency) {
			for {
				select {
				case dep := <-ch:

					var loc string
					if dep.Repository != "" {
						loc = dep.Repository
					} else {
						loc = "https://" + dep.Name
					}
					key, err := cache.Key(loc)
					if err != nil {
						msg.Die(err.Error())
					}
					cache.Lock(key)
					if err := VcsVersion(dep); err != nil {
						msg.Err("不能设置此版本 %s to %s: %s\n", dep.Name, dep.Reference, err)

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
		wg.Add(1)
		in <- dep

	}

	if resolveTest {
		for _, dep := range conf.DevImports {
			wg.Add(1)
			in <- dep
		}
	}

	wg.Wait()

	for i := 0; i < concurrentWorkers; i++ {
		done <- struct{}{}
	}
	// close(done)
	// close(in)

	return returnErr
}
