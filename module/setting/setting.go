//设置
package setting

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"zygopm/module/msg"
	"path/filepath"
)

const (
	VENDOR   = "vendor"

)

var gopmFileName = "gopm.toml"
var lockFile = "gopm.lock"
var gopaths string

func init() {
	if gopaths = os.Getenv("GOPATH"); len(gopaths) == 0 {
		out, err := exec.Command("go", "env", "GOPATH").Output()
		if err == nil {
			gopaths = strings.TrimSpace(string(out))
		}
	}
}


//获取第一个gopath
func GetFirstGOPATHSrc() string {
	s := strings.Split(gopaths, string(os.PathListSeparator))
	if len(s)==0 {
		msg.Die("gopath未设置")
	}
	return filepath.Join(s[0], "src")
}

//TODO 完善缓存目录获取
func Home() string {
	return filepath.Join(filepath.Dir(GetFirstGOPATHSrc()), "zygopm-home")

}

func IsDirectoryEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)

	if err == io.EOF {
		return true, nil
	}

	return false, err
}

func Vendor() (string, error) {

	return VENDOR, nil
}

//获取配置文件name
func GetGopmConfFileName() string {
	return gopmFileName
}

//设置配置文件name
func SetGopmConfFileName(f string) {
	gopmFileName = f+".toml"
}
//获取配置文件name
func GetGopmLockFileName() string {
	return lockFile
}

//设置配置文件name
func SetGopmLockFileName(f string) {
	lockFile = f+".lock"
}
