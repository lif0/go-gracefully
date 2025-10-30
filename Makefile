PKGS        := $(shell go list ./... | grep -v /vendor/)
COVERPKG    := $(PKGS)
COVERMODE   := atomic                 # safe for -race and cross-package
COVERFILE   := coverage.out
HTMLFILE    := coverage.html
TIMEOUT     := 5m
COUNT       := 1                      # disable cache for flaky / dev loops
SHUFFLE     := on

GO_TEST     := go test
GO_TOOL     := go tool

.DEFAULT_GOAL := test


TEST_FLAGS := -timeout $(TIMEOUT) -count=$(COUNT) -shuffle=$(SHUFFLE) -run=. \
	-cover -covermode=$(COVERMODE) -coverpkg=$(COVERPKG) -coverprofile=$(COVERFILE) ./...

test:
	$(GO_TEST) -race $(TEST_FLAGS)
	@$(MAKE) -s cover

cover:
	@$(GO_TOOL) cover -func=$(COVERFILE) | tail -n 1 | sed 's/^/total: /'

bench:
	$(GO_TEST) -run=^$$ -bench=. -benchmem -benchtime=1s $(PKGS)

clean:
	@rm -f $(COVERFILE) $(HTMLFILE)
