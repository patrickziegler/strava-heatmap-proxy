SUBDIRS := strava-cookie-exporter strava-heatmap-proxy

.PHONY: $(SUBDIRS) all clean

all clean:
	$(foreach dir,$(SUBDIRS),$(MAKE) -C $(dir) $@;)
