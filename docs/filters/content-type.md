---
title: Content Type
parent: Filters
nav_order: 9
---

# Content Type

Values in this filter define a whitelist of content types you want to **EXCLUDE** in the modifed manifest.

## Support

### Protocol

HLS | DASH |
:--:|:----:|
no  | yes  |

### Keys

| name          | key  |
|:-------------:|:----:|
| content type  | ct() |

### Values

| values | example   |
|:------:|:---------:|
| video  | ct(video) |
| audio  | ct(audio) |
| text   | ct(text)  |
| image  | ct(image) |

## Usage Example 
### Single value filter:

    // Removes any file stream of type audio
    $ http http://bakery.dev.cbsi.video/ct(audio)/star_trek_discovery/S01/E01.mpd

    // Removes any file stream of type video
    $ http http://bakery.dev.cbsi.video/ct(video)/star_trek_discovery/S01/E01.mpd

### Multi value filter:
Mutli value filters are `,` with no space in between

    // Removes any content of type audio and video
    $ http http://bakery.dev.cbsi.video/ct(audio,video)/star_trek_discovery/S01/E01.mpd

    // Removes any content of type text and image
    $ http http://bakery.dev.cbsi.video/ct(text,image)/star_trek_discovery/S01/E01.mpd

