package syaml

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadFile[T any](name string) (*T, error) {
	bs, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return Load[T](bs)
}

func Load[T any](data []byte) (*T, error) {
	var result T
	err := yaml.Unmarshal(data, &result)
	return &result, err
}
