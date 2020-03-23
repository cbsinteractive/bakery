---
title: Nested Filters
parent: Filters
nav_order: 5
---

# Nested Filters
A way to apply <a href="codec.html">codec</a> and <a href="bandwidth.html">bandwidth</a> filters to a specific type of content. The nested codec and bandwidth filters behave like their non-nested versions.


## Interaction with General Filters
When applying the nested bandwidth filter to a content type, the nested bandwidth range is limited by the overall bandwidth filter range. For example, if you specify a nested audio bandwidth range of 0 to 1Mbps and an overall bandwidth range of 0 to 500Kbps, any audio in the filtered manifest will be between the overall bandwidth range. If the content type's bandwidth filter does not overlap with the overall bandwidth filter, the content type's bandwidth filter won't be applied at all.

## Protocol Support

HLS | DASH |
:--:|:----:|
yes | yes  |

## Supported Values

| content types | example               |
|:-------------:|:---------------------:|
| audio         | a(co(ac-3),b(0,1000)) |
| video         | v(co(avc),b(0,1000))  |

| subfilters | example      |
|:----------:|:------------:|
| codec      | a(co(ac-3))  |
| bandwidth  | v(b(0,1000)) |


## Usage Example
### Single Nested Filter:

    // Removes MPEG-4 audio
    $ http http://bakery.dev.cbsivideo.com/a(codec(mp4a))/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc))/star_trek_discovery/S01/E01.m3u8

    // Removes audio outside of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/a(b(500,1000))/star_trek_discovery/S01/E01.m3u8

    // Removes video outside of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(b(500,1000))/star_trek_discovery/S01/E01.m3u8

### Multiple Nested Filter:
To use multiple nested filters, separate with `,` with no space between nested filters.

    // Removes MPEG-4 audio and audio not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/a(codec(mp4a),b(500,1000))/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video and video not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc),b(500,1000))/star_trek_discovery/S01/E01.m3u8

### Multiple Filters:
To use multiple filters, separated with `/` with no space between filters. You can use nested filters in conjunction with the general filters, such as the bandwidth filter.

    // Removes AVC video, MPEG-4 audio, audio not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc))/a(codec(mp4a),b(500,1000))/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video, MPEG-4 audio, and everything not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc))/a(codec(mp4a))/b(500,1000)/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video, all video not in range 750Kbps to 1MB, MPEG-4 audio, and non-video not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(codec(avc),b(750))/a(codec(mp4a))/b(500,1000)/star_trek_discovery/S01/E01.m3u8