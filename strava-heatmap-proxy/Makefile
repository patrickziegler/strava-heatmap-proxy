SERVICE_NAME ?= strava-heatmap-proxy
SERVICE_VERSION := $(shell cat version)
LOCAL_PORT ?= 9192

OUTPUT := \
	$(SERVICE_NAME)

.PHONY: all clean install uninstall formatcheck container

all: $(OUTPUT)

$(SERVICE_NAME):
	go build ./cmd/$@

clean:
	rm -f $(OUTPUT)

install:
	GOPATH=$(HOME)/.local go install -ldflags "-X main.version=${SERVICE_VERSION}" ./cmd/$(SERVICE_NAME)
	mkdir -p $(HOME)/.config/systemd/user/
	cp ./etc/systemd/$(SERVICE_NAME).service $(HOME)/.config/systemd/user/
	systemctl --user daemon-reload
	systemctl --user restart $(SERVICE_NAME)
	systemctl --user enable $(SERVICE_NAME)

uninstall:
	rm -f $(addprefix $(HOME)/.local/bin/, $(SERVICE_NAME))
	rm -f $(HOME)/.config/systemd/user/$(SERVICE_NAME).service
	systemctl --user disable $(SERVICE_NAME) || true
	systemctl --user stop $(SERVICE_NAME) || true
	systemctl --user daemon-reload

formatcheck:
	UNFORMATTED=$$(gofmt -l .) && \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "The following files are not properly formatted:"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	fi

container:
	docker build -t $(SERVICE_NAME) .
	docker run --rm -p $(LOCAL_PORT):8080 -v $(HOME)/.config/$(SERVICE_NAME):/home/nonroot/.config/$(SERVICE_NAME):ro $(SERVICE_NAME) || true
