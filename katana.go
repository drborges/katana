package katana

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

var (
	TypeSingleton   = InstanceType("Singleton Dependency")
	TypeNewInstance = InstanceType("New Instance Dependency")
)

type InstanceType string
type Instance interface{}
type Provider interface{}
type Callable func() []Instance

type Dependency struct {
	Type             InstanceType
	Provider         Provider
	InjectedProvider Callable
}

type Trace []string

func (trace *Trace) Add(typ reflect.Type) (err error) {
	for _, t := range *trace {
		if t == typ.String() {
			defer trace.Reset()
			*trace = append(*trace, typ.String())
			err = ErrCyclicDependency{*trace}
			break
		}
	}
	*trace = append(*trace, typ.String())
	return err
}

func (trace *Trace) Reset() {
	*trace = Trace{}
}

type Injector struct {
	dependencies map[reflect.Type]*Dependency
	instances    map[reflect.Type]Instance
	trace        *Trace
}

func New() *Injector {
	return &Injector{
		dependencies: make(map[reflect.Type]*Dependency),
		instances:    make(map[reflect.Type]Instance),
		trace:        &Trace{},
	}
}

func (injector *Injector) Clone() *Injector {
	newInjector := New()

	for t, p := range injector.dependencies {
		newInjector.dependencies[t] = p
	}

	for t, i := range injector.instances {
		newInjector.instances[t] = i
	}

	return newInjector
}

func (injector *Injector) ProvideNew(i interface{}, p Provider) *Injector {
	t := reflect.TypeOf(i)
	if _, registered := injector.dependencies[t]; registered {
		log.Fatalf("Dependency %v already registered", i)
	}

	injector.dependencies[t] = &Dependency{
		Type:     TypeNewInstance,
		Provider: p,
	}

	return injector
}

func (injector *Injector) ProvideSingleton(i interface{}, p Provider) *Injector {
	t := reflect.TypeOf(i)
	if _, registered := injector.dependencies[t]; registered {
		log.Fatalf("Dependency %v already registered", i)
	}

	injector.dependencies[t] = &Dependency{
		Type:     TypeSingleton,
		Provider: p,
	}

	return injector
}

func (injector *Injector) Provide(values ...interface{}) *Injector {
	for _, value := range values {
		injector.ProvideSingleton(value, func(v interface{}) Provider {
			return func() Instance { return v }
		}(value))
	}
	return injector
}

func (injector *Injector) Resolve(items ...interface{}) error {
	for _, item := range items {
		val := reflect.ValueOf(item)
		typ := val.Type()

		if typ.Kind() != reflect.Ptr {
			return ErrNoSuchPtr{typ}
		}

		typ = typ.Elem()

		dep, registered := injector.dependencies[typ]
		if !registered {
			return ErrNoSuchProvider{typ}
		}

		if dep.Type == TypeSingleton {
			// Checks whether the item has already been resolved
			if inst, cached := injector.instances[typ]; cached {
				// Resolves the dependency with the cached instance
				val.Elem().Set(reflect.ValueOf(inst))
				continue
			}
		}

		if err := injector.trace.Add(typ); err != nil {
			return err
		}

		// Requests a new instance of the dependency from the provider
		if dep.InjectedProvider == nil {
			injectedProvider, err := injector.Inject(dep.Provider)
			if err != nil {
				return err
			}
			dep.InjectedProvider = injectedProvider
		}

		ret := dep.InjectedProvider()
		injector.trace.Reset()

		if len(ret) != 1 {
			return ErrInvalidProvider{reflect.TypeOf(dep.Provider)}
		}

		inst := ret[0]

		// Resolves the dependency with the new instance
		val.Elem().Set(reflect.ValueOf(inst))

		// Caches the instance in case the dependency is a singleton
		if injector.dependencies[typ].Type == TypeSingleton {
			injector.instances[typ] = inst
		}

	}

	return nil
}

func (injector *Injector) Inject(fn interface{}) (Callable, error) {
	val := reflect.ValueOf(fn)
	typ := val.Type()

	if typ.Kind() != reflect.Func {
		return nil, ErrNoSuchCallable{typ}
	}

	deps := make([]reflect.Value, typ.NumIn())
	for i := 0; i < typ.NumIn(); i++ {
		depVal := reflect.New(typ.In(i))
		dep := depVal.Interface()

		if err := injector.Resolve(dep); err != nil {
			return nil, err
		}

		deps[i] = depVal.Elem()
	}

	injected := func() []Instance {
		outVals := val.Call(deps)
		outs := make([]Instance, len(outVals))
		for i, val := range outVals {
			outs[i] = val.Interface()
		}
		return outs
	}

	return injected, nil
}

type ErrNoSuchPtr struct {
	Type reflect.Type
}

func (err ErrNoSuchPtr) Error() string {
	return fmt.Sprintf("Cannot resolve %v. Expected a pointer or an interface.", err.Type.Kind())
}

type ErrNoSuchCallable struct {
	Type reflect.Type
}

func (err ErrNoSuchCallable) Error() string {
	return fmt.Sprintf("Cannot inject dependencies into non callable type %v", err.Type.Kind())
}

type ErrNoSuchProvider struct {
	Type reflect.Type
}

func (err ErrNoSuchProvider) Error() string {
	return fmt.Sprintf("No providers registered for dependency type %v", err.Type)
}

type ErrCyclicDependency struct {
	Trace Trace
}

func (err ErrCyclicDependency) Error() string {
	return fmt.Sprintf("Cyclic dependency detected: [%v]", strings.Join(err.Trace, " -> "))
}

type ErrInvalidProvider struct {
	Type reflect.Type
}

func (err ErrInvalidProvider) Error() string {
	return fmt.Sprintf("Invalid provider function: %v", err.Type.String())
}
