---
title: Stream Type
parent: Filters
nav_order: 2
---

# Stream Type

Values in this filter define stream types you wish to <b>remove</b> from your manifest. The filter in this  example will filter out all audio streams from the modified manifest.

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

You can add mutliple values. EX: `fs(audio, text)`

