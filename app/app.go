// Package app is responsible for the types and behavior required
// to run the application.
package app

import (
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

const (
	ServiceName = "limits"
)

type Dependencies struct {
	ServiceName string
	Build       string
	Host        string
	Kubernetes  KubeInfo
	Shutdown    chan os.Signal
	Logger      *zap.SugaredLogger
}

type KubeInfo struct {
	Pod       string
	PodIP     string
	Node      string
	Namespace string
}

// RootDir returns the absolute path to the root directory of project
func RootDir() string {
	_, f, _, _ := runtime.Caller(0)

	thisDir := path.Join(path.Dir(f))
	return filepath.Dir(thisDir)
}
