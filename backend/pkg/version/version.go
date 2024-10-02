package version

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"text/tabwriter"
)

var (
	version string = "v0.0.0-dev"
)

type Version struct {
	Version      string
	GoVersion    string
	GitCommit    string `json:",omitempty"`
	RevisionTime string `json:",omitempty"`
	OS           string
	Arch         string
}

func Get() Version {
	v := Version{
		Version:   version,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
	if bldInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range bldInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				v.GitCommit = setting.Value
			case "vcs.time":
				v.RevisionTime = setting.Value
			}
		}
	}
	return v
}

func (v Version) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "Version:\t%s\n", version)
	fmt.Fprintf(w, "Go version:\t%s\n", v.GoVersion)
	fmt.Fprintf(w, "Git commit:\t%s\n", v.GitCommit)
	fmt.Fprintf(w, "Revision time:\t%s\n", v.RevisionTime)
	fmt.Fprintf(w, "Os/Arch:\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
	w.Flush()
	return buf.String()
}

func (v Version) MarshalJSON() ([]byte, error) {
	type version Version
	return json.Marshal((version)(v))
}
