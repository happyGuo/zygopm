package path

import (
	"io"
	"os"
	"path/filepath"
	"github.com/mitchellh/go-homedir"
)


var homeDir = ""
func Home() string {
	if homeDir != "" {
		return homeDir
	}

	if h, err := homedir.Dir(); err == nil {
		homeDir = filepath.Join(h, ".gopm")
	} else {
		cwd, err := os.Getwd()
		if err == nil {
			homeDir = filepath.Join(cwd, ".gopm")
		} else {
			homeDir = ".gopm"
		}
	}

	return homeDir
}




// 当前目录
func Basepath() string {
	base, err := os.Getwd()
	if err != nil {
		return "."
	}
	return base
}


// 复制dir
func CopyDir(source string, dest string) error {

	// get properties of source dir
	si, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, si.Mode())
	if err != nil {
		return err
	}

	d, err := os.Open(source)
	if err != nil {
		return err
	}
	defer d.Close()

	objects, err := d.Readdir(-1)

	for _, obj := range objects {

		sp := filepath.Join(source, "/", obj.Name())

		dp := filepath.Join(dest, "/", obj.Name())

		if obj.IsDir() {
			err = CopyDir(sp, dp)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(sp, dp)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

//复制文件
func CopyFile(source string, dest string) error {
	ln, err := os.Readlink(source)
	if err == nil {
		return os.Symlink(ln, dest)
	}
	s, err := os.Open(source)
	if err != nil {
		return err
	}

	defer s.Close()

	d, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}

	si, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.Chmod(dest, si.Mode())

	return err
}

//todo移动到安装目录
func MoveToInstallDir(o, n string) error {
	err := CopyDir(o, n)
	return err
}
