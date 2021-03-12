/*
Copyright 2021 The Cockroach Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logging

import (
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

type Logging struct {
	log logr.Logger
}

func NewLogging(log logr.Logger) *Logging {
	return &Logging{
		log: log,
	}
}

func (l Logging) GetLog() logr.Logger {
	return l.log
}

func (l Logging) Error(err error, msg string, keysAndValues ...interface{}) {
	if keysAndValues == nil {
		l.log.Error(err, msg)
	} else {
		l.log.Error(err, msg, keysAndValues)
	}
}

func (l Logging) LogAndWrapError(err error, msg string) error {
	l.log.Error(err, msg)
	return errors.Wrap(err, msg)
}

func (l Logging) Warn(msg string, keysAndValues ...interface{}) {
	l.WithLevel(int(zapcore.WarnLevel), msg, keysAndValues...)
}

func (l Logging) Info(msg string, keysAndValues ...interface{}) {
	l.WithLevel(int(zapcore.InfoLevel), msg, keysAndValues...)
}

func (l Logging) Debug(msg string, keysAndValues ...interface{}) {
	l.WithLevel(int(zapcore.DebugLevel), msg, keysAndValues...)
}

func (l Logging) WithLevel(level int, msg string, keysAndValues ...interface{}) {
	l.log.V(level).Info(msg, keysAndValues...)
}

func isNil(v ...interface{}) bool {
	if v == nil {
		return true
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(v).IsNil()
	}
	return false
}
