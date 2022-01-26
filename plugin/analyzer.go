package plugin

import (
	"bufio"
	"io"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/luraproject/lura/v2/core"
)

// Descriptor lists all the deps and versions required by a binary/plugin
type Descriptor struct {
	Go   string
	Deps map[string]string
}

// Local returns a descriptor for the binary calling it
func Local() Descriptor {
	return Descriptor{
		Go:   core.GoVersion,
		Deps: getBuildInfo(),
	}
}

// Diff points an incompatibility between descriptors
type Diff struct {
	Name     string
	Expected string
	Have     string
}

// Compare generates a list of diffs (incompatibility) between two descriptors
func (d Descriptor) Compare(other Descriptor) []Diff {
	diffs := []Diff{}

	for k, expect := range d.Deps {
		have, ok := other.Deps[k]
		if !ok {
			continue
		}
		if have != expect {
			diffs = append(diffs, Diff{Name: k, Expected: expect, Have: have})
		}
	}

	sort.Slice(diffs, func(i, j int) bool { return diffs[i].Name < diffs[j].Name })

	if d.Go != other.Go {
		tmp := make([]Diff, len(diffs))
		copy(tmp[1:], diffs)
		tmp[0] = Diff{Name: "go", Expected: d.Go, Have: other.Go}
		diffs = tmp
	}

	return diffs
}

// Describe reads a go.sum and returns a descriptor for the build or an error
func Describe(r io.Reader) (Descriptor, error) {
	content, err := parseSumFile(r)
	if err != nil {
		return Descriptor{}, err
	}

	deps := map[string]string{}
	for _, dep := range content {
		parts := strings.Split(dep, " ")
		cleanedVersion := cleanVersion(parts[1])

		if deps[parts[0]] >= cleanedVersion {
			continue
		}
		deps[parts[0]] = cleanedVersion
	}
	return Descriptor{
		Go:   "",
		Deps: deps,
	}, nil
}

func cleanVersion(v string) string {
	if l := len(v); l > 7 && v[l-7:] == "/go.mod" {
		return v[:l-7]
	}
	return v
}

func parseSumFile(r io.Reader) ([]string, error) {
	lines := []string{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func getBuildInfo() map[string]string {
	res := map[string]string{}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return res
	}

	for _, dep := range bi.Deps {
		res[dep.Path] = dep.Version
	}
	return res
}
