package utils

import (
	"regexp"
	"sort"

	"github.com/hashicorp/go-version"
)

func GetResultsFromRegex(pattern, tmpl, content string) string {
	if pattern == "" || tmpl == "" {
		return content
	}
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatchIndex(content)
	result := regex.ExpandString([]byte{}, tmpl, content, matches)
	return string(result)
}

func MatchPattern(pattern, tmpl, version string) (bool, string) {
	result := GetResultsFromRegex(pattern, tmpl, version)
	return result != "", result
}

func MeetConstraint(constraint, ver string) (bool, error) {
	v, err := version.NewVersion(ver)
	if err != nil {
		return false, err
	}
	constraints, err := version.NewConstraint(constraint)
	if err != nil {
		return false, err
	}
	return constraints.Check(v), nil
}

func Contains(l []string, s string) bool {
	for _, a := range l {
		if a == s {
			return true
		}
	}
	return false
}

func RemoveDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	sort.Strings(list)
	return list
}
