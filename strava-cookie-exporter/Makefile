EXTENSION_NAME := strava-cookie-exporter

OUTPUT := \
	$(EXTENSION_NAME).zip

.PHONY: all clean

all: $(OUTPUT)

$(EXTENSION_NAME).zip:
	7z a $@ ./* -x!Makefile

clean:
	rm -f $(OUTPUT)
