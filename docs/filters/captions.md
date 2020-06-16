---
title: Caption Types
parent: Filters
nav_order: 8
---

# Caption Types
Values in this filter define a whitelist of the caption types you want to **EXCLUDE** in the modifed manifest.

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | yes  |

### Keys

| name          | key |
|:-------------:|:---:|
| caption type  | c() |

### Values

| values | example | description |
|:------:|:-------:|:-----------:|
| stpp   | c(stpp) | Subtitles   |
| wvtt   | c(wvtt) | WebVTT      |


## Usage Example 
### Single value filter:

    $ http http://bakery.dev.cbsi.video/c(stpp)/star_trek_discovery/S01/E01.m3u8

    $ http http://bakery.dev.cbsi.video/c(wvtt)/star_trek_discovery/S01/E01.m3u8


### Multi value filter:
Mutli value filters are `,` with no space in between

    $ http http://bakery.dev.cbsi.video/c(stpp,wvtt)/star_trek_discovery/S01/E01.m3u8

