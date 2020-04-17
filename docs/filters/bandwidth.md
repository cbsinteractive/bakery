---
title: Bandwidth
parent: Filters
nav_order: 3
---

# Bandwidth
An **INCLUSIVE RANGE** of variant bitrates to **INCLUDE** in the modified manifest. Variants outside this range will be filtered out. If a single value is provided, it will define the minimum bitrate desired in the modified manifest.

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | yes  |

### Keys

| name          | key |
|:-------------:|:---:|
| bandwidth     | b() |

### Values

| values (bps) | example   |
|:-------------:|:---------:|
| (min)         | b(500)    |
| (min, max)    | b(0,1000) |

## Usage Example
Range is supplied with `,` and no space in between

    // Define minimum bitrate as 500 Kbps
    $ http http://bakery.dev.cbsivideo.com/b(500000)/star_trek_discovery/S01/E01.m3u8

    // Define a maximum bitrate 1MB
    $ http http://bakery.dev.cbsivideo.com/b(0,1000000)/star_trek_discovery/S01/E01.m3u8

    // Define an inclusive range of 1MB and 5MB
    $ http http://bakery.dev.cbsivideo.com/b(1000000,5000000)/star_trek_discovery/S01/E01.m3u8