package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
)

var (
	version      = "1.5.0"
	appVersion   = "mapzip " + version
	startPath    = "./"
	prefixFolder = ""
	targetPath   string
	// commonFolder = "D:\\vital\\Documents\\Drofa\\Projects\\drofa.map\\Карты\\common"
	commonFolder string
	pom          PomStruct
)

// PomStruct - Структура pom.xml
type PomStruct struct {
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// type ProjectStruct struct {
// 	GroupID string `xml:"groupId"`
// 	Version string `xml:"version"`
// }

func init() {
	version := flag.Bool("v", false, "version")
	vers := flag.Bool("version", false, "version")
	// flag.StringVar(&prefixFolder, "prefix", "", "Префикс искомой папки")
	flag.StringVar(&startPath, "start", "./", "Стартовый путь")
	flag.StringVar(&commonFolder, "target", commonFolder, "Путь для копирования")
	flag.Parse()

	if *version || *vers {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if commonFolder == "" {
		commonFolder = filepath.Join(startPath, "dist")
	}

	if err := getPOM(); err != nil {
		log.Fatal(err)
	}

	prefixFolder = pom.ArtifactID + "-"

	// if strings.TrimSpace(prefixFolder) == "" {
	// 	log.Fatal("Не указан префикс искомой папки")
	// }
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
			copyFolder(f.Name())
		}
	}

	fmt.Println("Копирование плеера закончено.")
}

func getPOM() error {
	xmlFile := filepath.Join(startPath, "pom.xml")
	byteValue, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(byteValue, &pom)
	return err
}

func copyFolder(fldName string) error {
	err := removeFolderContents(commonFolder)
	if err != nil {
		sourceinfo, err := os.Stat(filepath.Dir(commonFolder))
		if err != nil {
			return err
		}
		err = os.MkdirAll(commonFolder, sourceinfo.Mode())
		if err != nil {
			return err
		}
	}

	// versionEngineValue := strings.TrimPrefix(fldName, prefixFolder)
	err = ioutil.WriteFile(filepath.Join(commonFolder, "version.txt"), []byte(pom.Version), 0644)
	if err != nil {
		fmt.Println("Error", err)
	}

	absPath, _ := filepath.Abs(path.Join(targetPath, fldName))
	ff, err := ioutil.ReadDir(absPath)
	if err == nil {
		for _, f := range ff {
			fname := path.Join(absPath, f.Name())
			if checkPath(f, fname) {
				if f.IsDir() {
					copyDir(fname, filepath.Join(commonFolder, f.Name()))
				} else {
					copyFile(fname, filepath.Join(commonFolder, f.Name()))
				}
			}
		}
	}
	return nil
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

	if f.IsDir() {
		if f.Name() == "images" || f.Name() == "map" {
			return true
		}
		return false
	}
	return true
}

// Копирование файла
func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}
	return
}

// Копирование директории
func copyDir(source string, dest string) (err error) {
	// Получаем сведения об исходной директории
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	// Создаем новую директорию
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}
	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)
	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()
		if obj.IsDir() {
			// Создаем рекурсивно папки в глубину
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				log.Println(err)
			}
		} else {
			// Копируем файлы
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return
}

// Удаление всех файлов в папке
func removeFolderContents(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil
}
