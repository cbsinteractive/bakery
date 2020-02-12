---
title: Caption Type
parent: Filters
nav_order: 4
---

# Caption Type
Values in this filter define a whitelist of the caption types you want <b>include</b> in the modifed manifest. Passing an empty value for this filter will removeall caption types from the manifest.

## Protocol Support

hls | dash |
----|------|
no  | yes  |

## Supported Values

| codec      | values | example    |
|------------|--------|------------|
| Subtitles  | "stpp" | ct("stpp") |
| WebVTT     | "wvtt" | ct("wvtt") |
