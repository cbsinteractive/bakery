---
title: Bandwidth
parent: Filters
nav_order: 3
---

# Bandwidth
An inclusive range of variant bitrates to <b>include</b> in the modified manifest, variants outside this range will be filtered out. If a single value is provided, it will define the minimum bitrate desired in the modified manifest

## Protocol Support

hls | dash |
----|------|
yes | no  |

## Supported Values

| values (Kbps) | example   |
|---------------|-----------|
| (min, max)    | b(0,1000) |
| (min)         | b(1000)   |