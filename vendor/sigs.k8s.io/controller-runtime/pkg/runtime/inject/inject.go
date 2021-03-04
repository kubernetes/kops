/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package inject is used by a Manager to inject types into Sources, EventHandlers, Predicates, and Reconciles.
// Deprecated: Use manager.Options fields directly. This package will be removed in v0.10.
package inject

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
)

// log is specifically to add a warning message for injectors.
var log = logf.RuntimeLog.WithName("injectors-warning")

// logWarningMsg logs a warning message if inject is used
func logWarningMsg() {
	log.Info("Injectors are deprecated, and will be removed in v0.10.x")
}

// Cache is used by the ControllerManager to inject Cache into Sources, EventHandlers, Predicates, and
// Reconciles
type Cache interface {
	InjectCache(cache cache.Cache) error
}

// CacheInto will set informers on i and return the result if it implements Cache.  Returns
// false if i does not implement Cache.
func CacheInto(c cache.Cache, i interface{}) (bool, error) {
	if s, ok := i.(Cache); ok {
		logWarningMsg()
		return true, s.InjectCache(c)
	}
	return false, nil
}

// APIReader is used by the Manager to inject the APIReader into necessary types.
type APIReader interface {
	InjectAPIReader(client.Reader) error
}

// APIReaderInto will set APIReader on i and return the result if it implements APIReaderInto.
// Returns false if i does not implement APIReader
func APIReaderInto(reader client.Reader, i interface{}) (bool, error) {
	if s, ok := i.(APIReader); ok {
		logWarningMsg()
		return true, s.InjectAPIReader(reader)
	}
	return false, nil
}

// Config is used by the ControllerManager to inject Config into Sources, EventHandlers, Predicates, and
// Reconciles
type Config interface {
	InjectConfig(*rest.Config) error
}

// ConfigInto will set config on i and return the result if it implements Config.  Returns
// false if i does not implement Config.
func ConfigInto(config *rest.Config, i interface{}) (bool, error) {
	if s, ok := i.(Config); ok {
		logWarningMsg()
		return true, s.InjectConfig(config)
	}
	return false, nil
}

// Client is used by the ControllerManager to inject client into Sources, EventHandlers, Predicates, and
// Reconciles
type Client interface {
	InjectClient(client.Client) error
}

// ClientInto will set client on i and return the result if it implements Client. Returns
// false if i does not implement Client.
func ClientInto(client client.Client, i interface{}) (bool, error) {
	if s, ok := i.(Client); ok {
		logWarningMsg()
		return true, s.InjectClient(client)
	}
	return false, nil
}

// Scheme is used by the ControllerManager to inject Scheme into Sources, EventHandlers, Predicates, and
// Reconciles
type Scheme interface {
	InjectScheme(scheme *runtime.Scheme) error
}

// SchemeInto will set scheme and return the result on i if it implements Scheme.  Returns
// false if i does not implement Scheme.
func SchemeInto(scheme *runtime.Scheme, i interface{}) (bool, error) {
	if is, ok := i.(Scheme); ok {
		logWarningMsg()
		return true, is.InjectScheme(scheme)
	}
	return false, nil
}

// Stoppable is used by the ControllerManager to inject stop channel into Sources,
// EventHandlers, Predicates, and Reconciles.
type Stoppable interface {
	InjectStopChannel(<-chan struct{}) error
}

// StopChannelInto will set stop channel on i and return the result if it implements Stoppable.
// Returns false if i does not implement Stoppable.
func StopChannelInto(stop <-chan struct{}, i interface{}) (bool, error) {
	if s, ok := i.(Stoppable); ok {
		logWarningMsg()
		return true, s.InjectStopChannel(stop)
	}
	return false, nil
}

// Mapper is used to inject the rest mapper to components that may need it
type Mapper interface {
	InjectMapper(meta.RESTMapper) error
}

// MapperInto will set the rest mapper on i and return the result if it implements Mapper.
// Returns false if i does not implement Mapper.
func MapperInto(mapper meta.RESTMapper, i interface{}) (bool, error) {
	if m, ok := i.(Mapper); ok {
		logWarningMsg()
		return true, m.InjectMapper(mapper)
	}
	return false, nil
}

// Func injects dependencies into i.
type Func func(i interface{}) error

// Injector is used by the ControllerManager to inject Func into Controllers
type Injector interface {
	InjectFunc(f Func) error
}

// InjectorInto will set f and return the result on i if it implements Injector.  Returns
// false if i does not implement Injector.
func InjectorInto(f Func, i interface{}) (bool, error) {
	if ii, ok := i.(Injector); ok {
		logWarningMsg()
		return true, ii.InjectFunc(f)
	}
	return false, nil
}

// Logger is used to inject Loggers into components that need them
// and don't otherwise have opinions.
type Logger interface {
	InjectLogger(l logr.Logger) error
}

// LoggerInto will set the logger on the given object if it implements inject.Logger,
// returning true if a InjectLogger was called, and false otherwise.
func LoggerInto(l logr.Logger, i interface{}) (bool, error) {
	if injectable, wantsLogger := i.(Logger); wantsLogger {
		logWarningMsg()
		return true, injectable.InjectLogger(l)
	}
	return false, nil
}
