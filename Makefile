GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
BUILD_DIR = dist/${GOOS}_${GOARCH}
GENERATED_CONF := pkg/config/conf.gen.go

ifeq ($(GOOS),windows)
OUTPUT_PATH = ${BUILD_DIR}/baton-successfactors.exe
else
OUTPUT_PATH = ${BUILD_DIR}/baton-successfactors
endif

# Set the build tag conditionally based on ENABLE_LAMBDA
ifdef BATON_LAMBDA_SUPPORT
	BUILD_TAGS=-tags baton_lambda_support
else
	BUILD_TAGS=
endif

.PHONY: build
build: ${GENERATED_CONF}
	go build ${BUILD_TAGS} -o ${OUTPUT_PATH} ./cmd/baton-successfactors

$(GENERATED_CONF): pkg/config/config.go go.mod
	@echo "Generating $(GENERATED_CONF)..."
	go generate -tags=generate ./pkg/config

generate: $(GENERATED_CONF)

.PHONY: update-deps
update-deps:
	go get -d -u ./...
	go mod tidy -v
	go mod vendor

.PHONY: add-dep
add-dep:
	go mod tidy -v
	go mod vendor

.PHONY: lint
lint:
	golangci-lint run --out-format=colored-line-number --timeout=3m

MOCK_DIR ?= .mock
MOCK_BIN = $(MOCK_DIR)/bin

.PHONY: mock
# Start the mock SuccessFactors server and print the BATON_* exports for the connector.
mock:
	@mkdir -p $(MOCK_BIN)
	go build -o $(MOCK_BIN)/mock-successfactors ./test/mock-successfactors
	@[ -f $(MOCK_DIR)/cert.pem ] || openssl req -x509 -newkey rsa:2048 -nodes -days 1 -subj /CN=mock \
		-keyout $(MOCK_DIR)/key.pem -out $(MOCK_DIR)/cert.pem 2>/dev/null
	@echo "Run these in a second terminal, then run: go run ./cmd/baton-successfactors"
	@./test/mock-env.sh $(MOCK_DIR)
	@echo
	$(MOCK_BIN)/mock-successfactors -cert $(MOCK_DIR)/cert.pem

.PHONY: test-mock
# Run the connector and the baton CLI against the mock (the CI mock-test job).
# BATON must be a baton CLI with the same version as baton-sdk in go.mod.
test-mock:
	@mkdir -p $(MOCK_BIN)
	go build -o $(MOCK_BIN)/mock-successfactors ./test/mock-successfactors
	go build -o $(MOCK_BIN)/baton-successfactors ./cmd/baton-successfactors
	MOCK=$(abspath $(MOCK_BIN))/mock-successfactors \
	BATON_SUCCESSFACTORS=$(abspath $(MOCK_BIN))/baton-successfactors \
	WORKDIR=$(MOCK_DIR)/run ./test/verify-mock.sh
