INSTALL_PREFIX ?= $(shell pwd)/build

BIN := \
	strava-heatmap-auth \
	strava-heatmap-proxy

.PHONY: all
all: $(BIN)

.PHONY: clean
clean:
	rm -f $(addprefix $(INSTALL_PREFIX)/, $(BIN))
	rm -d $(INSTALL_PREFIX)

%:
	go build -o $(INSTALL_PREFIX)/$@ cmd/$@/main.go
