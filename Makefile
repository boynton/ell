NAME=ell
PKG=github.com/boynton/$(NAME)
CMD=$(PKG)/cmd/$(NAME)
VERSION="v1.0.0"
all:
	go build -ldflags "-X $(PKG).Version=`git describe --tag`" -o bin/$(NAME) $(CMD)

release::
	git tag -a $(VERSION) -m "$(VERSION)"
	git push origin $(VERSION)

test:
	go test $(PKG)

clean:
	go clean $(PKG)/...
	rm -rf *~

proper::
	go fmt $(PKG)
	go vet $(PKG)
	go fmt $(CMD)
	go vet $(CMD)
