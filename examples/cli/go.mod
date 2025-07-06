module cli

go 1.21

replace github.com/chmenegatti/gocachex => ../..

require github.com/chmenegatti/gocachex v0.0.0-00010101000000-000000000000

require (
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.3.0 // indirect
)
