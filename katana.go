package katana

import (
	"fmt"
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

func ValidateProvider(provider Provider) error {
	typ := reflect.TypeOf(provider)

	if typ.Kind() != reflect.Func {
		return ErrNoSuchCallable{typ}
	}

	if typ.NumOut() != 1 {
		return ErrInvalidProvider{typ}
	}

	return nil
}

type Dependency struct {
	Type     InstanceType
	Provider Provider
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

func (injector *Injector) ProvideNew(dep interface{}, p Provider) *Injector {
	typ := reflect.TypeOf(dep)

	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Interface {
		typ = typ.Elem()
	}

	if _, registered := injector.dependencies[typ]; registered {
		panic(ErrProviderAlreadyRegistered{typ})
	}

	if err := ValidateProvider(p); err != nil {
		panic(err)
	}

	injector.dependencies[typ] = &Dependency{
		Type:     TypeNewInstance,
		Provider: p,
	}

	return injector
}

func (injector *Injector) ProvideSingleton(dep interface{}, p Provider) *Injector {
	typ := reflect.TypeOf(dep)

	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Interface {
		typ = typ.Elem()
	}

	if _, registered := injector.dependencies[typ]; registered {
		panic(ErrProviderAlreadyRegistered{typ})
	}

	if err := ValidateProvider(p); err != nil {
		panic(err)
	}

	injector.dependencies[typ] = &Dependency{
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

func (injector *Injector) Resolve(items ...interface{}) {
	for _, item := range items {
		val := reflect.ValueOf(item)
		typ := val.Type()

		if typ.Kind() != reflect.Ptr {
			panic(ErrNoSuchPtr{typ})
		}

		typ = typ.Elem()

		// Checks whether there is a registered provider for the dependency
		dep, registered := injector.dependencies[typ]
		if !registered {
			panic(ErrNoSuchProvider{typ})
		}

		// Checks instances cache for previous resolved dependency in case it is a singleton one
		if dep.Type == TypeSingleton {
			if inst, cached := injector.instances[typ]; cached {
				// Resolves the dependency with the cached instance
				val.Elem().Set(reflect.ValueOf(inst))
				continue
			}
		}

		// Add to the trace the current dependency type being resolved
		// so that cyclic dependencies may be detected
		if err := injector.trace.Add(typ); err != nil {
			panic(err)
		}

		// Resolves the provider arguments -- if any -- as dependencies returning
		// a closure with the resolved arguments injected
		inst := injector.Inject(dep.Provider)()[0]
		injector.trace.Reset()

		// Resolves the dependency with the new instance
		val.Elem().Set(reflect.ValueOf(inst))

		// Caches the instance in case the dependency is a singleton
		if injector.dependencies[typ].Type == TypeSingleton {
			injector.instances[typ] = inst
		}
	}
}

func (injector *Injector) Inject(fn interface{}) Callable {
	val := reflect.ValueOf(fn)
	typ := val.Type()

	if typ.Kind() != reflect.Func {
		panic(ErrNoSuchCallable{typ})
	}

	deps := make([]reflect.Value, typ.NumIn())
	for i := 0; i < typ.NumIn(); i++ {
		depVal := reflect.New(typ.In(i))
		dep := depVal.Interface()
		injector.Resolve(dep)
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

	return injected
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

type ErrProviderAlreadyRegistered struct {
	Type reflect.Type
}

func (err ErrProviderAlreadyRegistered) Error() string {
	return fmt.Sprintf("Provider for %v already registered", err.Type.String())
}
