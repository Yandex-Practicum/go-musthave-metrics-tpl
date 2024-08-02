all: iter1 iter2

iter1:
	echo "iter1 starting tests for first iteration"
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration1$ -binary-path=cmd/server/server

iter2:
	echo "iter2 starting tests for second iteration"
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