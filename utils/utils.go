package utils

import (
	"regexp"
	"sort"

	"github.com/hashicorp/go-version"
)

func GetResultsFromRegex(pattern, tmpl, content string) (string, error) {
	if pattern == "" || tmpl == "" {
		return content, nil
	}
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	matches := regex.FindStringSubmatchIndex(content)
	result := regex.ExpandString([]byte{}, tmpl, content, matches)
	return string(result), nil
}

func MatchPattern(pattern, tmpl, version string) (bool, string, error) {
	result, err := GetResultsFromRegex(pattern, tmpl, version)
	if err != nil {
		return false, "", err
	}
	return result != "", result, nil
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

func ContainsInt(l []int, i int) bool {
	for _, a := range l {
		if a == i {
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
