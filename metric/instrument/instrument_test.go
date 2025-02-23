// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package instrument // import "go.opentelemetry.io/otel/metric/instrument"

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel/attribute"
)

type attrConf interface {
	Attributes() attribute.Set
}

func TestConfigAttrs(t *testing.T) {
	t.Run("AddConfig", testConfAttr(func(mo ...MeasurementOption) attrConf {
		opts := make([]AddOption, len(mo))
		for i := range mo {
			opts[i] = mo[i].(AddOption)
		}
		return NewAddConfig(opts)
	}))

	t.Run("RecordConfig", testConfAttr(func(mo ...MeasurementOption) attrConf {
		opts := make([]RecordOption, len(mo))
		for i := range mo {
			opts[i] = mo[i].(RecordOption)
		}
		return NewRecordConfig(opts)
	}))

	t.Run("ObserveConfig", testConfAttr(func(mo ...MeasurementOption) attrConf {
		opts := make([]ObserveOption, len(mo))
		for i := range mo {
			opts[i] = mo[i].(ObserveOption)
		}
		return NewObserveConfig(opts)
	}))
}

func testConfAttr(newConf func(...MeasurementOption) attrConf) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("ZeroConfigEmpty", func(t *testing.T) {
			c := newConf()
			assert.Equal(t, *attribute.EmptySet(), c.Attributes())
		})

		t.Run("EmptySet", func(t *testing.T) {
			c := newConf(WithAttributeSet(*attribute.EmptySet()))
			assert.Equal(t, *attribute.EmptySet(), c.Attributes())
		})

		aliceAttr := attribute.String("user", "Alice")
		alice := attribute.NewSet(aliceAttr)
		t.Run("SingleWithAttributeSet", func(t *testing.T) {
			c := newConf(WithAttributeSet(alice))
			assert.Equal(t, alice, c.Attributes())
		})

		t.Run("SingleWithAttributes", func(t *testing.T) {
			c := newConf(WithAttributes(aliceAttr))
			assert.Equal(t, alice, c.Attributes())
		})

		bobAttr := attribute.String("user", "Bob")
		bob := attribute.NewSet(bobAttr)
		t.Run("MultiWithAttributeSet", func(t *testing.T) {
			c := newConf(WithAttributeSet(alice), WithAttributeSet(bob))
			assert.Equal(t, bob, c.Attributes())
		})

		t.Run("MergedWithAttributes", func(t *testing.T) {
			c := newConf(WithAttributes(aliceAttr, bobAttr))
			assert.Equal(t, bob, c.Attributes())
		})

		t.Run("MultiWithAttributeSet", func(t *testing.T) {
			c := newConf(WithAttributes(aliceAttr), WithAttributes(bobAttr))
			assert.Equal(t, bob, c.Attributes())
		})

		t.Run("MergedEmpty", func(t *testing.T) {
			c := newConf(WithAttributeSet(alice), WithAttributeSet(*attribute.EmptySet()))
			assert.Equal(t, alice, c.Attributes())
		})
	}
}

func TestWithAttributesRace(t *testing.T) {
	attrs := []attribute.KeyValue{
		attribute.String("user", "Alice"),
		attribute.Bool("admin", true),
		attribute.String("user", "Bob"),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		opt := []AddOption{WithAttributes(attrs...)}
		_ = NewAddConfig(opt)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		opt := []AddOption{WithAttributes(attrs...)}
		_ = NewAddConfig(opt)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		opt := []RecordOption{WithAttributes(attrs...)}
		_ = NewRecordConfig(opt)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		opt := []RecordOption{WithAttributes(attrs...)}
		_ = NewRecordConfig(opt)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		opt := []ObserveOption{WithAttributes(attrs...)}
		_ = NewObserveConfig(opt)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		opt := []ObserveOption{WithAttributes(attrs...)}
		_ = NewObserveConfig(opt)
	}()

	wg.Wait()
}
