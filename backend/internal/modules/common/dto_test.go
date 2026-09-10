package common

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNormalizePagination(t *testing.T) {
	for _, info := range []PageInfo{{}, {Page: -1, PageSize: -1}, {Page: 2, PageSize: 100000}, {Page: int(^uint(0) >> 1), PageSize: 100}} {
		info.Normalize()
		require.GreaterOrEqual(t, info.Page, 1)
		require.GreaterOrEqual(t, info.PageSize, 1)
		require.LessOrEqual(t, info.PageSize, 100)
		require.GreaterOrEqual(t, (info.Page-1)*info.PageSize, 0)
	}
}
