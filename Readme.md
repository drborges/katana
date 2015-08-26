# Katana

Dependency Injection package driven by providers

Katana follows a similar approach to `AngularJS`'s DI solution. In order to be able to inject a given type, you simply need to implement one of the available provider functions (`injector.Provide`, `injector.ProvideNew`, `injector.ProvideSingleton`) that knows how to build up an instance of a given type resolving its own dependencies by using katana's `Injector` passed to the provider.

Katana will detect any eventual `cyclic dependency` providing a dev friendly error as the result of a call to `injector.Resolve(...)`.

For a deep dive check the example below, though before you jump into it, have in mind the following note:

1. One `must` always pass a `pointer` of a variable to the `injector.Resolve(&dep)` call `even` if the dependency is already a pointer. That is required so that katana may set the resolved dependency value back into that variable.

### How Does It Work?

Lets say you have the following types each with their own dependencies:

```go
type Config struct {
	DatastoreURL string
	CacheTTL     int
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

In order to use `katana` to resolve instances of these types along with their dependencies you need to:
```go
// Grabs a new instance of katana.Injector
injector := katana.New()

// Global configuration instance
config := Config{
	DatastoreURL: "https://myawesomestartup.com/db",
	CacheTTL:     20000,
}

// Registers a provider for config. Katana will resolve dependencies to Config by
// setting them to this particular instance.
injector.Provide(config)

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
// 1. service1 == service2 since *AccountService provider is a singleton
// 2. db1 != db2 since their providers are not singleton ones and will always yield a new instance of the provided type.
// 3. cache1 != cache2 same reason as item 2.
// 4. config will point to the Config instance defined in the previous code block, since it was provided using Injector#Provide method.
injector.Resolve(&service1, &service2, &db1, &db2, &cache1, &cache2, &config)
```

# Thread-safety

In order to use `katana` in a `multi-thread` environment you can use `Injector.Clone()` to get a hold of a copy of the injector with no resolved instances cached. Every new provider registered in the new copy, won't be available to other copies of `katana.Injector`.

# Example: HTTP server

```go
// Assuming we have the injector instance from the example above ^
http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	var service *AccountService

	// There are two things to consider in the following code:
	//
	// 1. A copy of the injector is created so we can safely use it within a multi-thread environment.
	// 2. The instances of http.ResponseWriter and *http.Request are provided by the injector
	// 3. The instance of *AccountService is resolved as well as its dependencies
	injector.Clone().Provide(w, r).Resolve(&service)
})

log.Fatal(http.ListenAndServe(":8080", nil))
```

# Injecting Function Arguments

Katana allows you to inject arguments into functions as well <3

```go
allAccounts := injector.Inject(func(srv *AccountService, conf Config) ([]*Accounts, error) {
	return srv.Accounts()
})
```

`Injector#Inject` returns a closure holding all the resolved function arguments and when called returns a slice with the function outputs, being empty in case the function does not return any value.

```go
if result := allAccounts(); !result.Empty() {
	accounts, fetchAccountsErr := result[0], result[1]
}
```
