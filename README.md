# strava-heatmap-proxy

## Getting started

### :hammer: Build and Install

With [git](https://git-scm.com/downloads) and [golang](https://go.dev/) available on your system, the following three steps are sufficient to build and install this tool

```sh
git clone https://github.com/patrickziegler/strava-heatmap-proxy
cd strava-heatmap-proxy
GOBIN=~/.local/bin go install cmd/strava-heatmap-proxy/strava-heatmap-proxy.go
```

Then you can run the proxy with

```sh
strava-heatmap-proxy --config "<path-to-config>"
```

whereby the config is expected to be a json formatted file holding Strava login credentials

```json
{
  "Email": "...",
  "Password": "..."
}
```

### Usage

The following [TMS](https://wiki.openstreetmap.org/wiki/TMS) file can be used to define a new layer for streaming heatmap tiles into any kind of software that supports this

```xml
<TMS>
    <Layer idx="0">
        <ServerUrl>http://localhost:8080/tiles-auth/all/hot/%1/%2/%3.png</ServerUrl>
    </Layer>
</TMS>
```

This [Screenshot](https://imgur.com/a/6ELYene) shows how it would look like in [QMapShack](https://github.com/Maproom/qmapshack/wiki)

### Additional Note

It is also possible to put the `CloudFront-*` parameters directly into the TMS file as shown below (`strava-heatmap-proxy` is printing them out on startup). In this case, it would not be necessary to keep the proxy running in the background, but you would need to update the file every once in a while as those parameters will expire after some time

```xml

<TMS>
    <Layer idx="0">
        <ServerUrl>https://heatmap-external-a.strava.com/tiles-auth/all/hot/%1/%2/%3.png?v=19&amp;Key-Pair-Id=...&amp;Policy=...&amp;Signature=...</ServerUrl>
    </Layer>
</TMS>
```

## References

1. Discussion in https://github.com/bertt/wmts/issues/2 revealed the importance of `CloudFront-*` parameters
1. https://github.com/erik/strava-heatmap-proxy is following a similar approach but is designed as a Cloudflare worker

## License

This project is licensed under the GPL - see the [LICENSE](LICENSE) file for details
