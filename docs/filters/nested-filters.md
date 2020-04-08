---
title: Nested Filters
parent: Filters
nav_order: 10
---

# Nested Filters
A way to apply filters that targets specific content type within a given playlist. These filters behave like their non-nested versions, this time being supplied as values for our video, audio, and/or caption type filters. 

If you haven't had the chance, we suggest getting started with our Quick Start guide before trying to apply filters. You can find it <a href="/bakery/quick-start/2020/03/05/quick-start.html">here</a>!

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | yes  |

### Keys

| content types | key |
|:-------------:|:---:|
| audio         | a() |
| video         | v() |
| caption       | c() |

### Values

| sub filters | key  |
|:----------:|:----:|
| codec      | co() |
| bandwidth  | b()  |
| language   | l()  |


## Limitations
Nested filters were introduced as a way to target demuxed media represented in a playlist. If both nested and overall filters are applied, some filters may behave differently than expected. 

It is recommended that the nested version of these filters be used for demuxed content only. If your playlist has muxed content, we recommend applying an overall bandwidth filter. In cases where playlists are interleaved with both muxed and demuxed streams, only nested filters should be used.

### Language
We do not apply the language filter to a video target. 

## Usage Example
### Single Nested Filter:

    // Removes MPEG-4 audio
    $ http http://bakery.dev.cbsivideo.com/a(co(mp4a))/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video
    $ http http://bakery.dev.cbsivideo.com/v(co(avc))/star_trek_discovery/S01/E01.m3u8

    // Removes audio outside of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/a(b(500,1000))/star_trek_discovery/S01/E01.m3u8

    // Removes video outside of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(b(500,1000))/star_trek_discovery/S01/E01.m3u8

### Multiple Nested Filters:
To use multiple nested filters, separate with `,` with no space between nested filters.

    // Removes MPEG-4 audio and audio not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/a(co(mp4a),b(500,1000))/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video and video not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(co(avc),b(500,1000))/star_trek_discovery/S01/E01.m3u8

### Multiple Filters:
To use multiple filters, separated with `/` with no space between filters. You can use nested filters in conjunction with the general filters, such as the bandwidth filter.

    // Removes AVC video, MPEG-4 audio, audio not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(co(avc))/a(co(mp4a),b(500,1000))/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video, MPEG-4 audio, and everything not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(co(avc))/a(co(mp4a))/b(500,1000)/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video, all video not in range 750Kbps to 1MB, MPEG-4 audio, and non-video not in range of 500Kbps to 1MB
    $ http http://bakery.dev.cbsivideo.com/v(co(avc),b(750))/a(co(mp4a))/b(500,1000)/star_trek_discovery/S01/E01.m3u8