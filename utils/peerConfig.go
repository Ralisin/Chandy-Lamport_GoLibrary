package utils

import (
	"chandyLamportV2/protobuf/pb"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// Config Configuration struct to represent the JSON data
type Config struct {
	Localhost LocalhostConfig `json:"localhost"`
	Docker    DockerConfig    `json:"docker"`
}

type LocalhostConfig struct {
	ServiceRegistryAddr string `json:"ServiceRegistryAddr"`
	ServiceRegistryPort string `json:"ServiceRegistryPort"`
	PeerAddr            string `json:"PeerAddr"`
}

type DockerConfig struct {
	ServiceRegistryAddr string `json:"ServiceRegistryAddr"`
	ServiceRegistryPort string `json:"ServiceRegistryPort"`
	PeerAddr            string `json:"PeerAddr"`
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "Usage:\n"+
		"\tIf you are launching peer: go run peer.go <[-localhost/-docker]> <-readFile/-wordCount/-saver>\n"+
		"\tIf you are launching srRg: go run sr.go -sr\n")
	_, _ = fmt.Fprintf(os.Stderr, "Flags:\n"+
		"\t\t-readFile set peer as file reader\n"+
		"\t\t-wordCount set peer as word counter\n"+
		"\t\t-saver set peer as total count saver")
}

// FetchArgs return serviceRegistryAddr, serviceRegistryPort, peerAddr, peerRole
func FetchArgs() (string, string, string, pb.CountingRole) {
	var (
		serviceRegistryAddr string
		serviceRegistryPort string
		peerAddr            string

		role pb.CountingRole
	)

	var (
		docker    bool
		readFile  bool
		saver     bool
		sr        bool
		wordCount bool
	)

	flag.BoolVar(&docker, "docker", false, "Use Docker configurations")
	flag.BoolVar(&readFile, "readFile", false, "Set peer as file reader")
	flag.BoolVar(&saver, "saver", false, "Set peer as count saver")
	flag.BoolVar(&sr, "sr", false, "Use Service Registry setup")
	flag.BoolVar(&wordCount, "wordCount", false, "Set peer as word counter")
	flag.Usage = usage

	flag.Parse()

	var configFilePath string
	if sr {
		configFilePath = "../config.json"
	} else {
		configFilePath = "../config.json"
	}

	// Call the function to read app configuration
	config, err := readConfig(configFilePath)
	if err != nil {
		usage()

		os.Exit(1)
	}

	if !docker {
		serviceRegistryAddr = config.Localhost.ServiceRegistryAddr
		serviceRegistryPort = config.Localhost.ServiceRegistryPort

		peerAddr = config.Localhost.PeerAddr
	} else {
		serviceRegistryAddr = config.Docker.ServiceRegistryAddr
		serviceRegistryPort = config.Docker.ServiceRegistryPort

		peerAddr = config.Docker.PeerAddr
	}

	if !sr {
		if readFile && !wordCount && !saver {
			role = pb.CountingRole_FILE_READER
		} else if !readFile && wordCount && !saver {
			role = pb.CountingRole_WORD_COUNTER
		} else if !readFile && !wordCount && saver {
			role = pb.CountingRole_SAVER
		} else {
			usage()

			os.Exit(1)
		}
	}

	return serviceRegistryAddr, serviceRegistryPort, peerAddr, role
}

func readConfig(filename string) (Config, error) {
	var config Config

	// Read the JSON file
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("error reading JSON file: %v", err)
	}

	// Unmarshal JSON data into Config struct
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return config, nil
}
