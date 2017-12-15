package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
	"time"
	"zygopm/module/msg"
	"zygopm/module/setting"

	"crypto/md5"
	"github.com/Masterminds/vcs"
	"io"
)

type Dependencies []*Dependency

// 顶级配置文件
type Config struct {
	Path       string       `toml:"path"`
	Name       string       `toml:"package"`
	Imports    Dependencies `toml:"import"`
	DevImports Dependencies `toml:"testImport"`
}

type Dependency struct {
	Name        string   `toml:"package"`
	Reference   string   `toml:"version"`
	Pin         string   `toml:"-"`
	Repository  string   `toml:"repo"`
	VcsType     string   `toml:"vcs"`
	Subpackages []string `toml:"subpackages"`
	Arch        []string `toml:"arch"`
	Os          []string `toml:"os"`
}

//第二层的锁文件结构
type Lock struct {
	Name        string   `toml:"package"`
	Version     string   `toml:"version"`
	Repository  string   `toml:"repo,omitempty"`
	VcsType     string   `toml:"vcs,omitempty"`
	Subpackages []string `toml:"subpackages,omitempty"`
	Arch        []string `toml:"arch,omitempty"`
	Os          []string `toml:"os,omitempty"`
}

type Locks []*Lock

type Lockfile struct {
	Hash       string    `toml:"hash"`
	Updated    time.Time `toml:"updated"`
	Imports    Locks     `toml:"import"`
	DevImports Locks     `toml:"testImports"`
}

//返回一个锁文件结构
func NewLockfile(ds Dependencies, hash string) (*Lockfile, error) {
	lf := &Lockfile{
		Hash:       hash,
		Updated:    time.Now(),
		Imports:    make([]*Lock, len(ds)),
	}

	for i := 0; i < len(ds); i++ {
		lf.Imports[i] = LockFromDependency(ds[i])
	}

	return lf, nil
}

//转换
func LockFromDependency(dep *Dependency) *Lock {
	return &Lock{
		Name:        dep.Name,
		Version:     dep.Pin,
		Repository:  dep.Repository,
		VcsType:     dep.VcsType,
		Subpackages: dep.Subpackages,
		Arch:        dep.Arch,
		Os:          dep.Os,
	}
}

func DependencyFromLock(lock *Lock) *Dependency {
	return &Dependency{
		Name:        lock.Name,
		Reference:   lock.Version,
		Repository:  lock.Repository,
		VcsType:     lock.VcsType,
		Subpackages: lock.Subpackages,
		Arch:        lock.Arch,
		Os:          lock.Os,
	}
}

func GetInstallPath() string {
	c, err := ConfigFromToml()
	if err != nil || c.Path == "" {
		return setting.GetFirstGOPATH()
	}
	return c.Path
}

//读取锁文件
func ReadLockFile() (*Lockfile, error) {
	path, err := getTomlFilePath()
	if err != nil {
		msg.Die("锁配置文件没找到")
	}

	lockpath := filepath.Join(path, setting.GetGopmLockFIleName())
	lock := &Lockfile{}
	_, error := toml.DecodeFile(lockpath, &lock)

	return lock, error
}

//写锁文件
func (lf *Lockfile) WriteFile() error {
	path, err := getTomlFilePath()
	if err != nil {
		msg.Die("请先生成配置文件")
	}

	lockpath := filepath.Join(path, setting.GetGopmLockFIleName())
	f, err := os.Create(lockpath)
	if err != nil {
		return err
	}
	defer f.Close()
	e := toml.NewEncoder(f)

	if err = e.Encode(lf); err != nil {
		return err
	}
	return nil
}

//hash 配置文件
func Hash() (string, error) {
	path, err := getTomlFilePath()
	if err != nil {
		msg.Die("配置文件没找到")
	}
	p := filepath.Join(path, setting.GetGopmConfFIleName())
	file, err := os.Open(p)
	defer file.Close()
	if err != nil {
		msg.Die("无法加载配置文件")
	}
	md5h := md5.New()
	io.Copy(md5h, file)

	s := fmt.Sprintf("%x", md5h.Sum([]byte("")))
	return s, nil
}

//是否有锁文件
func HasLock() bool {
	basepath, err := getTomlFilePath()
	if err != nil {
		msg.Die("配置文件没找到")
	}
	_, e := os.Stat(filepath.Join(basepath, setting.GetGopmLockFIleName()))
	return e == nil
}

//读取配置文件
func ConfigFromToml() (*Config, error) {
	path, err := getTomlFilePath()
	if err != nil {
		msg.Die("配置文件没找到")
	}

	configPath := filepath.Join(path, setting.GetGopmConfFIleName())
	cfg := &Config{}
	_, error := toml.DecodeFile(configPath, &cfg)

	return cfg, error
}

//拿到配置文件目录
func getTomlFilePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	tomldir, err := ZYgopmWD(cwd)
	if err != nil {
		return cwd, err
	}

	return tomldir, nil
}

//获取gopmfile目录：默认有gopmfile 文件的目录就是项目
func ZYgopmWD(dir string) (string, error) {
	confpath := filepath.Join(dir, setting.GetGopmConfFIleName())
	lockpath := filepath.Join(dir, setting.GetGopmLockFIleName())

	if _, err := os.Stat(confpath); err == nil {
		return dir, nil
	}
	if _, err := os.Stat(lockpath); err == nil {
		return dir, nil
	}

	base := filepath.Dir(dir)
	if base == dir {
		return "", fmt.Errorf("不能找到配置文化在此目录下 %s", base)
	}

	return ZYgopmWD(base)
}

//写入配置文件
func (c *Config) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	e := toml.NewEncoder(f)
	if err = e.Encode(c); err != nil {
		return err
	}
	return nil
}

//获取
func (d Dependencies) Get(name string) *Dependency {
	for _, dep := range d {
		if dep.Name == name {
			return dep
		}
	}
	return nil
}

//子包判断
func (d *Dependency) HasSubpackage(sub string) bool {

	for _, v := range d.Subpackages {
		if sub == v {
			return true
		}
	}

	return false
}

func (d Dependencies) Has(name string) bool {
	for _, dep := range d {
		if dep.Name == name {
			return true
		}
	}
	return false
}

//获取依赖包远程地址
func (d *Dependency) Remote() string {
	var r string

	if d.Repository != "" {
		r = d.Repository
	} else {
		r = "https://" + d.Name
	}

	return r
}

//获取仓库对象
func (d *Dependency) GetRepo(dest string) (vcs.Repo, error) {
	remote := d.Remote()
	VcsType := d.VcsType

	//默认git
	if d.VcsType == "" {
		VcsType = "git"
	}

	if len(VcsType) > 0 && VcsType != "None" {
		switch vcs.Type(VcsType) {
		case vcs.Git:
			return vcs.NewGitRepo(remote, dest)
		case vcs.Svn:
			return vcs.NewSvnRepo(remote, dest)
		case vcs.Hg:
			return vcs.NewHgRepo(remote, dest)
		case vcs.Bzr:
			return vcs.NewBzrRepo(remote, dest)
		default:
			return nil, fmt.Errorf("不支持这种类型%s", VcsType)
		}
	}
	return vcs.NewRepo(remote, dest)
}
