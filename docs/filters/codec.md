---
title: Codec
parent: Filters
nav_order: 1
---

# Codec
Values in this filter define a whitelist of the codecs and formats you want to **EXCLUDE** in the modifed manifest, with the key denoting the content type you are targetting.

By default, the audio, video, and caption keys will accept codecs as their value. but this is not the only way to use them. You can <a href="nested-filters.html">nest</a> other filters to target video, audio, and caption media types.

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | yes  |

### Keys

| name    | key |
|:-------:|:---:|
| video   | v() |
| audio   | a() |
| caption | c() |

### Values

| values  | example    | description   |
|:-------:|:----------:|:-------------:|
| avc     | v(avc)     | AVC           |
| hvc     | v(hvc)     | HEVC          |
| mp4a    | a(mp4a)    | AAC           |
| ac-3    | a(ac-3)    | AC-3          |
| ec-3    | a(ec-3)    | Enhanced AC-3 |
| wvtt    | c(wvtt)    | Web VTT       |
| sptt    | c(sptt)    | Subtitle      |

## Usage Example 
### Single value filter:

    // Removes MPEG-4 audio
    $ http http://bakery.dev.cbsi.video/a(mp4a)/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video
    $ http http://bakery.dev.cbsi.video/v(avc)/star_trek_discovery/S01/E01.m3u8

    // Removes I-Frame
    $ http http://bakery.dev.cbsi.video/v(i-frame)/star_trek_discovery/S01/E01.m3u8

### Multi value filter:
Mutli value filters are `,` with no space in between

    // Removes AC-3 and Enhanced EC-3 audio from the manifest
    $ http http://bakery.dev.cbsi.video/a(ac-3,ec-3)/star_trek_discovery/S01/E01.m3u8

    // Removes HDR10 and Dolby Vision video from the manifest
    $ http http://bakery.dev.cbsi.video/v(hdr10,dvh)/star_trek_discovery/S01/E01.m3u8

### Multiple filters:
Mutliple filters are supplied by using the `/` with no space in between

    // Removes AVC video and MPEG-4 audio
    $ http http://bakery.dev.cbsi.video/v(avc)/a(mp4a)/star_trek_discovery/S01/E01.m3u8

