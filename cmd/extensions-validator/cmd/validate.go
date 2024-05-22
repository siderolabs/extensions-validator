// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/siderolabs/talos/pkg/machinery/extensions"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the extensions rootfs",
	Long:  `Usage: extensions-validator validate`,
	// define a rootfs path argument
	Args: cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return validateRootfs()
	},
}

var (
	rootfsPath string
	pkgFile    string
)

func init() {
	validateCmd.Flags().StringVar(&rootfsPath, "rootfs", "", "Path to the rootfs")
	validateCmd.MarkFlagRequired("rootfs") //nolint:errcheck
	validateCmd.Flags().StringVar(&pkgFile, "pkg-file", "", "Path to the pkg.yaml file")
	rootCmd.AddCommand(validateCmd)
}

// PartialPkgFile represents a partial package file
// we only care about the name field.
type PartialPkgFile struct {
	Name string `yaml:"name"`
}

func validateRootfs() error {
	if rootfsPath == "" {
		return errors.New("rootfs path is required")
	}

	extension, err := extensions.Load(rootfsPath)
	if err != nil {
		return fmt.Errorf("error loading extension: %w", err)
	}

	if pkgFile != "" {
		// load the pkg file
		pkgFileData, err := os.ReadFile(pkgFile)
		if err != nil {
			return fmt.Errorf("error loading pkg file: %w", err)
		}

		var pkg PartialPkgFile

		// unmarshal the pkg file
		if err := yaml.Unmarshal(pkgFileData, &pkg); err != nil {
			return fmt.Errorf("error unmarshalling pkg file: %w", err)
		}

		if pkg.Name != extension.Manifest.Metadata.Name {
			return fmt.Errorf("pkg name does not match extension name: %s != %s", pkg.Name, extension.Manifest.Metadata.Name)
		}
	}

	// validate extension version
	if _, err := semver.Parse(strings.TrimPrefix(extension.Manifest.Metadata.Version, "v")); err != nil {
		return fmt.Errorf("error parsing extension with version %s, : %w", extension.Manifest.Metadata.Version, err)
	}

	return extension.Validate()
}
