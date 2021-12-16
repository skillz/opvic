package version

import (
	"github.com/hashicorp/go-version"
	"github.com/skillz/opvic/utils"
)

type RemoteVersions []*version.Version

type Versions struct {
	RunningVersion *version.Version
	RemoteVersions RemoteVersions
}

func (r *RemoteVersions) Latest() *version.Version {
	var latest *version.Version
	for _, version := range *r {
		if latest == nil || version.GreaterThan(latest) {
			latest = version
		}
	}
	return latest
}

func (r *RemoteVersions) Earliest() *version.Version {
	var earliest *version.Version
	for _, version := range *r {
		if earliest == nil || version.LessThan(earliest) {
			earliest = version
		}
	}
	return earliest
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
	return v.RemoteVersions.Earliest()
}

func (v *Versions) Latest() *version.Version {
	return v.RemoteVersions.Latest()
}

func (v *Versions) StringList() []string {
	vers := []string{}
	for _, version := range v.RemoteVersions {
		vers = append(vers, version.Original())
	}
	return utils.RemoveDuplicateStr(vers)
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

// Only returns last available majors greater than running version
func (v *Versions) LastMajorsGreaterThan() *Versions {
	var vers []*version.Version
	var uniqueMajors []int
	for _, version := range v.RemoteVersions {
		if version.Segments()[0] > v.RunningVersion.Segments()[0] {
			vers = append(vers, version)
			if !utils.ContainsInt(uniqueMajors, version.Segments()[0]) {
				uniqueMajors = append(uniqueMajors, version.Segments()[0])
			}
		}
	}

	uniqueMajorVersions := make(map[int]RemoteVersions, len(uniqueMajors))
	for _, version := range vers {
		uniqueMajorVersions[version.Segments()[0]] = append(uniqueMajorVersions[version.Segments()[0]], version)
	}
	var uniqueMajorLatests []*version.Version
	for _, versions := range uniqueMajorVersions {
		uniqueMajorLatests = append(uniqueMajorLatests, versions.Latest())
	}
	return &Versions{
		RunningVersion: v.RunningVersion,
		RemoteVersions: uniqueMajorLatests,
	}
}

func (v *Versions) MinorsGreaterThan() *Versions {
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

func (v *Versions) PatchesGreaterThan() *Versions {
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
	return len(v.LastMajorsGreaterThan().StringList()) > 0
}

func (v *Versions) MinorAvailable() bool {
	return len(v.MinorsGreaterThan().StringList()) > 0
}

func (v *Versions) PatchAvailable() bool {
	return len(v.PatchesGreaterThan().StringList()) > 0
}
