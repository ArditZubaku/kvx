run:
	@go build -o kvx . && (trap 'go clean; exit' INT TERM EXIT; ./kvx) # runs clean no matter how the program exits

lint:
	@GOOS=linux GOARCH=amd64 golangci-lint run ./...

build:
	@docker build -t kvx:latest .

ping:
	@docker run -it --rm redis redis-cli -h host.docker.internal -p 7379 ping

send-pipeline:
	@docker run -it --rm \
		--network host \
		busybox sh -c "(printf '*1\r\n\$$4\r\nPING\r\n*3\r\n\$$3\r\nSET\r\n\$$1\r\nk\r\n\$$1\r\nv\r\n*2\r\n\$$3\r\nGET\r\n\$$1\r\nk\r\n';) | nc 127.0.0.1 7379"

test:
	@DOCKER_BUILDKIT=1 docker build --progress=plain --target test -t app-tests .

aof:
	@docker run -i --rm \
		--network host \
		redis redis-cli -h 127.0.0.1 -p 7379 <<< "SET k1 v1"$$'\n'"SET k2 v2"$$'\n'"SET k3 v4"$$'\n'"SET k3 v3"$$'\n'"BGREWRITEAOF"

aof-verify:
	@docker cp kvx:/app/kvx-master.aof - | docker run -i --entrypoint sh --rm redis -c "tar -xC /tmp && redis-check-aof /tmp/kvx-master.aof"

logs:
	@docker logs kvx

logs-live:
	@docker logs -f kvx

cli:
	@docker run -it --rm \
		--network host \
		redis redis-cli \
		-h 127.0.0.1 \
		-p 7379

deploy:
	@docker build -t kvx:latest .
	@docker rm -f kvx || true
	@docker run -d --rm \
		--name kvx \
		--network host \
		kvx:latest

exec:
	@docker exec -it kvx sh

kill:
	@docker stop kvx
	@docker rm -f kvx

benchmark-single-connection:
	@docker run -it --rm \
		--network host \
		redis redis-benchmark \
		-n 10000 \
		-t ping_mbulk \
		-c 1 \
		-h 127.0.0.1 \
		-p 7379

benchmark-500-concurrent:
	@docker run -it --rm \
		--network host \
		redis redis-benchmark \
		-n 10000 \
		-t ping_mbulk \
		-c 500 \
		-h 127.0.0.1 \
		-p 7379

redis:
	@docker run -it --rm --network host redis

benchmark-redis-single-connection:
	@docker run -it --rm \
		--network host \
		redis redis-benchmark \
		-n 10000 \
		-t ping_mbulk \
		-c 1 \
		-h 127.0.0.1 \
		-p 6379

benchmark-redis-500-concurrent:
	@docker run -it --rm \
		--network host \
		redis redis-benchmark \
		-n 10000 \
		-t ping_mbulk \
		-c 500 \
		-h 127.0.0.1 \
		-p 6379

# NOTE: This is super slow compared to the one above because of the networking
# redis:
# 	@docker run -it --rm -p 6379:6379 redis

# benchmark-redis:
# 	@docker run -it --rm redis redis-benchmark -n 10000 -t ping_mbulk -c 1 -h host.docker.internal -p 6379

attach:
	tmux a -t redis
