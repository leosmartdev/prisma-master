.PHONY: dist golang-ci-install pb-go-tag-bson-install godoc release docs serve-docs

all: compile

ci: ci-init init protobuf test version compile

ci-init:
	rsyslogd

init: mod golangci-install pb-go-tag-bson-install

mod:
	go mod download
	go get -u github.com/golang/protobuf/protoc-gen-go@v1.3.0

golangci-install:
	# curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b ${GOPATH}/bin v1.15.0

pb-go-tag-bson-install:
	go get -u github.com/arkavo-com/pb-go-tag-bson

protobuf:
	make/make-protobuf

demo:
	make/make-demo

compile:
	( cd tms ; go install ./... )

version:
	make/go-version

test:
	( cd tms ; go test ./... -v -count=1 ; cd ../gogroup ; go test ./... -v -count=1)

cover:
	( cd gogroup ; go test ./...  -covermode=atomic -coverprofile ../gogroup.txt ; cat ../gogroup.txt  > ../.cover.txt )
	( cd tms ; go test ./...  -covermode=atomic -coverprofile ../coverage.txt ; cat ../coverage.txt | grep -v "mode: atomic" >> ../.cover.txt )
	go tool cover -func=.cover.txt
	cat coverage.txt | grep -v "mode: atomic" >> gogroup.txt
	mv gogroup.txt coverage.txt

lint:
	( cd tms ; golangci-lint run ./... --deadline=5m --disable-all --tests=false --enable=errcheck --enable=unused --enable=deadcode --enable=gosec --issues-exit-code=0)

integration:
	export PYTHONWARNINGS="ignore:Unverified HTTPS request"
	tests/acceptance/test-tmsd
	( cd tests/acceptance ; ./run-tests $(tests))

acceptance: integration

testbed: dist
	make/testbed-deploy

dist: clean init protobuf version compile
	make/deb-release

stable: dist
	make/aws-deploy

release: dist
	make/make-release

docs:
	cd doc && $(MAKE) $@

serve-docs:
	cd doc && $(MAKE) serve

godoc:
	mkdir -p  build/
	-nohup godoc -http :8080 & \
	sleep 5; \
	serverPID=$$!; \
	cd build && (wget --quiet -e robots=off -m http://localhost:8080/pkg/prisma/ --domains localhost --include-directories="/pkg,/lib"); \
	kill $$serverPID
	cd build && mv localhost:8080 godoc

clean:
	rm -rf build
	rm -rf dist
	-rm tms/cmd/tools/tdemo/prisma
	-rm tests/acceptance/prisma
	find tms -name "*.pb.go" -delete
	cd doc && $(MAKE) $@
