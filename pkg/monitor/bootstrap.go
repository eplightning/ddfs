package monitor

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type BootstrapNode struct {
	Name    string
	Address string
}

type BootstrapVolume struct {
	Name   string
	Shards int
}

type BootstrapSettings struct {
	ShardSize    int64 `json:"shardSize" yaml:"shardSize"`
	MinFillSize  int32 `json:"minFillSize" yaml:"minFillSize"`
	Chunker      string
	RabinMaxSize int32  `json:"rabinMaxSize" yaml:"rabinMaxSize"`
	RabinMinSize int32  `json:"rabinMinSize" yaml:"rabinMinSize"`
	RabinPoly    uint64 `json:"rabinPoly" yaml:"rabinPoly"`
}

type BootstrapData struct {
	Settings BootstrapSettings
	Blocks   []BootstrapNode
	Indexes  []BootstrapNode
	Volumes  []BootstrapVolume
}

func DefaultBootstrapData() BootstrapData {
	return BootstrapData{
		Settings: BootstrapSettings{
			ShardSize:    256 * 1024 * 1024 * 1024,
			MinFillSize:  1024,
			Chunker:      "rabin",
			RabinMaxSize: 4 * 1024 * 1024,
			RabinMinSize: 100 * 1024,
			RabinPoly:    12313278162312893,
		},
		Blocks: []BootstrapNode{
			BootstrapNode{
				Address: "localhost:7301",
				Name:    "first",
			},
		},
		Indexes: []BootstrapNode{
			BootstrapNode{
				Address: "localhost:7302",
				Name:    "first",
			},
		},
	}
}

func LoadBootstrap(data []byte) (BootstrapData, error) {
	output := BootstrapData{}
	err := yaml.Unmarshal(data, &output)
	return output, err
}

func LoadBootstrapFromFile(name string) (BootstrapData, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return BootstrapData{}, err
	}
	return LoadBootstrap(data)
}
