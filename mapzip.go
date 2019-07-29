package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
)

var (
	version      = "1.2"
	appVersion   = "mapzip " + version
	startPath    = "./"
	prefixFolder = ""
	targetPath   string
)

func init() {
	version := flag.Bool("v", false, "version")
	vers := flag.Bool("version", false, "version")
	flag.StringVar(&prefixFolder, "prefix", "", "Префикс искомой папки")
	flag.StringVar(&startPath, "start", "./", "Стартовый путь")
	flag.Parse()

	if *version || *vers {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if strings.TrimSpace(prefixFolder) == "" {
		log.Fatal("Не указан префикс искомой папки")
	}
}

func main() {
	fmt.Println(appVersion)

	targetPath = filepath.Join(startPath, "target")

	files, err := ioutil.ReadDir(targetPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, f := range files {
		if f.IsDir() && strings.HasPrefix(f.Name(), prefixFolder) {
			zipFolder(f.Name())
		}
	}
}

func zipFolder(fldName string) error {
	pathZip := filepath.Join(startPath, fldName+".zip")
	if _, err := os.Stat(pathZip); !os.IsNotExist(err) {
		err = os.Remove(pathZip)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	err := compress(path.Join(targetPath, fldName), pathZip)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func compress(rootPath string, pathZip string) error {
	absPath, _ := filepath.Abs(rootPath)
	ff, err := ioutil.ReadDir(absPath)
	if err == nil {
		zfs := make([]string, 0)
		for _, zf := range ff {
			fname := path.Join(absPath, zf.Name())
			if checkPath(zf, fname) {
				zfs = append(zfs, fname)
			}
		}
		zerr := archiver.Archive(zfs, pathZip)
		if zerr != nil {
			return zerr
		}
	}
	return nil
}

func checkPath(f os.FileInfo, p string) bool {
	if filepath.Ext(f.Name()) == ".zip" {
		return false
	}
	if f.IsDir() && strings.HasPrefix(f.Name(), "data") {
		return false
	}
	if strings.Contains(p, "META-INF") {
		return false
	}
	if strings.Contains(p, "WEB-INF") {
		return false
	}
	return true
}
