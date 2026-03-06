// Copyright 2026 Northern.tech AS
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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSafeString(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		maxLen  int
		wantErr string
	}{
		{
			name:   "valid string",
			value:  "hello",
			maxLen: 256,
		},
		{
			name:    "exceeds max length",
			value:   strings.Repeat("a", 257),
			maxLen:  256,
			wantErr: "exceeds maximum length of 256",
		},
		{
			name:    "starts with dash",
			value:   "--malicious-flag",
			maxLen:  256,
			wantErr: "must not start with '-'",
		},
		{
			name:    "contains control character",
			value:   "hello\x00world",
			maxLen:  256,
			wantErr: "must not contain control characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSafeString("test", tt.value, tt.maxLen)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	makeArgs := func(filename, destDir string) string {
		b, _ := json.Marshal(args{Filename: filename, DestDir: destDir})
		return string(b)
	}

	makeArgsWithSoftware := func(filename, destDir, fs string) string {
		b, _ := json.Marshal(args{
			Filename:           filename,
			DestDir:            destDir,
			SoftwareFilesystem: fs,
		})
		return string(b)
	}

	validArgs := makeArgs("firmware.bin", "/opt/firmware")

	tests := []struct {
		name         string
		cmd          SingleFileCmd
		wantErr      string
		wantFileName string
	}{
		{
			name: "valid command",
			cmd: SingleFileCmd{
				Workdir:      "/var",
				ArtifactName: "my-artifact",
				DeviceTypes:  []string{"raspberrypi4"},
				Args:         validArgs,
			},
			wantFileName: "firmware.bin",
		},
		{
			name: "path traversal in filename is sanitized",
			cmd: SingleFileCmd{
				Workdir:      "/var",
				ArtifactName: "my-artifact",
				DeviceTypes:  []string{"raspberrypi4"},
				Args:         makeArgs("../../etc/shadow", "/opt/firmware"),
			},
			wantFileName: "shadow",
		},
		{
			name: "empty filename",
			cmd: SingleFileCmd{
				Workdir:      "/var",
				ArtifactName: "my-artifact",
				DeviceTypes:  []string{"raspberrypi4"},
				Args:         makeArgs("", "/opt/firmware"),
			},
			wantErr: "invalid filename",
		},
		{
			name: "empty required field",
			cmd: SingleFileCmd{
				Workdir:     "/var",
				DeviceTypes: []string{"raspberrypi4"},
				Args:        validArgs,
			},
			wantErr: "artifact-name can't be empty",
		},
		{
			name: "flag injection in device type",
			cmd: SingleFileCmd{
				Workdir:      "/var",
				ArtifactName: "my-artifact",
				DeviceTypes:  []string{"--inject"},
				Args:         validArgs,
			},
			wantErr: "device-type must not start with '-'",
		},
		{
			name: "empty device type",
			cmd: SingleFileCmd{
				Workdir:      "/var",
				ArtifactName: "my-artifact",
				DeviceTypes:  []string{""},
				Args:         validArgs,
			},
			wantErr: "device type can't be empty",
		},
		{
			name: "optional field validation",
			cmd: SingleFileCmd{
				Workdir:      "/var",
				ArtifactName: "my-artifact",
				DeviceTypes:  []string{"raspberrypi4"},
				Args:         makeArgsWithSoftware("firmware.bin", "/opt/firmware", "--inject"),
			},
			wantErr: "software_filesystem must not start with '-'",
		},
		{
			name: "path traversal in dest_dir",
			cmd: SingleFileCmd{
				Workdir:      "/var",
				ArtifactName: "my-artifact",
				DeviceTypes:  []string{"raspberrypi4"},
				Args:         makeArgs("firmware.bin", "/var/../../etc"),
			},
			wantErr: "invalid artifact destination dir",
		},
		{
			name: "invalid args json",
			cmd: SingleFileCmd{
				Workdir:      "/var",
				ArtifactName: "my-artifact",
				DeviceTypes:  []string{"raspberrypi4"},
				Args:         "not-json",
			},
			wantErr: "can't parse 'args'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				if tt.wantFileName != "" {
					assert.Equal(t, tt.wantFileName, tt.cmd.FileName)
				}
			}
		})
	}
}
