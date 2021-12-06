package version

import (
	"github.com/hashicorp/go-version"
)

type Versions struct {
	RunningVersion *version.Version
	RemoteVersions []*version.Version
}

func NewVersions(running string, remotes []string) (*Versions, error) {
	var runningVer *version.Version
	if len(remotes) == 0 && running == "" {
		return &Versions{}, nil
	}
	var vers []*version.Version
	for _, v := range remotes {
		v, err := version.NewVersion(v)
		if err != nil {
			return nil, err
		}
		vers = append(vers, v)
	}
	if running != "" {
		ver, err := version.NewVersion(running)
		if err != nil {
			return nil, err
		}
		runningVer = ver
	}
	return &Versions{
		RunningVersion: runningVer,
		RemoteVersions: vers,
	}, nil
}

func (v *Versions) SetRunningVersion(ver string) error {
	version, err := version.NewVersion(ver)
	if err != nil {
		return err
	}
	v.RunningVersion = version
	return nil
}

func (v *Versions) GetRunningVersion() *version.Version {
	return v.RunningVersion
}

func (v *Versions) SetRemoteVersions(remotes []string) error {
	var vers []*version.Version
	for _, v := range remotes {
		v, err := version.NewVersion(v)
		if err != nil {
			return err
		}
		vers = append(vers, v)
	}
	v.RemoteVersions = vers
	return nil
}

func (v *Versions) Earliest() *version.Version {
	var earliest *version.Version
	for _, v := range v.RemoteVersions {
		if earliest == nil || v.LessThan(earliest) {
			earliest = v
		}
	}
	return earliest
}

func (v *Versions) Latest() *version.Version {
	var latest *version.Version
	for _, version := range v.RemoteVersions {
		if latest == nil || version.GreaterThan(latest) {
			latest = version
		}
	}
	return latest
}

func (v *Versions) StringList() []string {
	vers := []string{}
	for _, version := range v.RemoteVersions {
		vers = append(vers, version.String())
	}
	return vers
}

func (v *Versions) GreaterThan() *Versions {
	var vers []*version.Version
	for _, version := range v.RemoteVersions {
		if version.GreaterThan(v.RunningVersion) {
			vers = append(vers, version)
		}
	}
	return &Versions{
		RunningVersion: v.RunningVersion,
		RemoteVersions: vers,
	}
}

func (v *Versions) MajorGreaterThan() *Versions {
	var vers []*version.Version
	for _, version := range v.RemoteVersions {
		if version.Segments()[0] > v.RunningVersion.Segments()[0] {
			vers = append(vers, version)
		}
	}
	return &Versions{
		RunningVersion: v.RunningVersion,
		RemoteVersions: vers,
	}
}

func (v *Versions) MinorGreaterThan() *Versions {
	var vers []*version.Version
	for _, version := range v.RemoteVersions {
		if version.Segments()[0] == v.RunningVersion.Segments()[0] && version.Segments()[1] > v.RunningVersion.Segments()[1] {
			vers = append(vers, version)
		}
	}
	return &Versions{
		RunningVersion: v.RunningVersion,
		RemoteVersions: vers,
	}
}

func (v *Versions) PatchGreaterThan() *Versions {
	var vers []*version.Version
	for _, version := range v.RemoteVersions {
		if version.Segments()[0] == v.RunningVersion.Segments()[0] && version.Segments()[1] == v.RunningVersion.Segments()[1] && version.Segments()[2] > v.RunningVersion.Segments()[2] {
			vers = append(vers, version)
		}
	}
	return &Versions{
		RunningVersion: v.RunningVersion,
		RemoteVersions: vers,
	}
}

func (v *Versions) MajorAvailable() bool {
	return len(v.MajorGreaterThan().StringList()) > 0
}

func (v *Versions) MinorAvailable() bool {
	return len(v.MinorGreaterThan().StringList()) > 0
}

func (v *Versions) PatchAvailable() bool {
	return len(v.PatchGreaterThan().StringList()) > 0
}
