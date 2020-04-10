---
title: I-Frame
parent: Filters
nav_order: 5
---

# I-Frame
When set, the I-Frame filter will remove the I-Frame from the playlist. I-Frame is suppressed via the Tags filter.

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | no   |

### Keys

| name    | key    |
|:-------:|:------:|
| tags    | tags() |

### Values

| values  | example       |
|:-------:|:-------------:|
| i-frame | tags(i-frame) |
| iframe  | tags(iframe)  |


## Usage Example 
### Single value filter:

    // Removes I-Frame
    $ http http://bakery.dev.cbsivideo.com/v(i-frame)/star_trek_discovery/S01/E01.m3u8
    $ http http://bakery.dev.cbsivideo.com/v(iframe)/star_trek_discovery/S01/E01.m3u8

### Multiple filters:
Mutliple filters are supplied by using the `/` with no space in between

    // Removes the I-Frame, HDR10 and Dolby Vision video from the manifest
    $ http http://bakery.dev.cbsivideo.com/v(hdr10,dvh)/tags(i-frame)/star_trek_discovery/S01/E01.m3u8

