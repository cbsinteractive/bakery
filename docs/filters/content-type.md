---
title: Content Type
parent: Filters
nav_order: 2
---

# Content Type

Values in this filter define a whitelist of content types you want to **EXCLUDE** in the modifed manifest.

## Protocol Support

HLS | DASH |
:--:|:----:|
no  | yes  |

## Supported Values

| stream type | values | example   |
|-------------|--------|-----------|
| video       | video  | ct(video) |
| audio       | audio  | ct(audio) |
| text        | text   | ct(text)  |
| image       | image  | ct(image) |

## Usage Example 
### Single value filter:

    // Removes any file stream of type audio
    $ http http://bakery.dev.cbsivideo.com/ct(audio)/star_trek_discovery/S01/E01.mpd

    // Removes any file stream of type video
    $ http http://bakery.dev.cbsivideo.com/ct(video)/star_trek_discovery/S01/E01.mpd

### Multi value filter:
Mutli value filters are `,` with no space in between

    // Removes any content of type audio and video
    $ http http://bakery.dev.cbsivideo.com/ct(audio,video)/star_trek_discovery/S01/E01.mpd

    // Removes any content of type text and image
    $ http http://bakery.dev.cbsivideo.com/ct(text,image)/star_trek_discovery/S01/E01.mpd

