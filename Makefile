iter1:
	echo "iter1 starting tests for first iteration"
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration1$ -binary-path=cmd/server/server

