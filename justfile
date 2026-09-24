set windows-shell := ["pwsh", "-NoLogo", "-NoProfile", "-Command"]

ldflags := "-s -w"
exe := if os_family() == "windows" { "uc.exe" } else { "uc" }

[doc("list recipes")]
default:
    @just --list

[doc("build uc here")]
build:
    go build -ldflags="{{ldflags}}" -o {{exe}} ./cmd/uc

[doc("install uc to Go bin dir on PATH")]
install:
    go install -ldflags="{{ldflags}}" ./cmd/uc

[doc("run uc with args, e.g. just run best claude")]
run *args:
    go run ./cmd/uc {{args}}

[doc("run all tests")]
test:
    go test ./...

[doc("format code")]
fmt:
    gofmt -w .

[doc("fail if code is unformatted or vet finds issues")]
[windows]
lint:
    go vet ./...
    $f = gofmt -l .; if ($f) { $f; exit 1 }

[doc("fail if code is unformatted or vet finds issues")]
[unix]
lint:
    go vet ./...
    test -z "$(gofmt -l .)" || { gofmt -l .; exit 1; }

[doc("lint then test")]
check: lint test

[doc("delete build output")]
[windows]
clean:
    go clean
    Remove-Item -Force -ErrorAction Ignore {{exe}}

[doc("delete build output")]
[unix]
clean:
    go clean
    rm -f {{exe}}
