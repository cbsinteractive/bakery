---
title: Prevent HTTP Error
parent: Filters
nav_order: 4
---

# Prevent HTTP Status Error
When enabled, it will obfuscate 4xx or 5xx errors from the origin and return an empty M3U8 or WebVTT file.

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | no  |

### Keys

| name               | key |
|:------------------:|:---:|
| Prevent HTTP Error | phe() |

### Values

| values       | example           |
|:------------:|:-----------------:|
| (true\|false) | phe(true)         |


## Usage

The main purpose of this filter is to prevent some on-the-fly services to break playback by surfacing
4xx and/or 5xx files to the player. When generating your manifest file, if you want to use this filter,
you'll need to make sure to encode the file endpoint in base64 as follows:

```
http://bakery.dev.cbsi.video/phe(true)/BASE64-STRING.[vtt|m3u8]
```

Where the `BASE64-STRING` contains the full URL and PATH for your file, including its extension.

## Example

Imagine that I want to prevent players to hit a 404 when accessing the url `https://08763bf0b1gb.airspace-cdn.cbsivideo.com/mtv-ema-uk-hls/dictate_caption_1234.vtt`. Here's how I'll advertise this URL on my WebVTT playlist:

```
https://bakery.dev.cbsi.video/phe(true)/aHR0cHM6Ly8wODc2M2JmMGIxZ2IuYWlyc3BhY2UtY2RuLmNic2l2aWRlby5jb20vbXR2LWVtYS11ay1obHMvZGljdGF0ZV9jYXB0aW9uXzEyMzQudnR0.vtt
```

