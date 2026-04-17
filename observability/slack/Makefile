app_name := $(shell basename "$$PWD")

.PHONY: build
build:
	@ go build -o $(app_name)

.PHONY: docker
docker:
	@ podman image build --tag $(app_name):latest .

.PHONY: clean
clean:
	@ rm -f $(app_name)
