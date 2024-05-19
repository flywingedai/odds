unit:
	go test -v -coverprofile=coverage.out `go list ./... | egrep -v '(/test|/mocks|/client|/cmd|/gateway)'`
	go tool cover -func coverage.out