
all: push

# TODO git describe --tags --abbrev=0
# get release from tags
RELEASE?=v0.2.2
GOOS?=linux
PREFIX?=cargo.caicloudprivatetest.com/caicloud/loadbalancer-controller

PKG=github.com/caicloud/loadbalancer-controller
REPO_INFO=$(shell git config --get remote.origin.url)
ifndef COMMIT
  COMMIT := git-$(shell git rev-parse --short HEAD)
endif

target=loadbalancer-controller

test:
	go list ./... | grep -v '/vendor/' | grep -v '/tests/' | xargs go test 

build: test
	GOOS=${GOOS} go build -i -v -o $(target) \
	-ldflags "-s -w -X $(PKG)/version.RELEASE=$(RELEASE) -X $(PKG)/version.COMMIT=$(COMMIT) -X $(PKG)/version.REPO=$(REPO_INFO)" \
	$(PKG)/cmd/controller

image: build
	docker build -t $(PREFIX):$(RELEASE) .

push: image
	docker push $(PREFIX):$(RELEASE)

debug: test
	go build -i -v -o $(target) \
	-ldflags "-s -w -X $(PKG)/version.RELEASE=$(RELEASE) -X $(PKG)/version.COMMIT=$(COMMIT) -X $(PKG)/version.REPO=$(REPO_INFO)" \
	$(PKG)/cmd/controller

	./$(target) --kubeconfig=${HOME}/.kube/config --debug --log-force-color

lint:
	cat .gofmt | xargs -I {} gofmt -w -s -d -r {}  $$(find . -name "*.go" -not -path "./vendor/*" -not -path ".git/*")
	gosimple $$(go list ./... | grep -v vendor)

tool:
	go get honnef.co/go/tools/cmd/gosimple
