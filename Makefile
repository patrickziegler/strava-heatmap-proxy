SERVICE_NAME := strava-heatmap-proxy
EXTENSION_NAME := strava-cookie-exporter
INSTALL_PREFIX := $(HOME)/.local

OUTPUT := \
	$(SERVICE_NAME) \
	$(EXTENSION_NAME).zip

.PHONY: all clean

all: $(OUTPUT)

$(SERVICE_NAME):
	go build $@.go

$(EXTENSION_NAME).zip:
	7z a $@ ./$(EXTENSION_NAME)/*

clean:
	rm -f $(OUTPUT)

install:
	GOPATH=$(INSTALL_PREFIX) go install $(SERVICE_NAME).go

uninstall:
	rm -f $(addprefix $(INSTALL_PREFIX)/bin/, $(SERVICE_NAME))
