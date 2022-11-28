PROJECTNAME = go

## linux: 编译打包linux
.PHONY: linux
linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(RACE) -o app ./main.go

## win: 编译打包win
.PHONY: win
win:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(RACE) -o app.exe ./main.go

## mac: 编译打包mac
.PHONY: mac
mac:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(RACE) -o app ./main.go

docker:
	make linux
	docker build -t energy .

tar:
	docker save energy -o ./energy.tar

build:
	make docker
	make tar

run:
	make docker
	docker run energy /home/admin/atec_project/app