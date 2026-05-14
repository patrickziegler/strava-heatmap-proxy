# strava-heatmap-proxy

This software allows streaming high resolution [Strava Global Heatmap](https://www.strava.com/maps/global-heatmap) tiles with clients like [QGIS](https://qgis.org/de/site/), [QMapShack](https://github.com/Maproom/qmapshack/wiki), [JOSM](https://josm.openstreetmap.de/) and many others without requiring them to be able to handle the Strava specific authentication and session management.

To do so, you need the following two pieces:
1. The [strava-cookie-exporter](#export-cookies) browser extension to export the necessary cookies as json file
1. The [strava-heatmap-proxy](#run-the-proxy-with-docker) server which adds the necessary cookies to your requests before redirecting them to Strava

## Usage

### Export cookies

The `strava-cookie-exporter` browser extension is available in the [Firefox add-on store](https://addons.mozilla.org/de/firefox/addon/strava-cookie-exporter/) and the [Chrome web store](https://chromewebstore.google.com/detail/strava-cookie-exporter/apkhbbckeaminpphaaaabpkhgimojlhk).

With this extension installed, you can:
- Use your browser to login and navigate to the [Strava Global Heatmap](https://www.strava.com/maps/global-heatmap)
- Use the extension to export the relevant cookies as json file

The exported file is needed for running [strava-heatmap-proxy](#run-the-proxy-with-docker).

> [!NOTE]
> The exported cookies have an expiration period of 24 hours, but you'll most likely need to export them only once as the proxy will automatically refresh expired cookies as long as the session is valid (the exact duration of that is unkown right now, but it seems to be several months at least).

### Run the proxy with Docker

You can use the prebuilt [Docker image](https://hub.docker.com/r/patrickziegler/strava-heatmap-proxy) for running a local instance of the proxy server in your terminal:

```sh
export LOCAL_PORT=9192 # optional, in case you want a different port than 8080 (the default)
docker run --pull=always --rm -p ${LOCAL_PORT:-8080}:8080 -v ${HOME}/.config/strava-heatmap-proxy:/home/nonroot/.config/strava-heatmap-proxy:ro docker.io/patrickziegler/strava-heatmap-proxy:latest
```
As a more simple alternative, you can use the provided [docker-compose.yml](docker-compose.yml):
```sh
export LOCAL_PORT=9192 # optional, see above
docker compose up -d
```

This will set up a local proxy server for `https://content-a.strava.com/`.
Every request to `http://localhost:8080/` will then be extended with session cookies before being forwarded to Strava.

By default, the necessary cookies are expected to be found in the file `${HOME}/.config/strava-heatmap-proxy/strava-cookies.json` (as created with the [strava-cookie-exporter](#export-cookies) extension).

### Configure your TMS client

To use this with your GIS software of choice, just define a simple [TMS](https://wiki.openstreetmap.org/wiki/TMS) layer like shown below that fetches high resolution heatmap tiles:

```xml
<TMS>
  <Title>StravaGlobalHeatmap</Title>
  <MinZoomLevel>5</MinZoomLevel>
  <MaxZoomLevel>16</MaxZoomLevel>
  <Layer idx="0">
    <ServerUrl>http://localhost:8080/identified/globalheat/all/bluered/{z}/{x}/{y}.png?v=19</ServerUrl>
  </Layer>
</TMS>
```

The `ServerUrl` can hold other elements than `all` and `bluered`. If you want to filter for certain activities or select different colorschemes you may check [this page](https://tjasz.github.io/heatmap/) for how to set it up accordingly.

### Advanced configuration for web clients

Web clients like [gpx.studio](https://gpx.studio/) need to be whitelisted via the `--allow-origins '["https://gpx.studio"]'` option.
Otherwise the browser would reject the responses due to a violation of the [same-origin policy](https://en.wikipedia.org/wiki/Same-origin_policy).

With the option in place, you can add the custom layer `http://localhost:8080/identified/globalheat/all/bluered/{z}/{x}/{y}.png?v=19` for accessing the heatmap.

### Accessing the personal heatmap

By default, `strava-heatmap-proxy` will grant acess to the global heatmap. If you want to access your personal heatmap instead, you can run the proxy like this
```sh
strava-heatmap-proxy --target "https://personal-heatmaps-external.strava.com/tiles/<USER_ID>/"
```
And then use the followng URL in your layer configuration
```sh
http://localhost:8080/bluered/{z}/{x}/{y}.png?missing=empty&amp;filter_type=all&amp;include_everyone=true&amp;include_followers_only=true&amp;include_only_me=true&amp;respect_privacy_zones=false&amp;include_commutes=false
```
Whereas the `<USER_ID>` is a unique identifier you can find out by using the developer tools while accessing the personal heatmap in your browser. We will make this id visible via the [strava-cookie-exporter](#export-cookies) in a future release.

## Screenshot

This is how the result might look like in [QMapShack](https://github.com/Maproom/qmapshack/wiki):

![screenshot of tms client](https://addons.mozilla.org/user-media/previews/full/351/351337.png)

## References

1. Discussion in https://github.com/bertt/wmts/issues/2 revealed the meaning of `CloudFront-*` tokens
1. https://github.com/erik/strava-heatmap-proxy was following a similar approach but is designed to be a Cloudflare worker

## License

This project is licensed under the GPL - see the [LICENSE](LICENSE) file for details
