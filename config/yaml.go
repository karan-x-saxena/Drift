package config

import (
	"drift/pkg"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type YamlConfig struct {
	config   Config
	yamlFile string
}

func (y *YamlConfig) generateYamlFile() error {
	y.config.addDefaults()
	file, err := yaml.Marshal(y.config)
	if err != nil {
		return err
	}

	f, err := os.Create("drift.yaml")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Writer.Write(f, file)
	if err != nil {
		return err
	}

	return nil

}

func (y *YamlConfig) parse() (Config, error) {
	err := pkg.IsFileExist(y.yamlFile)
	if err != nil {
		return y.config, err
	}

	file, err := os.ReadFile(y.yamlFile)
	if err != nil {
		return y.config, err
	}

	err = yaml.Unmarshal(file, &y.config)
	if err != nil {
		return y.config, err
	}

	y.config.addDefaults()

	err = y.config.validate()
	if err != nil {
		return y.config, err
	}

	return y.config, nil

}

func BaseYamlFile() error {
	yamlConfig := YamlConfig{}
	err := yamlConfig.generateYamlFile()
	return err
}

func NewYamlConfig(fileName string) (Config, error) {
	yamlConfig := YamlConfig{
		yamlFile: fileName,
	}

	config, err := yamlConfig.parse()

	return config, err
}
