package amproxy

import (
	"testing"

	"github.com/cheekybits/is"
)

func TestLoadFile(t *testing.T) {
	is := is.New(t)

	j, _, err := LoadUserConfigFile("fixtures/authfile.yaml")
	is.NoErr(err)

	_, ok := j["apikey"].Metrics["metric1"]
	is.True(ok)

	_, ok = j["apikey"].Metrics["metric3"]
	is.False(ok)
}
