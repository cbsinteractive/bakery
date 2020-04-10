---
title: Frame Rate
parent: Filters
nav_order: 10
---

# Frame Rate
When set, any variants or representations that match the supplied frame rate will be removed from their respective playlists. 

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | yes  |

### Keys

| name   | key   |
|:-------:|:----:|
| fps    | fps() |

### Values

| values         | example         |
|:--------------:|:---------------:|
| integer        | fps(60)         |
| floating point  | fps(59.94)      |
| fractions      | fps(30000:1001) |

The values above should match what is advertised in your playlist. HLS does not use fraction representations of frame rate, while DASH does. 

## Usage Example 
### Single value filter:

    // Removes 59.94 variant
    $ http http://bakery.dev.cbsivideo.com/fps(59.94)/star_trek_discovery/S01/E01.m3u8

### Multi value filter:
Mutli value filters are `,` with no space in between

    // Removes video representations with frame rate 29.97 (expressed as fraction) and 24 frames
    $ http http://bakery.dev.cbsivideo.com/v(i-frame)/fps(30000:1001,24)/star_trek_discovery/S01/E01.mpd

### Multiple filters:
Mutliple filters are supplied by using the `/` with no space in between

    // Removes the I-Frame and any variants with 60 fps
    $ http http://bakery.dev.cbsivideo.com/v(i-frame)/fps(60)/star_trek_discovery/S01/E01.m3u8
