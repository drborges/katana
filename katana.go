package katana

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	// TypeSingleton is an injectable whose provider is called at most once and its provided instance
	// is cached so that subsequent requests for that same type yield the same result.
	TypeSingleton = InjectableType("Singleton Dependency")
	// TypeNew is an injectable whose provider is called whenever an instance of the corresponding type
	// is requested. Different calls to the provider of this type of injectable will yield different instances
	TypeNew = InjectableType("New Instance Dependency")
)

// InjectableType describes the type of the registered injectable.
// It may assume two values: TypeSingleton or TypeNew
type InjectableType string

// Instance is an instance of a type provided by a registered provider function
type Instance interface{}

// Provider is a function that takes zero or more parameters and returns exactly one value
type Provider interface{}

// Callable wraps a provider function whose arguments have been resolved and injected
type Callable func() []Instance

// ValidateProvider validates whether or not a given provider is valid
// Providers must be callable a.k.a functions, taking zero or more arguments
// and returning exactly one value, the provided instance of the registered
// injectable.
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

// Injectable describes a particular type that can have instances injected as dependency
// provided by a registered provider function.
type Injectable struct {
	Type     InjectableType
	Provider Provider
}

// Trace keeps track of the current dependency graph under resolution watching out for
// cyclic dependencies.
// TODO merge trace and stack code into a separate file
type Trace struct {
	Stack
}

// Add adds a dependency type under resolution returning an ErrCyclicDependency in case
// of a cyclic dependency is detected, returns nil otherwise.
func (trace *Trace) Add(typ reflect.Type) (err error) {
	if trace.Contains(typ.String()) {
		defer trace.Reset()
		trace.Push(typ.String())
		return ErrCyclicDependency{trace}
	}
	trace.Push(typ.String())
	return err
}

// Injector is katana's DI implementation driven by typed provider functions.
//
// A provider function registered with the injector provides instances of a given type.
// Katana supports three types of providers:
//
// 1. Value Provider: For a given type it always provides a particular instance defined by the user.
// For detailed information see Injector#Provide method.
// 2. New Instance Provider: Always provides a new instance of the registered type, resolving any
// transitive dependency the instance may have.
// 3. Singleton Provider: Provides the same instance upon any request. The instance dependencies are
// resolved exactly once cached for further use.
type Injector struct {
	injectables map[reflect.Type]*Injectable
	instances   map[reflect.Type]Instance
	trace       *Trace
}

// New provides a new instance of katana's injector
func New() *Injector {
	return &Injector{
		injectables: make(map[reflect.Type]*Injectable),
		instances:   make(map[reflect.Type]Instance),
		trace:       &Trace{},
	}
}

// Clone returns a thread-safe copy of the injector
// This is particularly useful when used within web servers or any scenario where concurrency is present
func (injector *Injector) Clone() *Injector {
	newInjector := New()

	for t, p := range injector.injectables {
		newInjector.injectables[t] = p
	}

	return newInjector
}

func (injector *Injector) provide(injectable interface{}, injType InjectableType, p Provider) *Injector {
	typ := reflect.TypeOf(injectable)

	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Interface {
		typ = typ.Elem()
	}

	if _, registered := injector.injectables[typ]; registered {
		panic(ErrProviderAlreadyRegistered{typ})
	}

	if err := ValidateProvider(p); err != nil {
		panic(err)
	}

	injector.injectables[typ] = &Injectable{
		Type:     injType,
		Provider: p,
	}

	return injector
}

// ProvideNew provides a new instance of the registered injectable with all its dependencies (if any)
// resolved by calling their corresponding provider functions.
// Multiple calls to this method will yield a new result provided by the registered provider function
func (injector *Injector) ProvideNew(injectable interface{}, p Provider) *Injector {
	return injector.provide(injectable, TypeNew, p)
}

// ProvideSingleton provides the same instance of the registered injectable with all its dependencies (if any)
// resolved by calling their corresponding provider functions.
// The instance provided by the registered provider function is cached so that multiple calls to this
// method yield the same result.
func (injector *Injector) ProvideSingleton(injectable interface{}, p Provider) *Injector {
	return injector.provide(injectable, TypeSingleton, p)
}

// Provide is a short hand method that allows user defined instances to be injected as singletons
// Under the hood a singleton provider function is created for each user defined instance.
func (injector *Injector) Provide(instances ...interface{}) *Injector {
	for _, instance := range instances {
		injector.ProvideSingleton(instance, func(inst interface{}) Provider {
			return func() Instance { return inst }
		}(instance))
	}
	return injector
}

// Resolve resolves type references into actual instances provided by their corresponding provider
// functions if any.
// Each reference MUST be a pointer to the requested type, even if the requested type is already a
// pointer.
//
// For instance, lets say you need an instance of *Account type a.k.a an instance of a pointer to
// the Account type, assuming there is a provider of *Account, you can easialy then get an instance
// of it with its dependencies injected by doing the following:
//
// var acc *Account
// injector.Resolve(&acc)
func (injector *Injector) Resolve(refs ...interface{}) {
	for _, ref := range refs {
		val := reflect.ValueOf(ref)
		typ := val.Type()

		if typ.Kind() != reflect.Ptr {
			panic(ErrNoSuchPtr{typ})
		}

		typ = typ.Elem()

		// Checks whether there is a registered provider for the dependency
		injectable, registered := injector.injectables[typ]
		if !registered {
			panic(ErrNoSuchProvider{typ})
		}

		// Checks instances cache for previous resolved dependency in case it is a singleton one
		if injectable.Type == TypeSingleton {
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
		inst := injector.Inject(injectable.Provider)()[0]
		injector.trace.Pop()

		// Resolves the dependency with the new instance
		val.Elem().Set(reflect.ValueOf(inst))

		// Caches the instance in case the dependency is a singleton
		if injector.injectables[typ].Type == TypeSingleton {
			injector.instances[typ] = inst
		}
	}
}

// Inject resolves and injects all arguments of the given function 'fn' returning a Callable
// which is essentially a closure holding the resolved argument values.
func (injector *Injector) Inject(fn interface{}) Callable {
	val := reflect.ValueOf(fn)
	typ := val.Type()

	if typ.Kind() != reflect.Func {
		panic(ErrNoSuchCallable{typ})
	}

	args := make([]reflect.Value, typ.NumIn())
	for i := 0; i < typ.NumIn(); i++ {
		argVal := reflect.New(typ.In(i))
		arg := argVal.Interface()
		injector.Resolve(arg)
		args[i] = argVal.Elem()
	}

	callable := func() []Instance {
		outVals := val.Call(args)
		outs := make([]Instance, len(outVals))
		for i, val := range outVals {
			outs[i] = val.Interface()
		}
		return outs
	}

	return callable
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
	Trace *Trace
}

func (err ErrCyclicDependency) Error() string {
	return fmt.Sprintf("Cyclic dependency detected: [%v]", strings.Join(err.Trace.items, " -> "))
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
