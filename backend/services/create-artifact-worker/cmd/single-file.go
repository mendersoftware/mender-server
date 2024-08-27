// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mendersoftware/mender-server/services/create-artifact-worker/client"
	"github.com/mendersoftware/mender-server/services/create-artifact-worker/config"
	mlog "github.com/mendersoftware/mender-server/services/create-artifact-worker/log"
)

const (
	argToken          = "token"
	argArtifactName   = "artifact-name"
	argDescription    = "description"
	argDeviceType     = "device-type"
	argArtifactId     = "artifact-id"
	argGetArtifactUri = "get-artifact-uri"
	argDelArtifactUri = "delete-artifact-uri"
	argTenantId       = "tenant-id"
	argArgs           = "args"
)

type args struct {
	Filename           string `json:"filename"`
	DestDir            string `json:"dest_dir"`
	SoftwareFilesystem string `json:"software_filesystem"`
	SoftwareName       string `json:"software_name"`
	SoftwareVersion    string `json:"software_version"`
}

var singleFileCmd = &cobra.Command{
	Use:   "single-file",
	Short: "Generate an update using a single-file update module.",
	Long: "\nBesides command line args, supports the following env vars:\n\n" +
		"CREATE_ARTIFACT_SKIPVERIFY skip ssl verification (default: false)\n" +
		"CREATE_ARTIFACT_WORKDIR working dir for processing (default: /var)\n" +
		"CREATE_ARTIFACT_DEPLOYMENTS_URL internal deployments service url\n",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := NewSingleFileCmd(cmd, args)
		if err != nil {
			mlog.Error(err.Error())
			os.Exit(1)
		}

		err = c.Run()
		if err != nil {
			mlog.Error(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	singleFileCmd.Flags().String(argToken, "", "auth token")
	_ = singleFileCmd.MarkFlagRequired(argToken)

	singleFileCmd.Flags().String(argArtifactName, "", "artifact name")
	_ = singleFileCmd.MarkFlagRequired(argArtifactName)

	singleFileCmd.Flags().String(argArtifactId, "", "artifact id")
	_ = singleFileCmd.MarkFlagRequired(argArtifactId)

	singleFileCmd.Flags().String(
		argGetArtifactUri,
		"",
		"pre-signed s3 url to uploaded temp artifact (GET)",
	)
	_ = singleFileCmd.MarkFlagRequired(argGetArtifactUri)

	singleFileCmd.Flags().String(
		argDelArtifactUri,
		"",
		"pre-signed s3 url to uploaded temp artifact (DELETE)",
	)
	_ = singleFileCmd.MarkFlagRequired(argDelArtifactUri)

	singleFileCmd.Flags().String(argTenantId, "", "tenant id")
	_ = singleFileCmd.MarkFlagRequired(argTenantId)

	singleFileCmd.Flags().String(argDeviceType, "", "device type")
	_ = singleFileCmd.MarkFlagRequired(argDeviceType)

	// json string of specific args: dest dir, file name
	singleFileCmd.Flags().String(
		argArgs,
		"",
		"specific args in json form: {\"file\":<DESTINATION_FILE_NAME_ON_DEVICE>,"+
			" \"dest_dir\":<DESTINATION_DIR_ON_DEVICE>},"+
			" \"software_filesystem\":<SOFTWARE_FILESYSTEM>},"+
			" \"software_name\":<SOFTWARE_NAME>},"+
			" \"software_version\":<SOFTWARE_VERSION>}",
	)
	_ = singleFileCmd.MarkFlagRequired(argArgs)

	singleFileCmd.Flags().String(argDescription, "", "artifact description")
}

type SingleFileCmd struct {
	ServerUrl      string
	DeploymentsUrl string
	SkipVerify     bool
	Workdir        string

	ArtifactName   string
	Description    string
	DeviceTypes    []string
	ArtifactId     string
	GetArtifactUri string
	DelArtifactUri string
	Args           string
	TenantId       string
	AuthToken      string

	// type-specific args
	FileName           string
	DestDir            string
	SoftwareFilesystem string
	SoftwareName       string
	SoftwareVersion    string
}

func NewSingleFileCmd(cmd *cobra.Command, args []string) (*SingleFileCmd, error) {
	c := &SingleFileCmd{}

	if err := c.init(cmd); err != nil {
		return nil, err
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *SingleFileCmd) init(cmd *cobra.Command) error {
	c.DeploymentsUrl = viper.GetString(config.CfgDeploymentsUrl)
	c.SkipVerify = viper.GetBool(config.CfgSkipVerify)
	c.Workdir = viper.GetString(config.CfgWorkDir)

	var arg string
	arg, err := cmd.Flags().GetString(argArtifactName)
	c.ArtifactName = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argDescription)
	c.Description = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argDeviceType)
	c.DeviceTypes = strings.Split(arg, ",")
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argArtifactId)
	c.ArtifactId = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argGetArtifactUri)
	c.GetArtifactUri = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argDelArtifactUri)
	c.DelArtifactUri = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argTenantId)
	c.TenantId = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argToken)
	c.AuthToken = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argArgs)
	c.Args = arg
	if err != nil {
		return err
	}

	return nil
}

func (c *SingleFileCmd) Validate() error {
	if err := config.ValidAbsPath(c.Workdir); err != nil {
		return errors.Wrap(err, "invalid workdir")
	}

	var args args

	err := json.Unmarshal([]byte(c.Args), &args)
	if err != nil {
		return errors.Wrap(err, "can't parse 'args'")
	}

	c.FileName = args.Filename
	c.DestDir = args.DestDir
	c.SoftwareFilesystem = args.SoftwareFilesystem
	c.SoftwareName = args.SoftwareName
	c.SoftwareVersion = args.SoftwareVersion

	if c.FileName == "" {
		return errors.New("destination filename can't be empty")
	}

	if err := config.ValidAbsPath(c.DestDir); err != nil {
		return errors.Wrap(err, "invalid artifact destination dir")
	}

	return nil
}

func (c *SingleFileCmd) Run() error {
	mlog.Info("running single-file update module generation:\n%s", c.dumpArgs())
	mlog.Info("config:\n%s", config.Dump())

	cd, err := client.NewDeployments(c.DeploymentsUrl, c.SkipVerify)
	if err != nil {
		return errors.New("failed to configure 'deployments' client")
	}

	cs3 := client.NewStorage(c.SkipVerify)

	ctx := context.Background()

	mlog.Verbose("creating temp dir at", c.Workdir)

	downloadDir, err := os.MkdirTemp(c.Workdir, "single-file")
	if err != nil {
		return errors.Wrapf(err, "failed to create temp dir under workdir %s", c.Workdir)
	}

	//gotcha: must download under the correct name (destination name on the device)
	//artifact generator will not allow renaming it
	downloadFile := filepath.Join(downloadDir, c.FileName)

	mlog.Verbose("downloading temp artifact to %s", downloadFile)

	err = cs3.Download(ctx, c.GetArtifactUri, downloadFile)
	if err != nil {
		return errors.Wrapf(err, "failed to download input file at %s", c.GetArtifactUri)
	}

	// make the filename unique by naming it after the artifact
	outfile := c.ArtifactId + "-generated"
	outfile = filepath.Join(downloadDir, outfile)

	mlog.Verbose("generating output artifact %s", outfile)

	// run gen script
	args := []string{
		"-n", c.ArtifactName,
		"-d", c.DestDir,
		"-o", outfile,
	}
	if c.SoftwareFilesystem != "" {
		args = append(args, "--software-filesystem", c.SoftwareFilesystem)
	}
	if c.SoftwareName != "" {
		args = append(args, "--software-name", c.SoftwareName)
	}
	if c.SoftwareVersion != "" {
		args = append(args, "--software-version", c.SoftwareVersion)
	}

	for _, deviceType := range c.DeviceTypes {
		args = append(args, "-t", deviceType)
	}
	args = append(args, downloadFile)
	cmd := exec.Command("/usr/bin/single-file-artifact-gen", args...)

	std, err := cmd.CombinedOutput()
	mlog.Info(string(std))
	if err != nil {
		return errors.Wrapf(err, "single-file-artifact-gen exited with error %s", std)
	}

	mlog.Verbose("deleting temp file from S3")

	err = cs3.Delete(ctx, c.DelArtifactUri)
	if err != nil {
		return errors.Wrapf(err, "failed to delete artifact at %s", c.DelArtifactUri)
	}

	mlog.Verbose("uploading generated artifact")
	err = cd.UploadArtifactInternal(ctx, outfile, c.ArtifactId, c.TenantId, c.Description)
	if err != nil {
		return errors.Wrapf(err, "failed to upload generated artifact")
	}

	err = os.RemoveAll(downloadDir)
	if err != nil {
		mlog.Error("failed to remove temp working dir %s: %v", downloadDir, err.Error())
	}

	return nil
}

func (c *SingleFileCmd) dumpArgs() string {
	return dumpArg(argArtifactName, c.ArtifactName) +
		dumpArg(argDescription, c.Description) +
		dumpArg(argArtifactId, c.ArtifactId) +
		dumpArg(argDeviceType, strings.Join(c.DeviceTypes, ",")) +
		dumpArg(argTenantId, c.TenantId) +
		dumpArg(argGetArtifactUri, c.GetArtifactUri) +
		dumpArg(argDelArtifactUri, c.DelArtifactUri) +
		dumpArg(argArgs, c.Args)
}

func dumpArg(n, v string) string {
	return fmt.Sprintf("--%s: %s\n", n, v)
}
