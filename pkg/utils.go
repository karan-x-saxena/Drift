package pkg

import (
	"errors"
	"fmt"
	"os"
)

func IsFileExist(file string) error {
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("%s file does not exist", file)
	}
	if info.IsDir() {
		return errors.New("provided a dir not file")
	}
	return nil
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
