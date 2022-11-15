package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	netUrl "net/url"
	"os"
	"path/filepath"
)

func GetFilePathType(FilePath string) int {
	file, err := os.Open(FilePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0
	}

	if fileInfo.IsDir() {
		return 1
	} else {
		return -1
	}
}

func DetermineExtension(fileType string) string {
	log.Printf("FileMime")
	ext, err := mime.ExtensionsByType(fileType)
	if err != nil || len(ext) == 0 {
		return ""
	}
	return ext[0]
}

func main() {
	var PathToInputFile string
	var PathToOutputDir string

	flag.StringVar(&PathToInputFile, "file", "", "Path to file with url links")
	flag.StringVar(&PathToOutputDir, "dst", ".", "Path to output directory with url links content")
	flag.Parse()

	if PathToInputFile == "" {
		log.Fatalln("Flag -file is required")
	}
	if PathToOutputDir == "" {
		log.Fatalln("Flag -dst is required")
	}

	if GetFilePathType(PathToInputFile) != -1 {
		log.Fatalf("%s is not file", PathToInputFile)
	}

	if GetFilePathType(PathToOutputDir) != 1 {
		log.Fatalf("%s is not directory", PathToOutputDir)
	}

	file, err := os.OpenFile(PathToInputFile, os.O_RDONLY, os.ModePerm)

	if err != nil {
		log.Fatalf("%s is not file", PathToInputFile)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	FileIndex := 1
	for sc.Scan() {
		url := sc.Text()
		_, err := netUrl.ParseRequestURI(url)
		if err != nil {
			log.Panicf("%s is not valid url", url)
			continue
		}
		resp, err := http.Get(url)
		if err != nil {
			log.Panicf("Error on get url %s", url)
			log.Panicln(err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Panicf("%s with status code %d", url, resp.StatusCode)
		}
		defer resp.Body.Close()

		FileExtension := DetermineExtension(resp.Header.Get("Content-Type"))
		FilePath := fmt.Sprintf("%d%s", FileIndex, FileExtension)
		OutputContentFilePath := filepath.Join(PathToOutputDir, FilePath)
		OutputContentFile, err := os.Create(OutputContentFilePath)

		FileIndex++

		if err != nil {
			log.Printf("Can`t write file: %s", OutputContentFilePath)
			continue
		}
		defer OutputContentFile.Close()
		io.Copy(OutputContentFile, resp.Body)
		log.Printf("Downloaded file: %s", OutputContentFilePath)
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("Read file error: %v", err)
		return
	}
}
