---
title: Deweave
parent: Filters
nav_order: 11
---

# Deweave
If you have redundant streams in your playlist, you can use the deweave filter to remove any streams that are unavailable or stale. When set, Bakery will check your redundant streams and create a simple manifest with a single stream. 

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | no   |

### Keys

| name     | key    |
|:--------:|:------:|
| deweave  | dw()   |

### Values

| values  | example    |
|:-------:|:----------:|
| true    | dw(true)   |
| false   | dw(false)  |


## Usage Example 

Assuming we have the following weaved manifest:

```
http https://bakery.dev.cbsi.video/propeller/cbsi679d/testa5fe.m3u8
HTTP/1.1 200 OK

#EXTM3U
#EXT-X-VERSION:4
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=356400,CODECS="avc1.64000c,mp4a.40.2",RESOLUTION=400x224,FRAME-RATE=15.000
testa5fe_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=356400,CODECS="avc1.64000c,mp4a.40.2",RESOLUTION=400x224,FRAME-RATE=15.000
backup_testa5fe_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=528000,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,FRAME-RATE=15.000
testa5fe_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=528000,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,FRAME-RATE=15.000
backup_testa5fe_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100000,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,FRAME-RATE=30.000
testa5fe_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100000,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,FRAME-RATE=30.000
backup_testa5fe_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=2129600,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=960x540,FRAME-RATE=30.000
testa5fe_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=2129600,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=960x540,FRAME-RATE=30.000
backup_testa5fe_4.m3u8
```

We can make a request, setting the deweave filter to true, which will return the following manifest: 

```
http "https://bakery.dev.cbsi.video/dw(true)/propeller/cbsi679d/testa5fe.m3u8"
HTTP/1.1 200 OK

#EXTM3U
#EXT-X-VERSION:4
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=356400,CODECS="avc1.64000c,mp4a.40.2",RESOLUTION=400x224,FRAME-RATE=15.000
testa5fe_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=528000,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,FRAME-RATE=15.000
testa5fe_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100000,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,FRAME-RATE=30.000
testa5fe_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=2129600,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=960x540,FRAME-RATE=30.000
testa5fe_4.m3u8
``` 

In this case, the primary stream was found to be healthy and returned accordingly. If the primary stream returns a `404` or has gone stale, the manifest will consist of the backup stream instead.


### Single value filter:
    // Deweave manifest
    $ http http://bakery.dev.cbsivideo.com/dw(true)/star_trek_discovery/S01/E01.m3u8```

### Multiple filters:
Mutliple filters are supplied by using the `/` with no space in between

    // Deweave manifest and remove the I-frame
    $ http http://bakery.dev.cbsivideo.com/dw(true)/tags(i-frame)/star_trek_discovery/S01/E01.m3u8

