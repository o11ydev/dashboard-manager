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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func lazyMkdir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.Mkdir(path, os.ModePerm)
	}
	return err
}

func fetchDashboards(cfg *config) error {
	err := lazyMkdir(*fetchDirectory)
	if err != nil {
		return fmt.Errorf("error making base directory: %w", err)
	}

	for _, instance := range cfg.Input {
		client, err := instance.client()
		if err != nil {
			return err
		}
		basepath := filepath.Join(*fetchDirectory, instance.Name)
		err = lazyMkdir(basepath)
		if err != nil {
			return fmt.Errorf("error making directory for %s: %w", instance.Name, err)
		}

		dashboards, err := client.Search(context.TODO())
		if err != nil {
			return err
		}
		for _, d := range dashboards {
			folder := filepath.Join(basepath, d.FolderUID)
			err := lazyMkdir(folder)
			if err != nil {
				return fmt.Errorf("error making directory for %s / %s: %w", instance.Name, d.FolderUID, err)
			}
			filePath := filepath.Join(folder, d.UID+".json")
			board, _, err := client.GetDashboardByUID(context.TODO(), d.UID)
			if err != nil {
				return err
			}
			data, err := json.MarshalIndent(board, "", " ")
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(filePath, data, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
