package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBool(t *testing.T) {
	c := &gin.Context{
		Request: &http.Request{
			Form: map[string][]string{
				"hello": {"true"},
			},
		},
	}

	boolRes, err := Bool(c, "hello")
	require.True(t, err == nil)
	require.True(t, boolRes)
}

func TestBoolWithDefault(t *testing.T) {
	c := &gin.Context{
		Request: &http.Request{
			Form: map[string][]string{
				"hello": {"true"},
			},
		},
	}

	boolRes := BoolWithDefault(c, "hello", true)
	require.True(t, boolRes)

	c = &gin.Context{
		Request: &http.Request{
			Form: map[string][]string{
				"hello": {"aaaa"},
			},
		},
	}

	boolRes = BoolWithDefault(c, "hello", false)
	require.False(t, boolRes)
}

func TestInt32(t *testing.T) {
	c := &gin.Context{
		Request: &http.Request{
			Form: map[string][]string{
				"hello": {"10"},
			},
		},
	}
	res, err := Int64(c, "hello")
	require.True(t, err == nil)
	require.Equal(t, res, int32(10))
}

func TestStringPtr(t *testing.T) {
	c := &gin.Context{
		Request: &http.Request{
			Form: map[string][]string{
				"hello": {"world"},
			},
		},
	}
	res, err := String(c, "hello")
	require.True(t, err == nil)
	require.Equal(t, res, "world")
}

func TestInt64WithDefault(t *testing.T) {
	c := &gin.Context{
		Request: &http.Request{
			Form: map[string][]string{
				"hello": {"10"},
			},
		},
	}
	res := Int64WithDefault(c, "hello", 200)
	require.True(t, res == 10)
}
