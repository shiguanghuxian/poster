BINARY_NAME=poster

default:
	@echo 'Usage of make: [ build | linux_build | windows_build | docker_build | docker_run | run | clean ]'

build: 
	go build -o ./bin/${BINARY_NAME} ./

linux_build: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY_NAME} ./

windows_build: 
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY_NAME}.exe ./

docker_build: linux_build
	docker build -t shiguanghuxian/${BINARY_NAME} .

docker_run: docker_build
	docker-compose up --force-recreate

run: build
	cd ./bin && ./${BINARY_NAME}

build_web:
	cd static && npm run build && cp -r dist ../tpls && cd ../tpls && ./compile.sh

clean: 
	rm -f ./bin/${BINARY_NAME}*
	rm -f ./bin/logs/*

.PHONY: default build linux_build windows_build docker_build docker_run run clean