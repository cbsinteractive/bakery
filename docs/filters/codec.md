---
title: Codec
parent: Filters
nav_order: 1
---

# Codec
Values in this filter define a whitelist of the audio and video codecs you want to **EXCLUDE** in the modifed manifest. Passing an empty value for either video or audio will return all audio and video codecs available in the manifest.
<br>
By default, the audio and video filter keys take codecs, but this isn't the only way to use them. You can nest filters to filter more than just codecs, as detailed on the <a href="nested-filters.html">nested filters</a> page.

## Protocol Support

HLS | DASH |
:--:|:----:|
yes | yes  |

## Supported Values

| codec         | values | example |
|:-------------:|:------:|:-------:|
| AVC           | avc    | v(avc)  |
| HEVC          | hvc    | v(hvc)  |
| HDR10         | hdr10  | v(hdr10)|
| Dolby         | dvh    | v(dvh)  |
| AAC           | mp4a   | a(mp4a) |
| AC-3          | ac-3   | a(ac-3) |
| Enhanced AC-3 | ec-3   | a(ec-3) |

## Usage Example 
### Single value filter:

    // Removes MPEG-4 audio
    $ http http://bakery.dev.cbsivideo.com/a(mp4a)/star_trek_discovery/S01/E01.m3u8

    // Removes AVC video
    $ http http://bakery.dev.cbsivideo.com/v(avc)/star_trek_discovery/S01/E01.m3u8


### Multi value filter:
Mutli value filters are `,` with no space in between

    // Removes AC-3 and Enhanced EC-3 audio from the manifest
    $ http http://bakery.dev.cbsivideo.com/a(ac-3,ec-3)/star_trek_discovery/S01/E01.m3u8

    // Removes HDR10 and Dolby Vision video from the manifest
    $ http http://bakery.dev.cbsivideo.com/v(hdr10,dvh)/star_trek_discovery/S01/E01.m3u8

### Multiple filters:
Mutliple filters are supplied by using the `/` with no space in between

    // Removes AVC video and MPEG-4 audio
    $ http http://bakery.dev.cbsivideo.com/v(avc)/a(mp4a)/star_trek_discovery/S01/E01.m3u8

