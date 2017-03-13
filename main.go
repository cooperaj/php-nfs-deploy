package main

import (
	"errors"
	"io/ioutil"

	"os"
	"path/filepath"

	"github.com/juju/loggo"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

var (
	configFilename string
	source         string
	destination    string
	debug          bool
	logger         loggo.Logger
)

type conf struct {
	Shared []string `yaml:"shared"`
}

func main() {
	logger = loggo.GetLogger("")

	app := cli.NewApp()
	app.Name = "deploy"
	app.Version = "0.9.1"
	app.Usage = "Sets up modern PHP apps to work better when using docker"
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} source destination

COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Load configuration from `FILE`",
			Value:       ".ddply",
			Destination: &configFilename,
			EnvVar:      "DEPLOY_CONFIG_FILE",
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "Increase verbosity of running messages",
			Destination: &debug,
		},
	}

	app.Action = func(c *cli.Context) (err error) {
		if c.NArg() < 2 {
			return cli.NewExitError("Source and/or destination not specified", 1)
		}

		source = c.Args().Get(0)
		if !IsDir(source) {
			return cli.NewExitError("Source argument does not point to valid directory", 2)
		}

		destination = c.Args().Get(1)

		if c.IsSet("debug") {
			logger.SetLogLevel(loggo.DEBUG)
		}

		var configPath = filepath.Join(source, configFilename)
		if c.IsSet("config") {
			configPath = configFilename
		}

		foundConfig := findConfig(configPath)
		if c.IsSet("config") && !foundConfig {
			return cli.NewExitError("Specified config file not found", 4)
		}

		yamlFile, err := ioutil.ReadFile(configPath)
		if err != nil {
			// no config file. assume link only
			logger.Infof("No configuration file found or specified. Continuing with linked deploy")

			var location = []string{""}
			LinkShared(location, source, destination)
		} else {
			// config file. assume copy and link shared
			var config conf
			err := config.getConfig(yamlFile)
			if err != nil {
				return cli.NewExitError(err, 6)
			}

			// was able to find and read config file
			logger.Infof("Shared locations from config: %s\n", config.Shared)

			logger.Infof("Copying directories...")
			CopyDir(source, destination)
			LinkShared(config.Shared, source, destination)
		}

		return nil
	}

	app.Run(os.Args)
}

func findConfig(path string) (foundConfig bool) {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}

	return file.Mode().IsRegular()
}

func (c *conf) getConfig(yamlFile []byte) (err error) {
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return errors.New(err.Error())
	}

	return nil
}
