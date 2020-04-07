---
title: I-Frame
parent: Filters
nav_order: 5
---

# I-Frame
When set, the I-Frame filter wil remove the I-frame from the playlist.

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | no   |

### Keys

| name    | key |
|:-------:|:---:|
| video   | v() |

### Values

| values  | example    |
|:-------:|:----------:|
| i-frame | v(i-frame) |


## Usage Example 
### Single value filter:
    // Removes I-Frame
    $ http http://bakery.dev.cbsivideo.com/v(i-frame)/star_trek_discovery/S01/E01.m3u8

### Multi value filter:
Mutli value filters are `,` with no space in between
    // Removes HDR10 and Dolby Vision video from the manifest
    $ http http://bakery.dev.cbsivideo.com/v(i-frame,hdr10,dvh)/star_trek_discovery/S01/E01.m3u8

### Multiple filters:
Mutliple filters are supplied by using the `/` with no space in between

    // Removes AVC video and MPEG-4 audio
    $ http http://bakery.dev.cbsivideo.com/v(i-frame,hevc)/a(mp4a)/star_trek_discovery/S01/E01.m3u8

