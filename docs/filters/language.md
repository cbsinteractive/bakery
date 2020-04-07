---
title: Language
parent: Filters
nav_order: 6
---

# Language
Values in this filter define a whitelist of languages you want to **EXCLUDE** in the modifed manifest.

## Support

### Protocol

HLS | DASH |
:--:|:----:|
yes | yes  |

### Keys

| name          | key |
|:-------------:|:---:|
| language      | l() |

### Values
The values supplied to the language filter should match the language code used in your playlist. The values are not case sensitive.

## Limitations
### Content Type
When used, this filter is applied to **ALL** types of audio and captions. We do not apply the language filter to a video target. If you want to target a specific audio or caption track, check out our <a href="/nested-filters">documentation</a> on nested filters for targeting based on content type. 

### Lanugage Code
Different encoding engines will follow different language codes. Please use the language code used to build your playlist. For example, if using ISO 639-1, a value of `pt` targets only the country of Portugal and not the Portuguese language.

## Usage Example 
### Single value filter:

    //Remove Portuguese (Brazil)
    $ http http://bakery.dev.cbsivideo.com/l(pt-BR)/star_trek_discovery/S01/E01.m3u8

    //Remove Portuguese (Portugal)
    $ http http://bakery.dev.cbsivideo.com/l(pt)/star_trek_discovery/S01/E01.m3u8


### Multi value filter:
Mutli value filters are `,` with no space in between

    $ http http://bakery.dev.cbsivideo.com/l(pt,pt-BR)/star_trek_discovery/S01/E01.m3u8

