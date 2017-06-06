/*
Copyright 2017 Caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/urfave/cli.v1"

	log "github.com/zoumo/logdog"
)

var sentinels = []string{
	"Copyright",
	"Caicloud",
	`Licensed under the Apache License, Version 2.0 (the "License");`,
}

func Run(c *cli.Context) {
	root := c.Args().First()

	if root == "" {
		root = "./"
	}

	licenseBytes, err := ioutil.ReadFile(root + "/LICENSE")
	if err != nil {
		log.Fatal(err)
		return
	}

	license := []byte(fmt.Sprintf("/*\n%s*/\n\n", licenseBytes))

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip vendor
		if info.IsDir() &&
			(strings.Contains(path, "vendor") || strings.Contains(path, ".git")) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// skip not go file
		if ext := filepath.Ext(path); ext != ".go" {
			// log.Infof("Skip file: %s", path)
			return nil
		}

		allFile, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}

		src := allFile[:150]

		needLicense := false

		for _, sentinel := range sentinels {
			if !bytes.Contains(src, []byte(sentinel)) {
				needLicense = true
			}
		}

		if needLicense {
			log.Infof("Add License to file: %s", path)

			i := bytes.Index(allFile, []byte("package"))

			if !c.Bool("dryRun") {
				ioutil.WriteFile(path, append(license, allFile[i:]...), 0655)
			}
			return nil
		}

		log.Infof("Skip file: %s", path)

		return nil
	})

	if err != nil {
		log.Error(err)
	}
}

func main() {

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "dryRun",
		},
	}
	app.Action = func(c *cli.Context) error {
		Run(c)
		return nil
	}
	app.Run(os.Args)

}
