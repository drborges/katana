# Katana [![Build Status](https://travis-ci.org/drborges/katana.svg?branch=master)](https://travis-ci.org/drborges/katana)

Dependency Injection Driven By Provider Functions

## Brief Overview

katana approaches DI in a fairly simple manner. For each type that needs to be available for injection a provider function needs to be implemented and registered with an instance of `kanata.Injector`. Providers construct instances of a particular type a.k.a `Injectable` and can be of two types: `katana.TypeNew` and `katana.TypeSingleton`.

Provider functions are like [constructor functions](https://golang.org/doc/effective_go.html#composite_literals), they may take arguments representing the required dependencies to create an instance of that injectable and return a single value, the actual instance. Here is an example:

```go
func NewUserService(depA *DependencyA, depB *DependencyB) *UserService {
	return &UserService{depA, depB}
}
```

Once a provider is registered the corresponding injectable can be resolved and injected as dependency into other injectable providers, or even into arbitrary functions. Lets see how that translates into code:

```go
// Get an instance of katana's injector
injector := katana.New()

// Register the provider of *UserService with the injector
injector.ProvideNew(&UserService{}, NewUserService)

// Grab a new instance of *UserService with all its dependencies injected
var service *UserService
injector.Resolve(&service)
```

Katana will detect and panic upon any eventual `cyclic dependency` when resolving an injectable, providing the cyclic dependency graph so you can easily troubleshoot.

## Example

Lets say you have the following types each with their own dependencies:

```go
type Config struct {
	DatastoreURL string
	CacheTTL     int
	Debug        bool
}

type Cache struct {
	TTL int
}

type Datastore struct {
	Cache *Cache
	URL   string
}

type AccountService struct {
	Datastore *Datastore
}
```

A provider for each type of injectable is created and registered with a new instance of `katana.Injector`

```go
// Grabs a new instance of katana.Injector
injector := katana.New()

// Registers a provider for config. Katana will resolve dependencies on Config by
// setting them to this particular instance.
injector.Provide(Config{
	DatastoreURL: "https://myawesomestartup.com/db",
	CacheTTL:     20000,
})

// Registers a provider for *Cache whose result is never cached, which means
// different requests for an instance of *Cache will yield different instances.
injector.ProvideNew(&Cache{}, func(config Config) *Cache {
	return &Cache{config.CacheTTL}
})

// Registers a provider for *Datastore with all its dependencies (Config, *Cache)
// being resolved and injected into the provider function.
injector.ProvideNew(&Datastore{}, func(config Config, cache *Cache) *Datastore {
	return &Datastore{cache, config.DatastoreURL}
})

// Registers a singleton provider for *AccountService. The instance provided by
// a singleton provider is cached so further requests yield the same result.
injector.ProvideSingleton(&AccountService{}, func(db *Datastore) *AccountService {
	return &AccountService{db}
})
```

Finally you can get instances of the provided `injectables` with all their dependencies -- if any -- resolved:

```go
var service1, service2 *AccountService
var db1, db2 *Datastore
var cache1, cache2 *Cache
var config Config

// Katana allows you to resolve multiple instances on a single "shot"
// 
// Note that:
// 1. service1 == service2: *AccountService provider is a singleton
// 2. db1 != db2: *Datastore injectable is not singleton
// 3. cache1 != cache2: *Cache is not a singleton
// 4. config will point to the Config instance defined in the previous code block, since it was provided using Injector#Provide method.
injector.Resolve(&service1, &service2, &db1, &db2, &cache1, &cache2, &config)
```

# Thread-Safety

In order to use `katana` in a `multi-thread` environment you should use a copy of the injector per thread.

Copies of `katana.Injector` can be created using `Injector.Clone()`. This copy will have all the registered providers of the original injector and every new provider registered in the new copy will not be available to other copies of `katana.Injector`.

**Note** Singleton providers will still yield the same instances across different threads.

### Example: HTTP Server

Assuming we have the injector instance from the example above ^

```go
http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	var service *AccountService
	injector.Clone().Provide(w, r).Resolve(&service)
})

log.Fatal(http.ListenAndServe(":8080", nil))
```

# Injecting Function Arguments

Katana also allows you to inject arguments into functions (that is how it resolves the arguments of a injectable provider):

```go
fetchAllAccounts := injector.Inject(func(srv *AccountService, conf Config) ([]*Account, error) {
	if conf.Debug {
		return mocks.Accounts(), nil
	}
	return srv.Accounts()
})
```

`Injector#Inject` returns a closure holding all the resolved function arguments and when called returns a `katana.Output` with the function returning values.

```go
if result := fetchAllAccounts(); !result.Empty() {
	accounts, err := result[0], result[1]
}
```

# Contributing

# License