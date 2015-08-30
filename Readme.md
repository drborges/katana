# Katana [![Build Status](https://travis-ci.org/drborges/katana.svg?branch=master)](https://travis-ci.org/drborges/katana)

Dependency Injection Driven By Constructor Functions

## Brief Overview

katana approaches DI in a fairly simple manner. For each type that needs to be available for injection -- a.k.a `injectable` -- a [constructor function](https://golang.org/doc/effective_go.html#composite_literals) needs to be registered with an instance of `kanata.Injector`.

```go
func NewUserService(depA *DependencyA, depB *DependencyB) *UserService {
	return &UserService{depA, depB}
}
```

Once a provider is registered the corresponding injectable can be resolved and injected as dependency into other injectable providers, or even into arbitrary functions. Lets see how that translates into code:

```go
// Get an instance of katana's injector
injector := katana.New()

// Register the following instances as injectables
depA, depB := &DependencyA{}, &DependencyB{}

// Register a constructor function to provide instances of *UserService
injector.Provide(depA, depB).ProvideNew(&UserService{}, NewUserService)

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

A constructor function for each type of injectable is created and registered with a new instance of `katana.Injector`

```go
// Grabs a new instance of katana.Injector
injector := katana.New()

// Registers the given instance of Config to be provided as a singleton injectable
injector.Provide(Config{
	DatastoreURL: "https://myawesomestartup.com/db",
	CacheTTL:     20000,
})

// Registers a constructor function that always provides a new instance of *Cache
injector.ProvideNew(&Cache{}, func(config Config) *Cache {
	return &Cache{config.CacheTTL}
})

// Registers a constructor function that always provides a new instance of *Datastore
// resolving its dependencies -- Config and *Cache -- as part of the process
injector.ProvideNew(&Datastore{}, func(config Config, cache *Cache) *Datastore {
	return &Datastore{cache, config.DatastoreURL}
})

// Registers a constructor function that lazily provides the same instance of *AccountService
// resolving its dependencies -- *Datastore -- as part of the process.
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

# Injecting Interfaces

In Go there is no way to pass in types as function arguments and types are derived through reflection from actual instances.

In addition to that an interface cannot be instantiated either, which makes things a little trick when writing generic code like a DI container.

Katana solution for injecting into interface references might seem a bit strange at first, but you'll get used :)

Lets say we want to provide a particular implementation of `http.ResponseWriter` to be injected as dependency. With `katana` you would do the following:

```go
injector.ProvideAs((*http.ResponseWriter)(nil), writer)
```

`(*http.ResponseWriter)(nil)` is how we tell katana to treat `writer` as a `http.ResponseWriter` rather than its actual underlying implementation `*http.response`.

With that whenever a dependency to `http.ResponseWriter` is detected, it will be resolved as that particular `writer` instance.

# Thread-Safety

In order to use `katana` in a `multi-thread` environment you should use a copy of the injector per thread.

Copies of `katana.Injector` can be created using `Injector.Clone()`. This copy will have all the registered providers of the original injector and every new provider registered in the new copy will not be available to other copies of `katana.Injector`.

**Note** Singleton providers will still yield the same instances across different threads.

### Example: HTTP Server

Assuming we have the injector instance from the example above ^

```go
http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	var service *AccountService
	injector.Clone().
		ProvideAs((*http.ResponseWriter)(nil), w).
		Provide(r).Resolve(&service)
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

Please feel free to submit issues, fork the repository and send pull requests!

When submitting an issue, please include a test function that reproduces the issue, that will help a lot to reduce back and forth :~

# License

The MIT License (MIT)

Copyright (c) 2015 Diego da Rocha Borges

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

