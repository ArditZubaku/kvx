run:
	@go build -o kvx . && (trap 'go clean; exit' INT TERM EXIT; ./kvx) # runs clean no matter how the program exits

ping:
	@docker run -it --rm redis redis-cli -h host.docker.internal -p 7379 ping

cli:
	@docker run -it --rm redis redis-cli

deploy:
	@docker build -t kvx:latest .
	@docker rm -f kvx || true
	@docker run -d --rm \
		--name kvx \
		--network host \
		kvx:latest

benchmark:
	@docker run -it --rm \
		--network host \
		redis redis-benchmark \
		-n 10000 \
		-t ping_mbulk \
		-c 1 \
		-h 127.0.0.1 \
		-p 7379

redis:
	@docker run -it --rm --network host redis

benchmark-redis:
	@docker run -it --rm --network host redis redis-benchmark -n 10000 -t ping_mbulk -c 1 -h 127.0.0.1 -p 6379

# NOTE: This is super slow compared to the one above because of the networking
# redis:
# 	@docker run -it --rm -p 6379:6379 redis

# benchmark-redis:
# 	@docker run -it --rm redis redis-benchmark -n 10000 -t ping_mbulk -c 1 -h host.docker.internal -p 6379

attach:
	tmux a -t redis
