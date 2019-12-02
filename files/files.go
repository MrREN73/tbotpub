package files

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func GetPath(rawPath string) string {
	_, f := filepath.Split(rawPath)
	return f
}

func Download(url string, pathFile string) error {
	//nolint:gosec
	resp, err := http.Get(url) //potential SSRF attack https://securego.io/docs/rules/g107.html
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	file, err := os.Create(pathFile)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func ScannerChan(path string, bufSize int) (chan string, error) {
	out := make(chan string, bufSize)

	go func() {
		defer close(out)

		file, err := os.Open(path)
		if err != nil {
			log.Println(err)
			return
		}

		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			out <- scanner.Text()
		}

		if err := os.Remove(path); err != nil {
			log.Println(err)
		}
	}()

	return out, nil
}
