// Copyright 2021 Inuits
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	app        = kingpin.New("dashboard-manager", "A command-line dashboard manager.")
	configFile = app.Flag("config-file", "Path to the configuration file.").Short('c').Required().ExistingFile()

	fetch          = app.Command("fetch", "Fetch dashboards from input grafana.")
	fetchDirectory = fetch.Flag("output-directory", "Directory to fetch the dashboards to.").Required().String()

	compare          = app.Command("compare", "Compare dashboards.")
	compareDirectory = compare.Flag("dashboards-directory", "Directory where the dashboards were fetched.").Required().String()
	compareResults   = compare.Flag("results", "File to write result to.").Required().String()

	upload               = app.Command("upload", "Upload dashboards.")
	uploadDirectory      = upload.Flag("dashboards-directory", "Directory where the dashboards were fetched.").Required().String()
	uploadSource         = upload.Flag("input-instance", "Name of the output instance").Required().String()
	uploadOutput         = upload.Flag("output-instance", "Name of the output instance").Required().String()
	uploadDashboardsList = upload.Flag("dashboards", "Dashboards to upload").Required().Strings()

	snapshot               = app.Command("snapshot", "Upload snapshots.")
	snapshotDirectory      = snapshot.Flag("dashboards-directory", "Directory where the dashboards were fetched.").Required().String()
	snapshotSource         = snapshot.Flag("input-instance", "Name of the output instance").Required().String()
	snapshotOutput         = snapshot.Flag("output-instance", "Name of the output instance").Required().String()
	snapshotDashboardsList = snapshot.Flag("dashboards", "Dashboards to snapshot").Required().Strings()
	snapshotExpire         = snapshot.Flag("expire", "Expiration time").Default("1h").Duration()
)

type gitConfig struct {
	URL    string `yaml:"url"`
	Branch string `yaml:"branch"`
}

type config struct {
	Git    gitConfig         `yaml:"git_config"`
	Input  []grafanaInstance `yaml:"grafana_instances_input"`
	Output []grafanaInstance `yaml:"grafana_instances_output"`
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case fetch.FullCommand():
		cfg, err := loadConfig(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		err = fetchDashboards(cfg)
		if err != nil {
			log.Fatal(err)
		}
	case compare.FullCommand():
		cfg, err := loadConfig(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		err = compareDashboards(cfg)
		if err != nil {
			log.Fatal(err)
		}
	case upload.FullCommand():
		cfg, err := loadConfig(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		err = uploadDashboards(cfg)
		if err != nil {
			log.Fatal(err)
		}
	case snapshot.FullCommand():
		cfg, err := loadConfig(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		err = snapshotDashboards(cfg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func loadConfig(configFile string) (*config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	cfg := &config{}
	err = yaml.Unmarshal(data, cfg)
	return cfg, err
}
