//nolint:all // This is all test code
package common

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/mendersoftware/mender-artifact/artifact"
	"github.com/mendersoftware/mender-artifact/awriter"
	"github.com/mendersoftware/mender-artifact/handlers"
)

// ArtifactOption modifies the WriteArtifactArgs an artifact is built from.
type ArtifactOption func(args *awriter.WriteArtifactArgs) error

func WithCompatibleDevices(compatibleDevices []string) ArtifactOption {
	return func(args *awriter.WriteArtifactArgs) error {
		args.Depends.CompatibleDevices = compatibleDevices
		return nil
	}
}

func WithUpdates(updates []handlers.Composer) ArtifactOption {
	return func(args *awriter.WriteArtifactArgs) error {
		args.Updates.Updates = updates
		return nil
	}
}

func WithModuleImage(module *handlers.ModuleImage) ArtifactOption {
	return func(args *awriter.WriteArtifactArgs) error {
		args.TypeInfoV3.Type = module.GetUpdateType()
		args.Updates.Updates = []handlers.Composer{module}
		return nil
	}
}

// WithDependsProvides sets the type-info depends/provides maps, used by
// the artifact selection semantics tests.
func WithDependsProvides(depends, provides map[string]string) ArtifactOption {
	return func(args *awriter.WriteArtifactArgs) error {
		if depends != nil {
			deps := make(artifact.TypeInfoDepends, len(depends))
			for k, v := range depends {
				deps[k] = v
			}
			args.TypeInfoV3.ArtifactDepends = deps
		}
		if provides != nil {
			provs := make(artifact.TypeInfoProvides, len(provides))
			for k, v := range provides {
				provs[k] = v
			}
			args.TypeInfoV3.ArtifactProvides = provs
		}
		return nil
	}
}

// WithPayloadFile attaches a randomly-generated update payload of the given
// size (in bytes) to the artifact's module image, so the resulting artifact
// file has a controllable, predictable size. This is needed by tests that
// exercise deployments' "select the smallest matching artifact" tie-break
// logic. Must be applied after the option that sets the module image (or
// rely on the default one CreateArtifact installs).
func WithPayloadFile(t interface{ TempDir() string }, size int) ArtifactOption {
	return func(args *awriter.WriteArtifactArgs) error {
		f, err := os.CreateTemp(t.TempDir(), "mender-artifact-payload-*")
		if err != nil {
			return fmt.Errorf("WithPayloadFile: failed to create payload file: %w", err)
		}
		defer f.Close()

		data := make([]byte, size)
		if _, err := rand.Read(data); err != nil {
			return fmt.Errorf("WithPayloadFile: failed to generate random payload: %w", err)
		}
		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("WithPayloadFile: failed to write payload file: %w", err)
		}

		if len(args.Updates.Updates) == 0 {
			return fmt.Errorf("WithPayloadFile: no updates to attach a payload to")
		}
		mi, ok := args.Updates.Updates[len(args.Updates.Updates)-1].(*handlers.ModuleImage)
		if !ok {
			return fmt.Errorf("WithPayloadFile: last update is not a *handlers.ModuleImage")
		}
		if err := mi.SetUpdateFiles([]*handlers.DataFile{{Name: f.Name()}}); err != nil {
			return fmt.Errorf("WithPayloadFile: failed to set update files: %w", err)
		}
		return nil
	}
}

// CreateArtifact writes a minimal module-image mender artifact to a temp
// file and returns it seeked to the start, ready for upload.
func CreateArtifact(
	name string,
	t interface{ TempDir() string },
	artifactArgsOpts ...ArtifactOption,
) (*os.File, error) {

	artifactDst := path.Join(t.TempDir(), fmt.Sprintf("%s.mender", name))
	file, err := os.Create(artifactDst)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", artifactDst, err)
	}

	w := awriter.NewWriter(file, artifact.NewCompressorGzip())
	i := handlers.NewModuleImage("foo")

	args := &awriter.WriteArtifactArgs{
		Format:  "mender",
		Version: 3,
		Name:    name,
		Provides: &artifact.ArtifactProvides{
			ArtifactName: name,
		},
		Depends: &artifact.ArtifactDepends{
			CompatibleDevices: []string{"foo"},
		},
		TypeInfoV3: &artifact.TypeInfoV3{
			Type:             i.GetUpdateType(),
			ArtifactProvides: artifact.TypeInfoProvides{"foo": "bar"},
			ArtifactDepends:  artifact.TypeInfoDepends{"foo": "bar"},
		},
		Updates: &awriter.Updates{
			Updates: []handlers.Composer{i},
		},
	}

	for _, opt := range artifactArgsOpts {
		if err := opt(args); err != nil {
			return nil, fmt.Errorf("failed to apply artifact option: %w", err)
		}
	}

	err = w.WriteArtifact(args)
	if err != nil {
		return nil, fmt.Errorf("failed to write module-image artifact: %w", err)
	}

	// Seek to the start of the file so the caller can read it
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare %s for reading: %w", artifactDst, err)
	}

	return file, nil
}
