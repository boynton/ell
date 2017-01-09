GOPATH=$(HOME)
PKG=github.com/boynton/ell
all:
	go install $(PKG)/cmd/ell

test:
	go test $(PKG)

clean:
	go clean $(PKG)/...
	rm -rf *~

check:
	(cd $(GOPATH)/src/$(PKG); go vet $(PKG); go fmt $(PKG))
