package supervisor

import (
	"regexp"
	"strconv"
	"strings"
)

var reCleanSemver = regexp.MustCompile(`[^0-9.]{1,}`)

func parseSemver(s string) ([]int, error) {
	v := strings.TrimLeft(s, "v \n\r")
	v = strings.TrimRight(v, " \n\r")
	v = reCleanSemver.ReplaceAllString(v, ".")

	parsed := []int{}
	chunks := strings.Split(v, ".")
	for _, c := range chunks {
		i, err := strconv.Atoi(c)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, i)
	}
	return parsed, nil
}

func compareSemver(v, w []int) int {
	length := MaxInt(len(v), len(w))
	v = fillIntSlice(v, length-len(v), 0)
	w = fillIntSlice(w, length-len(w), 0)

	for idx := 0; idx < length; idx++ {
		if v[idx] < w[idx] {
			return 1
		}
	}
	return 0
}

func fillIntSlice(s []int, amount int, value int) []int {
	for i := 0; i < amount; i++ {
		s = append(s, value)
	}
	return s
}

func MaxInt(v, w int) int {
	if v > w {
		return v
	}
	return w
}
