build:
	rm raft
	GOOS=linux GOARCH=amd64 go build -o raft
docker: build
	docker build --no-cache -f ./Dockerfile -t raft:latest .
