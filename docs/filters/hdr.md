---
title: HDR
parent: Filters
nav_order: 4
---

# HDR
HDR can be filtered out of your playlist when specifiying the HDR format used. 

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | yes   |

### Keys

| name    | key |
|:-------:|:---:|
| video   | v() |

### Values

| values  | example    | description |
|:-------:|:----------:|:-----------:|
| hdr10   | v(hdr10)   | HDR10       |
| dvh     | v(dvh)     | Dolby       |


## Usage Example 
### Single value filter:
    // Removes HDR10
    $ http http://bakery.dev.cbsivideo.com/v(hdr10)/star_trek_discovery/S01/E01.m3u8

    // Removes Dolby Vision
    $ http http://bakery.dev.cbsivideo.com/v(dvh)/star_trek_discovery/S01/E01.m3u8

### Multi value filter:
Mutli value filters are `,` with no space in between
    // Removes I-Frame, HDR10 and Dolby Vision video from the manifest
    $ http http://bakery.dev.cbsivideo.com/v(i-frame,hdr10,dvh)/star_trek_discovery/S01/E01.m3u8

### Multiple filters:
Mutliple filters are supplied by using the `/` with no space in between

    // Removes HDR10 and include Audio with bitrate range of 500Kbps and 1MB
    $ http http://bakery.dev.cbsivideo.com/v(hdr10)/a(b(500,1000))/star_trek_discovery/S01/E01.m3u8

