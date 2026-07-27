# SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
# SPDX-License-Identifier: CC0-1.0

.PHONY: build test coverage lint fmt check spdx clean

build:
	go build ./...
	go build -o daikin ./cmd/daikin

test:
	go test -race ./...

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run

fmt:
	gofmt -w .

check: fmt lint spdx

spdx:
	@fail=0; for f in $$(git ls-files '*.go' '*.md' '*.yml' '*.yaml' 'Makefile' '.gitignore' 'go.mod'); do \
		grep -q 'SPDX-License-Identifier:' $$f || { echo "missing SPDX header: $$f"; fail=1; }; \
	done; exit $$fail

clean:
	rm -f daikin coverage.out
