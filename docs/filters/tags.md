---
title: Tags
parent: Filters
nav_order: 6
---

# Tags
When set, tags passed in as values will be supressed from the manifest.

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
| ads     | tags(ads)     |

## Limitations
### Ads
Suppressing the ad tags will only be done when trimming your media playlists. 

## Usage Example 
### Single value filter:

    // Removes Ad Tags while trimming your media playlist
    $ http http://bakery.dev.cbsivideo.com/t(1585335477,1585335677)/(tags(ads)/star_trek_discovery/S01/E01.m3u8

### Multiple filters:
Mutliple filters are supplied by using the `/` with no space in between

    // Removes I-Frame from master playlist, suppresses Ad tags when trimming the media playlist
    $ http http://bakery.dev.cbsivideo.com/tags(i-frame,ads)/t(1585335477,1585335677)/star_trek_discovery/S01/E01.m3u8

