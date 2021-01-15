/*
Copyright 2020 The Kubernetes Authors.

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
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	srcPath = "../../config/prow/cluster/monitoring"
)

var (
	optionalConfigPaths = []string{
		"mixins/grafana_dashboards",
		"mixins/prometheus",
	}
	requiredConfigPaths = []string{
		"mixins/lib/config_util.libsonnet",
	}
	wantExtentions = sets.NewString(
		".jsonnet",
		".libsonnet",
	)
)

type options struct {
	srcPath string
	dstPath string
}

type client struct {
	srcPath string
	dstPath string
	paths   []string
}

func (c *client) addNewPath(p string) {
	if wantExtentions.Has(path.Ext(p)) {
		c.paths = append(c.paths, p)
	}
}

func (c *client) walk(rootPath, subPath string) {
	fullPath := path.Join(rootPath, subPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		log.Fatalf("failed to get the file info for %q: %v", fullPath, err)
	}
	if info.IsDir() {
		filepath.Walk(fullPath, func(leafPath string, info os.FileInfo, err error) error {
			relPath, _ := filepath.Rel(rootPath, leafPath)
			c.addNewPath(relPath)
			return nil
		})
	} else {
		c.addNewPath(subPath)
	}
}

func main() {
	o := options{}
	flag.StringVar(&o.srcPath, "src", "", "Src dir of monitoring")
	flag.StringVar(&o.dstPath, "dst", "", "Dst dir of monitoring")
	flag.Parse()

	c := client{
		srcPath: o.srcPath,
		dstPath: o.dstPath,
		paths:   make([]string, 0),
	}

	log.Print("Processing optional configs: ", c.srcPath)
	for _, subPath := range optionalConfigPaths {
		c.walk(c.dstPath, subPath)
	}

	log.Print("Processing required configs: ", c.dstPath)
	for _, subPath := range requiredConfigPaths {
		c.walk(c.srcPath, subPath)
	}

	log.Print("Files to be copied: ", c.paths)
	for _, subPath := range c.paths {
		srcPath := path.Join(c.srcPath, subPath)
		dstPath := path.Join(c.dstPath, subPath)
		content, err := ioutil.ReadFile(srcPath)
		if err != nil {
			log.Fatalf("Failed reading file %q: %v", srcPath, err)
		}
		if err := ioutil.WriteFile(dstPath, content, 0644); err != nil {
			log.Fatalf("Failed writing file %q: %v", dstPath, err)
		}
	}
}
