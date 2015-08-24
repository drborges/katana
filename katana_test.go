package katana_test

import (
	"github.com/drborges/katana"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"reflect"
)

// @katana.New
// @katana.Generate.PtrProvider
type Dependency struct {
	Field string
}

// @katana.Singleton
// @katana.Generate.ValueProvider
type DependencyB struct {
	Dep *Dependency
}

// @katana.Singleton
// @katana.Generate.PtrProvider
type DependencyA struct {
	Dep *DependencyB
}

// @katana.Singleton
// @katana.Generate.PtrProvider
type DependencyC struct {
	Dep *DependencyD
}

// @katana.Singleton
// @katana.Generate.PtrProvider
type DependencyD struct {
	Dep *DependencyC
}

func TestKatanaProvideValues(t *testing.T) {
	type Cache struct {
		Field string
	}
	type Request struct {
		Field string
	}

	Convey("Given I have an instance of katana injector with a few value providers", t, func() {
		cache := &Cache{}
		req := Request{}

		injector := katana.New().ProvideValue(cache, req)

		Convey("When I resolve instances of the provided values", func() {
			var cache1, cache2 *Cache
			var req1, req2 Request

			err := injector.Resolve(&cache1, &cache2, &req1, &req2)

			Convey("Then instances of the same type are the same", func() {
				So(err, should.BeNil)
				So(req1, should.NotBeNil)
				So(req2, should.NotBeNil)
				So(req1, should.Resemble, req2)
				So(cache1, should.NotBeNil)
				So(cache2, should.NotBeNil)
				So(cache1, should.Equal, cache2)
			})
		})
	})
}

func TestKatanaProvideNewInstance(t *testing.T) {
	Convey("Given I have an instance of katana injector with a new instance provider of a dependency", t, func() {
		injector := katana.New().ProvideNew(&Dependency{}, func() *Dependency {
			return &Dependency{}
		})

		Convey("When I resolve multiple instances of the provided dependency", func() {
			var dep1, dep2 *Dependency
			err := injector.Resolve(&dep1, &dep2)

			Convey("Then the resolved dependnecies points to different memory address", func() {
				So(err, should.BeNil)
				So(dep1, should.NotEqual, dep2)
			})
		})
	})
}

func TestKatanaProvideSingletonInstance(t *testing.T) {
	Convey("Given I have an instance of katana injector with a singleton dependency provider", t, func() {
		injector := katana.New().ProvideSingleton(&Dependency{}, func() *Dependency {
			return &Dependency{}
		})

		Convey("When I resolve multiple instances of the provided dependency", func() {
			var dep1, dep2 *Dependency
			err := injector.Resolve(&dep1, &dep2)

			Convey("Then the resolved dependencies points to the same memory address", func() {
				So(err, should.BeNil)
				So(dep1, should.Equal, dep2)
			})
		})
	})
}

func TestKatanaResolvesTransitiveDependencies(t *testing.T) {
	Convey("Given I have transitive dependencies", t, func() {
		injector := katana.New().ProvideNew(&DependencyA{}, func(dep *DependencyB) *DependencyA {
			return &DependencyA{dep}
		})

		injector.ProvideNew(&DependencyB{}, func(dep *Dependency) *DependencyB {
			return &DependencyB{dep}
		})

		injector.ProvideValue(&Dependency{})

		Convey("When I resolve the root dep", func() {
			var depA *DependencyA
			err := injector.Resolve(&depA)

			Convey("Then all dependencies are resolved recursively", func() {
				So(err, should.BeNil)
				So(depA, should.NotBeNil)
				So(depA.Dep, should.NotBeNil)
				So(depA.Dep.Dep, should.NotBeNil)
			})
		})
	})
}

func TestKatanaDetectsCyclicDependencies(t *testing.T) {

	Convey("Given I have cyclic dependencies", t, func() {
		injector := katana.New().ProvideNew(&DependencyC{}, func(dep *DependencyD) *DependencyC {
			return &DependencyC{dep}
		})

		injector.ProvideNew(&DependencyD{}, func(dep *DependencyC) *DependencyD {
			return &DependencyD{dep}
		})

		Convey("When I resolve the cyclic dependency", func() {
			var dep *DependencyC
			err := injector.Resolve(&dep)

			Convey("Then all dependencies are resolved recursively", func() {
				So(err, should.Resemble, katana.ErrCyclicDependency{katana.Trace{
					reflect.TypeOf(&DependencyC{}).String(),
					reflect.TypeOf(&DependencyD{}).String(),
					reflect.TypeOf(&DependencyC{}).String(),
				}})
			})
		})
	})
}
