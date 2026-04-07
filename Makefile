.PHONY: run

run-users:
	cd ms-users && go run cmd/main.go

run-transactions:
	cd ms-transactions && go run cmd/main.go

mock-users:
	cd ms-users && rm -rf mocks && go install go.uber.org/mock/mockgen@v0.6.0 && go generate -tags=tool mockgen.go

mock-transactions:
	cd ms-transactions && rm -rf mocks && go install go.uber.org/mock/mockgen@v0.6.0 && go generate -tags=tool mockgen.go

test-users: mock-users
	cd ms-users && go fmt ./... && go test -count=1 -vet=all ./...

test-integration-users: mock-users
	cd ms-users && go fmt ./... && go test -count=1 -vet=all -tags=integration ./...

test-transactions: mock-transactions
	cd ms-transactions && go fmt ./... && go test -count=1 -vet=all ./...
	
test-integration-transactions: mock-transactions
	cd ms-transactions && go fmt ./... && go test -count=1 -vet=all -tags=integration ./...

test: test-users test-transactions

lint: get-linter
	cd ms-users && golangci-lint run
	cd ms-transactions && golangci-lint run

lint-fix: get-linter
	cd ms-users && golangci-lint run --fix
	cd ms-transactions && golangci-lint run --fix

get-linter:
	command -v golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(LINTER_VERSION)
