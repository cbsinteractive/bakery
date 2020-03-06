---
title: Stream Type
parent: Filters
nav_order: 2
---

# Stream Type

Values in this filter define a whitelist of stream types you want to **EXCLUDE** in the modifed manifest. Passing an empty value for stream type will return all audio and video codecs available in the manifest.

## Protocol Support

hls | dash |
----|------|
no  | yes  |

## Supported Values

| stream type | values | example   |
|-------------|--------|-----------|
| video       | video  | fs(video) |
| audio       | audio  | fs(audio) |
| text        | text   | fs(text)  |
| image       | image  | fs(image) |

## Usage Example 
### Single value filter:

    // Removes any file stream of type audio
    $ http http://bakery.dev.cbsivideo.com/fs(audio)/star_trek_discovery/S01/E01.mpd

    // Removes any file stream of type video
    $ http http://bakery.dev.cbsivideo.com/fs(video)/star_trek_discovery/S01/E01.mpd

### Multi value filter:
Mutli value filters are `,` with no space in between

    // Removes any file stream of type audio and video
    $ http http://bakery.dev.cbsivideo.com/fs(audio,video)/star_trek_discovery/S01/E01.mpd

    // Removes any file stream of type text and image
    $ http http://bakery.dev.cbsivideo.com/fs(text,image)/star_trek_discovery/S01/E01.mpd

