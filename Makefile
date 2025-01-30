INSTALL_PREFIX ?= $(HOME)/.local

BIN := \
	strava-heatmap-auth \
	strava-heatmap-proxy

.PHONY: install
install: $(BIN)

.PHONY: clean
clean:
	rm -f $(addprefix $(INSTALL_PREFIX)/bin/, $(BIN))

%:
	GOPATH=$(INSTALL_PREFIX) go install cmd/$@/$@.go
