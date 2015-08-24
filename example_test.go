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
	injector.Provide(config)

	injector.ProvideNew(&Cache{}, func(config Config) *Cache {
		return &Cache{config.CacheTTL}
	})

	// The provider below provides a new instance of *Datastore whenever it is requested. Its resolved instance is never cached and subsequent resolution calls of the same type will always call the provider function.
	injector.ProvideNew(&Datastore{}, func(config Config, cache *Cache) *Datastore {
		return &Datastore{cache, config.DatastoreURL}
	})

	// A singleton provider is called at most once and its resolved value is then cached so further requests yield the same result.
	injector.ProvideSingleton(&AccountService{}, func(db *Datastore) *AccountService {
		return &AccountService{db}
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
