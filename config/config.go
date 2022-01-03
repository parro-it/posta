package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Version of the command
var Version string = "development"

// Type contains the app configuration.
// It is read from the file specified
type Type struct {
	ConfigFile string
	Showver    bool
	Accounts   []Account
}

type Account struct {
	Addr string
	User string
	Pass string
	Name string
}

// Values contains the configuration as read from the toml file.
var Values Type

// Init fill the configuration variable with content read from
// file Values.ConfigFile.
func Init() error {
	if _, err := toml.DecodeFile(Values.ConfigFile, &Values); err != nil {

		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(Values.ConfigFile), os.FileMode(0755)); err != nil {
				return err
			}

			f, err := os.OpenFile(Values.ConfigFile, os.O_WRONLY|os.O_CREATE, os.FileMode(0644))
			if err != nil {
				return err
			}
			defer f.Close()

			encoder := toml.NewEncoder(f)

			if err := encoder.Encode(Values); err != nil {
				return err
			}

			return nil
		}
		return fmt.Errorf("Cannot read config file %s: %w", Values.ConfigFile, err)
	}

	return nil
}

// ParseCommandLine ...
func ParseCommandLine() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return true
	}
	flag.StringVar(&Values.ConfigFile, "c", filepath.Join(home, ".posta/posta.cfg.toml"), "config file path")
	//flag.IntVar(&Values.Hours, "h", 3, "number of times a download is retried vefore failing")

	showver := flag.Bool("v", false, "print version to stdout and exit")

	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		return true
	}

	if showver != nil && *showver {
		fmt.Printf("posta %s\n", Version)
		return true
	}

	return false
}
