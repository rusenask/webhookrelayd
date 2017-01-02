test:
	go test -v `go list ./... | egrep -v /vendor/`

build:
	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo  -ldflags  -'w' -o webhookrelayd .

image: build
	docker build -t webhookrelay/webhookrelayd -f Dockerfile .

push: 
	docker push webhookrelay/webhookrelayd

image-push: image push