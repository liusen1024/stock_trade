package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewURLWithValues(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		in     string
		values map[string]string
		want   string
	}{
		{in: "http://www.huaxingsec.com",
			values: map[string]string{
				"k1": "v1",
			},
			want: "http://www.huaxingsec.com?k1=v1",
		},
		{in: "http://www.huaxingsec.com?k1=v1",
			values: map[string]string{
				"k2": "v2",
			},
			want: "http://www.huaxingsec.com?k1=v1&k2=v2",
		},
	}

	for _, v := range cases {
		url, err := NewURLWithValues(v.in, v.values)
		fmt.Println(url.String())
		a.NoError(err)
		a.Equal(v.want, url.String())
	}
}
