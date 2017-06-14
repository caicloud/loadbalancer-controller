target := loadbalancer-controller

.PHONY: controller
controller:
	go build -i -v -o $(target) github.com/caicloud/loadbalancer-controller/cmd/controller

.PHONY: image
image:
	GOOS=linux GOARCH=amd64 go build -i -v -o $(target) github.com/caicloud/loadbalancer-controller/cmd/controller 
	docker build -t $(target) -f Dockerfile .
	docker tag $(target) cargo.caicloudprivatetest.com/caicloud/$(target)
	docker push cargo.caicloudprivatetest.com/caicloud/$(target)
