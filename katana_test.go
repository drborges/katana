package katana_test

import (
	"github.com/drborges/katana"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
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
	type DepA struct {
		Field string
	}
	type DepB struct {
		Field string
	}

	Convey("Given I have an instance of katana injector with a few value providers", t, func() {
		depA := &DepA{}
		depB := DepB{}

		injector := katana.New().Provide(depA, depB)

		Convey("When I resolve instances of the provided values", func() {
			var depA1, depA2 *DepA
			var depB1, depB2 DepB

			err := injector.Resolve(&depA1, &depA2, &depB1, &depB2)

			Convey("Then instances of the same type are the same", func() {
				So(err, should.BeNil)
				So(depB1, should.NotBeNil)
				So(depB2, should.NotBeNil)
				So(depB1, should.Resemble, depB2)
				So(depA1, should.NotBeNil)
				So(depA2, should.NotBeNil)
				So(depA1, should.Equal, depA2)
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

		injector.Provide(&Dependency{})

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

func TestInvalidProviderFunction(t *testing.T) {
	// TODO write test case to catch whether the error is properly built
	Convey("Given I have a provider function with no return value for a given dependency", t, func() {
		invalidProvider := func() {
			katana.New().ProvideNew(&DependencyC{}, func() {})
		}

		Convey("Then it fails with an invalid provider error", func() {
			So(invalidProvider, should.Panic)
		})
	})

	Convey("Given I have a provider function with multiple return values for a given dependency", t, func() {
		invalidProvider := func() {
			katana.New().ProvideNew(&DependencyC{}, func() (*DependencyC, error) {
				return nil, nil
			})
		}

		Convey("Then it fails with an invalid provider error", func() {
			So(invalidProvider, should.Panic)
		})
	})
}

func TestProviderAlreadyRegistered(t *testing.T) {
	Convey("Given I have a provider registered for a given dependency", t, func() {
		injector := katana.New().ProvideNew(&DependencyC{}, func() *DependencyC {
			return &DependencyC{}
		})

		Convey("When I register another provider for that same dependency type", func() {
			alreadyRegisteredProvider := func() {
				injector.ProvideNew(&DependencyC{}, func() *DependencyC {
					return &DependencyC{}
				})
			}

			Convey("Then it fails with an already registered provider", func() {
				So(alreadyRegisteredProvider, should.Panic)
			})
		})
	})
}
