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
### Single value filter:

    // Deweave manifest
    $ http http://bakery.dev.cbsivideo.com/dw(true)/star_trek_discovery/S01/E01.m3u8

### Multiple filters:
Mutliple filters are supplied by using the `/` with no space in between

    // Deweave manifest and remove the I-frame
    $ http http://bakery.dev.cbsivideo.com/dw(true)/tags(i-frame)/star_trek_discovery/S01/E01.m3u8

