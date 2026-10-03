package data

import (
	"os"
	"path/filepath"
	"reelens/utils/system"
)

const (
	changelogsDirName = "changelogs"
)

type Changelogs struct{}

func mkChangelogsDir() error {
	dataDirPath, err := getDataDirPath()
	if err != nil {
		return err
	}

	path := filepath.Join(dataDirPath, changelogsDirName)

	return os.MkdirAll(path, dirPerm)
}

func (Changelogs) Read(changelogName string) ([]byte, error) {
	dataDirPath, err := getDataDirPath()
	if err != nil {
		return nil, err
	}

	changelogPath := filepath.Join(dataDirPath, changelogsDirName, changelogName)
	file, err := system.ReadFile(changelogPath)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (Changelogs) Write(changelogName string, content []byte) error {
	relPath := filepath.Join(changelogsDirName, changelogName)

	return mkFile(relPath, content)
}
