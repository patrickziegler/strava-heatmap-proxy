SERVICE_NAME := strava-heatmap-proxy
EXTENSION_NAME := strava-cookie-exporter

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
	GOPATH=$(HOME)/.local go install $(SERVICE_NAME).go
	mkdir -p $(HOME)/.config/systemd/user/
	cp $(SERVICE_NAME).service $(HOME)/.config/systemd/user/
	systemctl --user daemon-reload
	systemctl --user start $(SERVICE_NAME)
	systemctl --user enable $(SERVICE_NAME)

uninstall:
	rm -f $(addprefix $(HOME)/.local/bin/, $(SERVICE_NAME))
	rm $(HOME)/.config/systemd/user/$(SERVICE_NAME).service
	systemctl --user disable $(SERVICE_NAME)
	systemctl --user stop $(SERVICE_NAME)
	systemctl --user daemon-reload
