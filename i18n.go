// Tomato static website generator
// Copyright Quentin Ribac, 2018
// Free software license can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
)

// LoadLocales loads
func LoadLocales(localesDir string) *i18n.I18n {
	paths, err := filepath.Glob(filepath.Join(localesDir, "*.yml"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: bad pattern in localesDir")
		os.Exit(1)
	}

	return i18n.New(yaml.New(paths...))
}
