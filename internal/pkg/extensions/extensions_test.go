// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package extensions_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/siderolabs/talos/internal/pkg/extensions"
	"github.com/siderolabs/talos/pkg/machinery/version"
)

func TestLoadValidate(t *testing.T) {
	ext, err := extensions.Load("testdata/good/extension1")
	require.NoError(t, err)

	assert.Equal(t, "gvisor", ext.Manifest.Metadata.Name)

	// override Talos version to make it predictable
	oldVersion := version.Tag
	version.Tag = "v1.0.0"

	t.Cleanup(func() {
		version.Tag = oldVersion
	})

	assert.NoError(t, ext.Validate())
}

func TestCompress(t *testing.T) {
	// Compress is going to change contents of the extension, copy to some temporary directory
	extDir := t.TempDir()

	require.NoError(t, exec.Command("cp", "-r", "testdata/good/extension1", extDir).Run())

	ext, err := extensions.Load(filepath.Join(extDir, "extension1"))
	require.NoError(t, err)

	squashDest, initramfsDest := t.TempDir(), t.TempDir()
	squashFile, err := ext.Compress(squashDest, initramfsDest)
	assert.NoError(t, err)

	assert.FileExists(t, squashFile)
	assert.FileExists(t, filepath.Join(initramfsDest, "lib", "firmware", "amd", "cpu"))
}

func TestValidateFailures(t *testing.T) {
	// override Talos version to make it predictable
	oldVersion := version.Tag
	version.Tag = "v1.0.0"

	t.Cleanup(func() {
		version.Tag = oldVersion
	})

	for _, tt := range []struct {
		name          string
		loadError     string
		validateError string
	}{
		{
			name:      "wrongfiles",
			loadError: "unexpected file \"a\"",
		},
		{
			name:      "emptymanifest",
			loadError: "unsupported manifest version: \"\"",
		},
		{
			name:      "norootfs",
			loadError: "extension rootfs is missing",
		},
		{
			name:          "badpaths",
			validateError: "path \"/boot/vmlinuz\" is not allowed in extensions",
		},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ext, err := extensions.Load(filepath.Join("testdata/bad", tt.name))

			if tt.loadError == "" {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.loadError)
			}

			if err == nil {
				err = ext.Validate()
				assert.EqualError(t, err, tt.validateError)
			}
		})
	}
}
