package katana_test

import (
	"github.com/drborges/katana"
	"fmt"
)

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

func ExampleKatanaAPI() {
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

	var service1, service2 *AccountService
	injector.Resolve(&service1, &service2)

	fmt.Println("service1 == service2:", service1 == service2)
	fmt.Println("service1.Datastore.URL:", service1.Datastore.URL)
	fmt.Println("service1.Datastore.Cache.TTL:", service1.Datastore.Cache.TTL)

	// Output:
	// service1 == service2: true
	// service1.Datastore.URL: https://myawesomestartup.com/db
	// service1.Datastore.Cache.TTL: 20000
}
