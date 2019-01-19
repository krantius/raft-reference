build:
	@if [ -z raft ]; then rm raft; fi
	@GOOS=linux GOARCH=amd64 go build -o raft
docker: build
	@docker build --no-cache -f ./Dockerfile -t raft:latest .
run: docker
	@docker-compose down
	@docker-compose up