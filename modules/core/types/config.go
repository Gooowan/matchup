package types

type Config struct {
	Name    string   `yaml:"name"`
	Version string   `yaml:"version"`
	Modules []string `yaml:"modules"`
}