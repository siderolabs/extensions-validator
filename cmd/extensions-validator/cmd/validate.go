// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/siderolabs/talos/pkg/machinery/extensions"
	"github.com/spf13/cobra"
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
	pkgName    string
)

func init() {
	validateCmd.Flags().StringVar(&rootfsPath, "rootfs", "", "Path to the rootfs")
	validateCmd.MarkFlagRequired("rootfs") //nolint:errcheck
	validateCmd.Flags().StringVar(&pkgName, "pkg-name", "", "Pkg name defined in the pkg file")
	rootCmd.AddCommand(validateCmd)
}

func validateRootfs() error {
	if rootfsPath == "" {
		return errors.New("rootfs path is required")
	}

	extension, err := extensions.Load(rootfsPath)
	if err != nil {
		return fmt.Errorf("error loading extension: %w", err)
	}

	if pkgName != "" {
		if pkgName != extension.Manifest.Metadata.Name {
			return fmt.Errorf("pkg name does not match extension name: %s != %s", pkgName, extension.Manifest.Metadata.Name)
		}
	}

	// validate extension version
	if _, err := semver.Parse(strings.TrimPrefix(extension.Manifest.Metadata.Version, "v")); err != nil {
		return fmt.Errorf("error parsing extension with version %s, : %w", extension.Manifest.Metadata.Version, err)
	}

	return extension.Validate()
}
