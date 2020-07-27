package utils

import (
	"regexp"
	"strconv"
	"strings"
)

func StringToInt32Array(v string, split string) []int32 {
	str := StringToArray(v,split)
	if str == nil {
		return nil
	}

	data := make([]int32, 0)

	for i := 0; i < len(str); i++ {
		d := str[i]
		if d != "" {
			va, err := strconv.ParseInt(d, 10, 32)
			if err == nil {
				data = append(data, int32(va))
			}
		}
	}
	return data
}

func StringToArray(v string, split string) []string {
	if v == "null" || v == "" {
		return nil
	}
	reg := regexp.MustCompile("\\[|\\]|\"|\n|\t|\r| ")
	v = reg.ReplaceAllString(v, "")
	str := strings.Split(v, split)
	return str
}
