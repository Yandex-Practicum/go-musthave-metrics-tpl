all: iter1 iter2 iter3

test:
	@echo "running tests"
	go test -coverprofile=coverage.out ./... && \
	go tool cover -html=coverage.out -o coverage.html && \
	open coverage.html

iter1:
	@echo "iter1 starting tests for first iteration"
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration1$ -binary-path=cmd/server/server

iter2:
	@echo "iter2 starting tests for second iteration"
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration2[AB]*$ \
            -source-path=. \
            -agent-binary-path=cmd/agent/agent

iter3:
	@echo "iter3 starting tests for third iteration"
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration3[AB]*$ \
            -source-path=. \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server

iter4:
	@echo "iter4 starting tests for fourth iteration"
	@SERVER_PORT="8080"; \
	ADDRESS="localhost:$${SERVER_PORT}"; \
	TEMP_FILE=$$(mktemp); \
	./metricstest -test.v -test.run=^TestIteration4$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=$${SERVER_PORT} \
		-source-path=.

iter5:
	@echo "iter5 starting tests for fifth iteration"
	@SERVER_PORT=$$(python3 -c "import random; print(random.randint(8080, 8090))"); \
	ADDRESS="localhost:$${SERVER_PORT}"; \
	TEMP_FILE=$$(mktemp); \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration5$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=$${SERVER_PORT} \
		-source-path=.


