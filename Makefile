worker.test:
	go test -tags nautilus ./worker -c

.PHONY: worker.test-nautilus-docker
worker.test-nautilus-docker:
	docker build -t go-ds-rados-builder .
	docker run -v ${PWD}/out:/out go-ds-rados-builder
