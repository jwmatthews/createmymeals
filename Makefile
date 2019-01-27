VERSION = $(shell git describe --dirty --tags --always)
REPO = github.com/jwmatthews/createmymeals
BUILD_PATH = $(REPO)/commands
PKGS = $(shell go list ./... | grep -v /vendor/)

export CGO_ENABLED:=0
ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif


all: format build/list_recipes

format:
	$(Q)go fmt $(PKGS)

dep:
	$(Q)dep ensure -v

dep-update:
	$(Q)dep ensure -update -v

clean:
	$(Q)rm build/list_recipes*

.PHONY: all test format dep clean

install:
	$(Q)go install $(BUILD_PATH)

release_x86_64 := \
	build/list_recipes-$(VERSION)-x86_64-linux-gnu \
	build/list_recipes-$(VERSION)-x86_64-apple-darwin

release: clean $(release_x86_64) $(release_x86_64:=.asc)

build/list_recipes-%-x86_64-linux-gnu: GOARGS = GOOS=linux GOARCH=amd64
build/list_recipes-%-x86_64-apple-darwin: GOARGS = GOOS=darwin GOARCH=amd64

build/%:
	$(Q)$(GOARGS) go build -o $@ $(BUILD_PATH)

build/%.asc:
	$(Q){ \
	default_key=$$(gpgconf --list-options gpg | awk -F: '$$1 == "default-key" { gsub(/"/,""); print toupper($$10)}'); \
	git_key=$$(git config --get user.signingkey | awk '{ print toupper($$0) }'); \
	if [ "$${default_key}" = "$${git_key}" ]; then \
		gpg --output $@ --detach-sig build/$*; \
		gpg --verify $@ build/$*; \
	else \
		echo "git and/or gpg are not configured to have default signing key $${default_key}"; \
		exit 1; \
	fi; \
	}

.PHONY: install release_x86_64 release





