APP_AGENT=./bin/agent
APP_SERVER=./bin/server
METRICS_TEST=./metricstest-darwin-arm64

.PHONY: build build-agent build-server run run-agent run-server clean

build: build-agent build-server

build-agent:
	mkdir -p bin
	go build -o $(APP_AGENT) ./cmd/agent

build-server:
	mkdir -p bin
	go build -o $(APP_SERVER) ./cmd/server

run: build
	$(APP_SERVER) -a=localhost:8080 &
	$(APP_AGENT) -a=localhost:8080 -p=2 -r=10

run-server: build-server
	$(APP_SERVER) -a=localhost:8080

run-agent: build-agent
	$(APP_AGENT) -a=localhost:8080 #-p=2 -r=10

clean:
	rm -rf bin


# Тесты итераций
test: test-1 test-2 test-3 test-4 test-5

test-1: build-server
	$(METRICS_TEST) -test.v -test.run='^TestIteration1$$' \
		-binary-path=bin/server

test-2: build-server
	$(METRICS_TEST) -test.v -test.run='^TestIteration2$$' \
		-binary-path=bin/server

test-3: build-server
	$(METRICS_TEST) -test.v -test.run='^TestIteration3$$' \
		-binary-path=bin/server

test-4: build-server
	$(METRICS_TEST) -test.v -test.run='^TestIteration4$$' \
		-binary-path=bin/server \
		-source-path=. \
		-agent-binary-path=bin/agent \
		-server-port=8080

test-5: build
	$(METRICS_TEST) -test.v -test.run='^TestIteration5$$' \
		-binary-path=bin/server \
		-source-path=. \
		-agent-binary-path=bin/agent \
		-server-port=8080