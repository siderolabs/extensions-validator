// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/siderolabs/talos/pkg/machinery/extensions"
	"github.com/siderolabs/talos/pkg/machinery/extensions/services"
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
	pkgName    string

	// semverRegex is a regex to match a semver version.
	semverRegex = regexp.MustCompile(`^v?(\d+\.\d+\.\d+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?)$`)
	// yyyymmddRegex is a regex to match a yyyymmdd date.
	// Eg:
	// 		20210914
	// 		20210914.1
	yyyymmddRegex = regexp.MustCompile(`^\d{8}(\.\d)?$`)
	// buildArgRegex is a regex to match a build arg version.
	// Eg:
	// 		535.129.03-v1.8.0-alpha.0-10-g336fa0f-dirty
	// 		535.129.03-v1.8.0-alpha.0-10-g336fa0f
	buildArgRegex = regexp.MustCompile(`^(\d+\.\d+\.\d+)-v(\d+\.\d+\.\d+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?)(-(\d+)-g([0-9a-f]+)(-dirty)?)?$`)
	// commitBuildArgRegex is a regex to match a commit build arg version.
	// Eg:
	// 		5815ee3-v1.8.0-alpha.0-10-g336fa0f-dirty
	// 		5815ee3-v1.8.0-alpha.0-10-g336fa0f
	commitBuildArgRegex = regexp.MustCompile(`^([0-9a-f]+)-v(\d+\.\d+\.\d+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?)(-(\d+)-g([0-9a-f]+)(-dirty)?)?$`)
	// partialSemverRegex is a regex to match a partial semver version.
	// Eg:
	// 		v4.3
	// 		4.3
	partialSemverRegex = regexp.MustCompile(`^v?(\d+\.\d+)$`)
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
	switch _ = extension.Manifest.Metadata.Version; {
	case semverRegex.MatchString(extension.Manifest.Metadata.Version):
	case yyyymmddRegex.MatchString(extension.Manifest.Metadata.Version):
	case buildArgRegex.MatchString(extension.Manifest.Metadata.Version):
	case commitBuildArgRegex.MatchString(extension.Manifest.Metadata.Version):
	case partialSemverRegex.MatchString(extension.Manifest.Metadata.Version):
	default:
		return fmt.Errorf("invalid version format %s for extension: %s", extension.Manifest.Metadata.Version, extension.Manifest.Metadata.Name)
	}

	// find all yaml files in usr/local/etc/containers
	extensionServiceSpecFiles, err := filepath.Glob(filepath.Join(extension.RootfsPath(), "usr/local/etc/containers", "*.yaml"))
	if err != nil {
		return fmt.Errorf("error finding service spec files: %w", err)
	}

	for _, extensionServiceSpecFile := range extensionServiceSpecFiles {
		var extensionServiceSpec services.Spec

		extensionServiceSpecData, err := os.ReadFile(extensionServiceSpecFile)
		if err != nil {
			return fmt.Errorf("error reading service spec file %s: %w", extensionServiceSpecFile, err)
		}

		if err := yaml.Unmarshal(extensionServiceSpecData, &extensionServiceSpec); err != nil {
			return fmt.Errorf("error unmarshalling service spec file %s: %w", filepath.Base(extensionServiceSpecFile), err)
		}

		if err := extensionServiceSpec.Validate(); err != nil {
			return fmt.Errorf("error validating service spec file %s: %w", filepath.Base(extensionServiceSpecFile), err)
		}
	}

	return extension.Validate(
		extensions.WithValidateConstraints(),
		extensions.WithValidateContents(),
	)
}
