package amproxy_test

import (
	"path"
	"runtime"
	"testing"

	"github.com/jasonhancock/amproxy"
)

func TestLoadFile(t *testing.T) {

	_, filename, _, _ := runtime.Caller(0)
	f := path.Join(path.Dir(filename), "fixtures", "authfile.yaml")

	j, _ := amproxy.LoadUserConfigFile(f)

	_, ok := j["apikey"].Metrics["metric1"]
	if !ok {
		t.Errorf("metric1 should be defined, but wasn't")
	}

	_, ok = j["apikey"].Metrics["metric3"]
	if ok {
		t.Errorf("metric3 should not be defined, but was")
	}
}
