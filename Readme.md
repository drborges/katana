# Katana

Dependency Injection package driven by providers

Katana follows a similar approach to `AngularJS`'s DI solution. In order to be able to inject a given type, you simply need to implement one of the available provider functions (`injector.ProvideValues`, `injector.ProvideNew`, `injector.ProvideSingleton`) that knows how to build up an instance of a given type resolving its own dependencies by using katana's `Injector` passed to the provider.

Katana will detect any eventual `cyclic dependency` providing a dev friendly error as the result of a call to `injector.Resolve(...)`.

For a deep dive check the example below, though before you jump into it, have in mind the following note:

1. One `must` always pass a `pointer` of a variable to the `injector.Resolve(&dep)` call `even` if the dependency is already a pointer. That is required so that katana may set the resolved dependency value back into that variable.

### Commented Example:

```go
package katana_test

import (
	"github.com/drborges/katana"
	"log"
	"testing"
)

type Config struct {
	DatastoreURL string
	CacheTTL     int
	// Omitted code
}

type Cache struct {
	TTL int
	// Omitted code
}

type Datastore struct {
	Cache *Cache
	URL   string
	// Omitted code
}

type AccountService struct {
	Datastore *Datastore
	// Omitted code
}

func TestKatanaAPI(t *testing.T) {
	injector := katana.New()

	config := Config{
		DatastoreURL: "https://myawesomestartup.com/db",
		CacheTTL:     20000,
	}

	// Registers a provider for the given instance. This type of provider returns the same object in case of registering a pointer
	// or a copy of the object in case of a value
	injector.ProvideValue(config)

	injector.ProvideNew(&Cache{}, func(injector *katana.Injector) (katana.Instance, error) {
		var config Config
		
		// Resolves the config dependency required to set up a new Cache object
		// Note that we are passing a pointer of config to the resolve call. You must always pass the address of the variable where the dependency will be resolved to, even if the dependency is already a pointer
		if err := injector.Resolve(&config); err != nil {
			return nil, err
		}

		return &Cache{config.CacheTTL}, nil
	})

	// The provider below provides a new instance of *Datastore whenever it is requested. Its resolved instance is never cached and subsequent resolution calls of the same type will always call the provider function.
	injector.ProvideNew(&Datastore{}, func(injector *katana.Injector) (katana.Instance, error) {
		var config Config
		var cache *Cache
		if err := injector.Resolve(&config, &cache); err != nil {
			return nil, err
		}

		return &Datastore{cache, config.DatastoreURL}, nil
	})

	// A singleton provider is called at most once and its resolved value is then cached so further requests yield the same result.
	injector.ProvideSingleton(&AccountService{}, func(injector *katana.Injector) (katana.Instance, error) {
		var db *Datastore
		if err := injector.Resolve(&db); err != nil {
			return nil, err
		}
		return &AccountService{db}, nil
	})

	var service1, service2 *AccountService

	// service1 and service2 will hold the same value since the provider for *AccountService is a singleton one
	if err := injector.Resolve(&service1, &service2); err != nil {
		log.Fatal(err)
	}

	// service dependencies resolved, enjoy! ;)

	if service1 != service2 {
		t.Fatal("Expected %+v == %+v", service1, service2)
	}

	if service1.Datastore.URL != config.DatastoreURL {
		t.Fatal("Expected datastore URL to be %v. Got %+v", config.DatastoreURL, service1.Datastore.URL)
	}

	if service1.Datastore.Cache.TTL != config.CacheTTL {
		t.Fatal("Expected cache TTL to be %v. Got %+v", config.CacheTTL, service1.Datastore.Cache.TTL)
	}
}
```

# Thread-safety

In order to use katana in a multi-thread environment (across multiple goroutines) you can use `Injector.Clone()` to get a hold of a copy of the injector that you can safely use without worring about getting it messed up by other threads.

# Example: HTTP server

```go

// Assuming we have the injector instance from the example above ^
http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
    var service *AccountService
    // There are two things to consider in the following code:
    // 1. We get a copy of the injector so that we can use it in a multi-thread environment.
    // 2. We register one provider for http.ResponseWriter and another one for *http.Request so we can inject them as dependencies into any type requesting them (though in this case we don't do anything with them, will work on a better example ><)
	if err := injector.Clone().ProvideValue(w, r).ResolveNew(&service); err != nil {
	    log.Fatal(err)
	}
	
	// do something with service...
})

log.Fatal(http.ListenAndServe(":8080", nil))
```

# Injecting Function Arguments

Katana provides a way to inject arguments into functions as well <3

```go
allAccounts, err := injector.Inject(func(srv *AccountService, conf Config) ([]*Accounts, error) {
    return srv.Accounts()
})

// allAccounts is the resulting function holding all the resolved arguments
// subsequent calls to allAccounts won't have to resolve the arguments again
result := allAccounts()
accounts, fetchAccountsErr := result[0], result[1]
```
