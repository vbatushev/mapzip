package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	version      = "1.0"
	appVersion   = "mapzip " + version
	startPath    = "./"
	prefixFolder = ""
	targetPath   string
)

func init() {
	log.SetOutput(os.Stdout)
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
	absPath, _ := filepath.Abs(path.Join(targetPath, fldName))
	zfs := getZipFilesSlice(absPath)
	err := zipFiles(pathZip, absPath, zfs)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func getZipFilesSlice(absPath string) (result []string) {
	result = make([]string, 0)
	filepath.Walk(absPath, func(p string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			if checkPath(f, p) {
				result = append(result, p)
			} else {
				fmt.Println("Exclude", p)
			}
		}
		return nil
	})
	return result
}

func checkPath(f os.FileInfo, p string) bool {
	if filepath.Ext(f.Name()) == ".zip" {
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

func zipFiles(filename string, fld string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(file, fld)

		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
		}
		zipfile.Close()
	}
	return nil
}
