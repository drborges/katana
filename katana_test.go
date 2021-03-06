package katana_test

import (
	"github.com/drborges/katana"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type Dependency struct {
	Field string
}

type DependencyA struct {
	Dep *Dependency
}

type DependencyB struct {
	Dep *DependencyA
}

type DependencyC struct {
	Dep *DependencyD
}

type DependencyD struct {
	Dep *DependencyC
}

type InterfaceDependency interface {
	DoStuff()
}

type InterfaceDependencyImpl struct {
	Field string
}

func (dep *InterfaceDependencyImpl) DoStuff() {}

func TestKatanaProvide(t *testing.T) {
	Convey("Given I have an instance of katana injector with a few value providers", t, func() {
		depA := &Dependency{}
		depB := Dependency{}

		injector := katana.New().Provide(depA, depB)

		Convey("When I resolve instances of the provided values", func() {
			var depA1, depA2 *Dependency
			var depB1, depB2 Dependency

			injector.Resolve(&depA1, &depA2, &depB1, &depB2)

			Convey("Then instances of the same type are the same", func() {
				So(depA1, should.NotBeNil)
				So(depA2, should.NotBeNil)
				So(depA1, should.Equal, depA2)
				So(depB1, should.NotBeNil)
				So(depB2, should.NotBeNil)
				So(depB1, should.NotEqual, depB2)
				So(depB1, should.Resemble, depB2)
			})
		})
	})
}

func TestKatanaProvideAs(t *testing.T) {
	Convey("Given I have an instance of katana injector with an instance provided as an interface type", t, func() {
		dep := &InterfaceDependencyImpl{}

		injector := katana.New().ProvideAs((*InterfaceDependency)(nil), dep)

		Convey("When I resolve a reference to the provided dependency", func() {
			var depA InterfaceDependency

			injector.Resolve(&depA)

			Convey("Then the reference is successfully resolved", func() {
				So(depA, should.NotBeNil)
				So(depA, should.Equal, dep)
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
			injector.Resolve(&dep1, &dep2)

			Convey("Then the resolved dependnecies points to different memory address", func() {
				So(dep1, should.NotEqual, dep2)
			})
		})
	})

	Convey("Given I have two new instance providers one for *Dependency and another one for *DependencyB", t, func() {
		injector := katana.New().ProvideNew(&Dependency{}, func() *Dependency {
			return &Dependency{}
		})

		injector.ProvideNew(&DependencyA{}, func(dep *Dependency) *DependencyA {
			return &DependencyA{dep}
		})

		Convey("When I resolve multiple instances of *DependencyB which depends on *Dependency", func() {
			var dep1, dep2 *DependencyA
			injector.Resolve(&dep1, &dep2)

			Convey("Then the resolved dependencies point to different memory addresses", func() {
				So(dep1, should.NotEqual, dep2)

				Convey("And its dependencies also point to different memory addresses", func() {
					So(dep1.Dep, should.NotBeNil)
					So(dep2.Dep, should.NotBeNil)
					So(dep1.Dep, should.NotEqual, dep2.Dep)
				})
			})
		})
	})
}

func TestKatanaProvidesSingletonInstance(t *testing.T) {
	Convey("Given I have an instance of katana injector with a singleton dependency provider", t, func() {
		injector := katana.New().ProvideSingleton(&Dependency{}, func() *Dependency {
			return &Dependency{}
		})

		Convey("When I resolve multiple instances of the provided dependency", func() {
			var dep1, dep2 *Dependency
			injector.Resolve(&dep1, &dep2)

			Convey("Then the resolved dependencies points to the same memory address", func() {
				So(dep1, should.Equal, dep2)
			})
		})
	})
}

func TestKatanaProvidesSingletonInstanceOfInterfaceType(t *testing.T) {
	Convey("Given I have a provider of an interface dependency type", t, func() {
		injector := katana.New().ProvideSingleton((*InterfaceDependency)(nil), func() InterfaceDependency {
			return &InterfaceDependencyImpl{}
		})

		Convey("When I resolve multiple instances of the provided dependency", func() {
			var dep1, dep2 InterfaceDependency
			injector.Resolve(&dep1, &dep2)

			Convey("Then the dependencies are resolved", func() {
				So(dep1, should.Equal, dep2)
				So(dep1, should.HaveSameTypeAs, &InterfaceDependencyImpl{})
				So(dep2, should.HaveSameTypeAs, &InterfaceDependencyImpl{})
			})
		})
	})
}

func TestKatanaProvidesNewInstanceOfInterfaceType(t *testing.T) {
	Convey("Given I have a provider of an interface dependency type", t, func() {
		injector := katana.New().ProvideNew((*InterfaceDependency)(nil), func() InterfaceDependency {
			return &InterfaceDependencyImpl{}
		})

		Convey("When I resolve multiple instances of the provided dependency", func() {
			var dep1, dep2 InterfaceDependency
			injector.Resolve(&dep1, &dep2)

			Convey("Then the dependencies are resolved", func() {
				So(dep1, should.NotEqual, dep2)
				So(dep1, should.HaveSameTypeAs, &InterfaceDependencyImpl{})
				So(dep2, should.HaveSameTypeAs, &InterfaceDependencyImpl{})
			})
		})
	})
}

func TestKatanaResolvesTransitiveDependencies(t *testing.T) {
	Convey("Given I have transitive dependencies", t, func() {
		injector := katana.New().ProvideNew(&DependencyB{}, func(dep *DependencyA) *DependencyB {
			return &DependencyB{dep}
		})

		injector.ProvideNew(&DependencyA{}, func(dep *Dependency) *DependencyA {
			return &DependencyA{dep}
		})

		injector.Provide(&Dependency{})

		Convey("When I resolve the root dep", func() {
			var depA *DependencyB
			injector.Resolve(&depA)

			Convey("Then all dependencies are resolved recursively", func() {
				So(depA, should.NotBeNil)
				So(depA.Dep, should.NotBeNil)
				So(depA.Dep.Dep, should.NotBeNil)
			})
		})
	})
}

type DepA struct {
	depB *DepB
	depD *DepD
}

type DepB struct {
	field string
}

type DepC struct {
	dep *DepA
}

type DepD struct {
	dep *DepC
}

func TestErrCyclicDependency(t *testing.T) {

	Convey("Given I have cyclic dependencies", t, func() {
		injector := katana.New().ProvideNew(&DepA{}, func(depB *DepB, depD *DepD) *DepA {
			return &DepA{depB, depD}
		})

		injector.ProvideNew(&DepB{}, func() *DepB {
			return &DepB{}
		})

		injector.ProvideNew(&DepC{}, func(dep *DepA) *DepC {
			return &DepC{dep}
		})

		injector.ProvideNew(&DepD{}, func(dep *DepC) *DepD {
			return &DepD{dep}
		})

		Convey("When I resolve the cyclic dependency", func() {
			var dep *DepA
			resolveWithCyclicDependency := func() {
				injector.Resolve(&dep)
			}

			Convey("Then all dependencies are resolved recursively", func() {
				So(resolveWithCyclicDependency, should.Panic)
			})
		})
	})
}

func TestErrInvalidProvider(t *testing.T) {
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

func TestErrProviderAlreadyRegistered(t *testing.T) {
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

func TestErrInvalidReference(t *testing.T) {
	Convey("Given I have a provider registered for a given dependency", t, func() {
		injector := katana.New().ProvideNew(Dependency{}, func() *Dependency {
			return &Dependency{}
		})

		Convey("Then it panics resolving a reference that is a zero value", func() {
			var ref *Dependency
			resolvingZeroValue := func() { injector.Resolve(ref) }

			So(resolvingZeroValue, should.Panic)
		})
	})
}

func TestInjectorClone(t *testing.T) {
	Convey("Given I have an injector with a few injectables and cached instances", t, func() {
		injector := katana.New()

		injector.Provide(&DependencyB{})

		injector.ProvideNew(&Dependency{}, func() *Dependency {
			return &Dependency{}
		})

		injector.ProvideSingleton(&DependencyA{}, func(dep *Dependency) *DependencyA {
			return &DependencyA{dep}
		})

		Convey("When I clone the injector", func() {
			newInjector := injector.Clone()

			Convey("Then the new injector inherits the original injector providers", func() {
				var dep *Dependency
				var depA *DependencyA
				var depB *DependencyB
				newInjector.Resolve(&dep, &depA, &depB)

				So(dep, should.NotBeNil)
				So(depA, should.NotBeNil)
				So(depB, should.NotBeNil)
			})

			Convey("And I register new providers with the new injector", func() {
				newInjector.ProvideSingleton((*InterfaceDependency)(nil), func() InterfaceDependency {
					return &InterfaceDependencyImpl{}
				})

				Convey("Then the provider is available only in the cloned injector", func() {
					var dep1, dep2 InterfaceDependency
					newInjector.Resolve(&dep1)

					So(dep1, should.NotBeNil)

					resolveWithOriginalInjector := func () { injector.Resolve(&dep2) }

					So(resolveWithOriginalInjector, should.Panic)
				})
			})
		})
	})
}
