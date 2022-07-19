package utils

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func MkDir(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		return err
	}
	return nil
}

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func UnzipSource(source string) (*string, error) {
	// 1. Open the zip file
	reader, err := zip.OpenReader(source)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	destination, err := filepath.Abs(source)
	if err != nil {
		return nil, err
	}

	path := strings.Replace(destination, ".zip", "", 1)
	err = MkDir(path)
	if err != nil {
		return nil, err
	}

	// 2. Get the absolute destination path
	// destination, err = filepath.Abs(destination)

	// 3. Iterate over zip files inside the archive and unzip each of them
	for _, f := range reader.File {
		err := UnzipFile(f, path)
		if err != nil {
			return nil, err
		}
	}

	return &path, nil
}

func UnzipFile(f *zip.File, destination string) error {
	// 4. Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// 5. Create directory tree
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// 6. Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// 7. Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}

func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func ConfigKeyValuePairDelete(key string) {
	DeleteKeyHack(key)
}

func DeleteKeyHack(key string) {
	settings := viper.AllSettings()
	delete(settings, key)

	var parsedSettings string
	for key, value := range settings {
		parsedSettings = fmt.Sprintf("%s\n%s: %s", parsedSettings, key, value)
	}

	d1 := []byte(parsedSettings)
	HandleError(ioutil.WriteFile(viper.ConfigFileUsed(), d1, 0644))
}

func ConfigKeyValuePairUpdate(key string, value string) {
	writeKeyValuePair(key, value)
}

func ConfigKeyValuePairAdd(key string, value string) {
	if validateKeyValuePair(key, value) {
		log.Printf("Validation not met for %s.", key)
	} else {
		writeKeyValuePair(key, value)
	}
}

func validateKeyValuePair(key string, value string) bool {
	if len(key) == 0 || len(value) == 0 {
		fmt.Println("The key and value must both contain contents to write to the configuration file.")
		return true
	}
	// Determine if an existing key, if so notify.
	if findExistingKey(key) {
		fmt.Println("This key already exists. Create a key value pair with a different key, or if this is an update use the update command.")
		return true
	}
	return false
}

func writeKeyValuePair(key string, value interface{}) {
	viper.Set(key, value)
	err := viper.WriteConfig()
	HandleError(err)
	fmt.Printf("Wrote the %s pair.\n", key)
}

func findExistingKey(theKey string) bool {
	existingKey := false
	for i := 0; i < len(viper.AllKeys()); i++ {
		if viper.AllKeys()[i] == theKey {
			existingKey = true
		}
	}
	return existingKey
}

func HandleError(e error) {
	if e != nil {
		fmt.Println(e)
	}
}
