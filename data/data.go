package data

import (
	"fmt"
	"os"
	"path/filepath"
	"reelens/appinfo"
	"reelens/utils/system"
)

const (
	dirName = appinfo.Name
	dirPerm = 0o755
)

func init() {
	err := mkDataDir()
	if err != nil {
		fmt.Println(err)
	}

	err = mkChangelogsDir()
	if err != nil {
		fmt.Println(err)
	}
}

func getDataDirPath() (string, error) {
	base, err := system.UserStateDir()

	if err != nil {
		return "", err
	}

	dataDirPath := filepath.Join(base, dirName)
	return dataDirPath, nil
}

func mkDataDir() error {
	dataDirPath, err := getDataDirPath()
	if err != nil {
		return err
	}

	return os.MkdirAll(dataDirPath, dirPerm)
}

func mkFile(relPath string, content []byte) error {
	dataDirPath, err := getDataDirPath()
	if err != nil {
		return err
	}

	path := filepath.Join(dataDirPath, relPath)
	return system.WriteFile(path, content)
}
