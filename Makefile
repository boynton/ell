GOPATH=$(HOME)
PKG=github.com/boynton/ell
all:
	go install $(PKG)/cmd/ell

clean:
	go clean $(PKG)/...
	rm -rf *~

check:
	(cd $(GOPATH)/src/$(PKG); go vet $(PKG); go fmt $(PKG))
