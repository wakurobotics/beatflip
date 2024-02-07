package supervisor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupervisor_compareSemver(t *testing.T) {
	res := compareSemver([]int{1, 0, 0}, []int{1, 0, 0})
	assert.Equal(t, 0, res)

	res = compareSemver([]int{1, 0, 0}, []int{1, 1, 0})
	assert.Equal(t, 1, res)

	res = compareSemver([]int{1, 0, 0}, []int{1, 0, 2})
	assert.Equal(t, 1, res)

	res = compareSemver([]int{1, 0, 0}, []int{1, 1, 1})
	assert.Equal(t, 1, res)

	res = compareSemver([]int{1, 0, 0}, []int{2, 0, 0})
	assert.Equal(t, 1, res)

	res = compareSemver([]int{1, 0, 0}, []int{1, 0})
	assert.Equal(t, 0, res)

	res = compareSemver([]int{1, 0}, []int{1, 0, 0})
	assert.Equal(t, 0, res)
}

func TestSupervisor_parseSemver(t *testing.T) {
	res, err := parseSemver("1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 0, 0}, res)

	res, err = parseSemver("v1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 0, 0}, res)

	res, err = parseSemver("v1.0.1")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 0, 1}, res)

	res, err = parseSemver("v1.0")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 0}, res)
}
