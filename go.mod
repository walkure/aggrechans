module github.com/walkure/aggrechans

go 1.25

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/slack-go/slack v0.27.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// +heroku goVersion go1.25
// +heroku install ./webhook/
