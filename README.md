# strava-heatmap-proxy

This is a proxy server that allows fetching high resolution tiles from the [Strava Global Heatmap](https://www.strava.com/heatmap), which would usually require an account and login for accessing data below a certain zoom level.

## Getting started

### :hammer: Build and Install

With [git](https://git-scm.com/downloads) and [golang](https://go.dev/) available on your system, the following three steps are sufficient to build and install this tool.

```sh
git clone https://github.com/patrickziegler/strava-heatmap-proxy
cd strava-heatmap-proxy
GOBIN=~/.local/bin go install cmd/strava-heatmap-proxy/strava-heatmap-proxy.go
```

Then you can run the proxy with

```sh
strava-heatmap-proxy --config "<path-to-config>"
```

whereby the config is expected to be a json formatted file holding [Strava](https://www.strava.com/) login credentials

```json
{
  "Email": "...",
  "Password": "..."
}
```

### Usage

`strava-heatmap-proxy` will automatically login to Strava and subsequently set up a proxy server for `https://heatmap-external-a.strava.com/`.
Every request to `http://localhost:8080/` (or a different port that you can configure via `--port`) will then be extended with session cookies and forwarded to Strava.

This allows to use a [TMS](https://wiki.openstreetmap.org/wiki/TMS) file like shown below to define a new layer for fetching heatmap tiles from any kind of software that supports this (like [QMapShack](https://github.com/Maproom/qmapshack/wiki), [QGIS](https://www.qgis.org/en/site/) or [JOSM](https://josm.openstreetmap.de/)).

```xml
<TMS>
    <Layer idx="0">
        <ServerUrl>http://localhost:8080/tiles-auth/all/hot/%1/%2/%3.png</ServerUrl>
    </Layer>
</TMS>
```

This [Screenshot](https://i.imgur.com/WVHWyjR.jpeg) shows how it would look like in QMapShack.

### Additional Note

It is also possible to put the `CloudFront-*` parameters directly into the TMS file as shown below (`strava-heatmap-proxy` is printing them out on startup). In this case, it would not be necessary to keep the proxy running in the background, but you would need to update the file every once in a while as those parameters will expire after some time.

```xml

<TMS>
    <Layer idx="0">
        <ServerUrl>https://heatmap-external-a.strava.com/tiles-auth/all/hot/%1/%2/%3.png?v=19&amp;Key-Pair-Id=...&amp;Policy=...&amp;Signature=...</ServerUrl>
    </Layer>
</TMS>
```

## References

1. Discussion in https://github.com/bertt/wmts/issues/2 revealed the importance of `CloudFront-*` parameters
1. https://github.com/erik/strava-heatmap-proxy is following a similar approach but is designed to be a Cloudflare worker

## License

This project is licensed under the GPL - see the [LICENSE](LICENSE) file for details
