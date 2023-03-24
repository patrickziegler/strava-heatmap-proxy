# strava-heatmap-proxy

This software allows streaming high resolution [Strava Global Heatmap](https://www.strava.com/heatmap) tiles with clients like [QGIS](https://qgis.org/de/site/), [QMapShack](https://github.com/Maproom/qmapshack/wiki), [JOSM](https://josm.openstreetmap.de/) and many others without requiring them to be able to handle the Strava specific authentication and session management.

You can use the software either as [static file configurator](#using-the-static-file-configurator) or as [proxy server](#using-the-authentication-proxy) that handles the authentication on the fly.

## Getting started

### :hammer: Build and Install

With [git](https://git-scm.com/downloads), [golang](https://go.dev/) and [make](https://www.gnu.org/software/make/) available on your system, the following steps are sufficient to build and install the binaries `strava-heatmap-auth` and `strava-heatmap-proxy` to the given path `INSTALL_PREFIX`

```sh
git clone https://github.com/patrickziegler/strava-heatmap-proxy
cd strava-heatmap-proxy
INSTALL_PREFIX=~/.local/bin make
```

### :wrench: Configuration

The tools expect a config file located in `~/.config/strava-heatmap-proxy/config.json` containing [Strava](https://www.strava.com/) login credentials like shown below.
The option `--config <path>` can be used to specify a custom location.

```json
{
  "Email": "...",
  "Password": "..."
}
```

## Using the static file configurator

The tool `strava-heatmap-auth` will read from `stdin`, replace all occurences of `%CloudFront-Key-Pair-Id%`, `%CloudFront-Policy%` and `%CloudFront-Signature%` tokens with the values retrieved as authentication cookie and write the result to `stdout`.

This design allows us to provide a template [TMS](https://wiki.openstreetmap.org/wiki/TMS) file like the following:

```xml
<TMS>
  <Title>StravaHeatmap</Title>
  <MinZoomLevel>3</MinZoomLevel>
  <MaxZoomLevel>11</MaxZoomLevel>
  <Layer idx="0">
    <ServerUrl>https://heatmap-external-a.strava.com/tiles-auth/all/bluered/%1/%2/%3.png?v=19&amp;Key-Pair-Id=%CloudFront-Key-Pair-Id%&amp;Policy=%CloudFront-Policy%&amp;Signature=%CloudFront-Signature%</ServerUrl>
  </Layer>
</TMS>
```

And create the actual file with all `%CloudFront-*%` tokens replaced by their correct values with the following [pipeline](https://en.wikipedia.org/wiki/Pipeline_(Unix)):

```sh
cat StravaHeatmapAuth.tms.in | strava-heatmap-auth --config <path> | tee StravaHeatmapAuth.tms
```

Be aware that those parameters may expire after some time.

## Using the authentication proxy

The tool `strava-heatmap-proxy` will set up a local proxy server for `https://heatmap-external-a.strava.com/`.
Every request to `http://localhost:8080/` (or a different port that you can configure via `--port`) will then be extended with session cookies before being forwarded to Strava.

This design allows us to use a simple [TMS](https://wiki.openstreetmap.org/wiki/TMS) file like shown below for fetching high resolution heatmap tiles:

```xml
<TMS>
  <Title>StravaHeatmap</Title>
  <MinZoomLevel>3</MinZoomLevel>
  <MaxZoomLevel>11</MaxZoomLevel>
  <Layer idx="0">
    <ServerUrl>http://localhost:8080/tiles-auth/all/bluered/%1/%2/%3.png</ServerUrl>
  </Layer>
</TMS>
```

## References

1. Discussion in https://github.com/bertt/wmts/issues/2 revealed the meaning of `CloudFront-*` parameters
1. https://github.com/erik/strava-heatmap-proxy is following a similar approach but is designed to be a Cloudflare worker

## License

This project is licensed under the GPL - see the [LICENSE](LICENSE) file for details
