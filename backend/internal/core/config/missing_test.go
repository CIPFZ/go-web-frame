package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMissingListsEmptyItemsInOrder(t *testing.T) {
	require.Equal(t, []string{"api_base", "app_key"}, Missing(map[string]bool{"app_key": true, "region": false, "api_base": true}))
	require.Empty(t, Missing(map[string]bool{"region": false}))
}
