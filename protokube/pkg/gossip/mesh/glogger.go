package mesh

import (
	"github.com/golang/glog"
	"github.com/weaveworks/mesh"
)

// glogLogger sends mesh log messages to glog
type glogLogger struct {
}

var _ mesh.Logger = &glogLogger{}

func (g *glogLogger) Printf(format string, args ...interface{}) {
	glog.Infof(format, args...)
}
