all:
	@make dep
	@make build
	@make test

build:
	go build github.com/anticpp/gomqtt

test:
	go test github.com/anticpp/gomqtt -v


run:
	go build broker/*
	./main

sync:
	rsync -avz ./ root@www.supergui.cn:/root/gowork/src/github.com/anticpp/gomqtt --delete --exclude=.git
