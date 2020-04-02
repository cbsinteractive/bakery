---
title: Audio
parent: Filters
nav_order: 3
---

# Audio
Values in this filter define a whitelist of filters you want to apply to your audio content. Filters passed with an audio key will ignore video and caption type content that is advertised in your manifest. 

## Protocol Support

HLS | DASH |
:--:|:----:|
yes | yes  |

## Supported Filters 

1. <a href="codec.html">Codec Filters</a>

2. <a href="bandwidth.html">Bandwidth Filters</a>

3. <a href="language.html">Language Filters</a>

| codec      | values | example  |
|:----------:|:------:|:--------:|
| Subtitles  | stpp   | c(stpp) |
| WebVTT     | wvtt   | c(wvtt) |


## Usage Example 
### Single value filter:

    $ http http://bakery.dev.cbsivideo.com/c(stpp)/star_trek_discovery/S01/E01.m3u8

    $ http http://bakery.dev.cbsivideo.com/c(wvtt)/star_trek_discovery/S01/E01.m3u8


### Multi value filter:
Mutli value filters are `,` with no space in between

    $ http http://bakery.dev.cbsivideo.com/c(stpp,wvtt)/star_trek_discovery/S01/E01.m3u8

