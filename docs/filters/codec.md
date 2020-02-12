---
title: Codec
parent: Filters
nav_order: 1
---

# Codec
Values in this filter define a whitelist of the codecs you want to <b>include</b> in the modifed manifest. Video and Audio Filters are defined seperately. Passing an empty value for either video or audio filter will remove all.

## Protocol Support

hls | dash |
----|------|
no  | yes  |

## Supported Values

| codec         | values | example |
|---------------|--------|---------|
| AVC           | avc    | v(avc)  |
| HEVC          | hvc    | v(hvc)  |
| Dolby         | dvh    | v(dvh)  |
| AAC           | mp4a   | a(mp4a) |
| AC-3          | ac-3   | a(ac-3) |
| Enhanced AC-3 | ec-3   | a(ec-3) |

## Usage Examples

You can add mutliple codec values to each video and audio filter respecitively. EX: `v(avc, hvc)`
