check:
	golangci-lint run  \
		--build-tags "${BUILD_TAG}" \
		--timeout=20m0s \
		--enable=gofmt \
		--enable=unconvert \
		--enable=unparam \
		--enable=asciicheck \
		--enable=misspell \
		--enable=revive \
		--enable=decorder \
		--enable=reassign \
		--enable=usestdlibvars \
		--enable=nilerr \
		--enable=gosec \
		--enable=exportloopref \
		--enable=whitespace \
		--enable=gocyclo \
		--enable=nestif \
		--enable=gochecknoinits \
		--enable=gocognit \
		--enable=funlen \
		--enable=forbidigo \
		--enable=godox \
		--enable=gocritic \
		--enable=gci \
		--enable=lll \
		--config=issues.exclude.yaml

fmt:
	gofumpt -l -w .;gci write ./

build:
	docker build -t buildf .