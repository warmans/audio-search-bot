package searchterms

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestType_Equal(t *testing.T) {
	require.True(t, StringType.Equal(StringType))
	require.True(t, IntType.Equal(IntType))
}

func TestType_String(t *testing.T) {
	require.EqualValues(t, "int", IntType.String())
}
