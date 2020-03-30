---
title: Trim
parent: Filters
nav_order: 5
---

# Trim
An **INCLUSIVE RANGE** of segments to **INCLUDE** in the modified variant playlist. Segments outside this range will be filtered out. The Playlist returned will be a Video on Demand Playlist. 

## Protocol Support

HLS | DASH |
:--:|:----:|
yes | no   |

## Supported Values

| values (epoch) | example                  |
|:--------------:|:------------------------:|
| (start, end)   | t(1585335477,1585335677) |

## Usage Example
Range is supplied with `,` and no space in between the epoch timestamps

    // Define range of variant playlists
    $ http http://bakery.dev.cbsivideo.com/t(1585335477,1585335677)/star_trek_discovery/S01/E01.m3u8

## Limitations

### Segments
Bakery will trim segments based on what is already advertised in the Variant Playlist. If you have a Live Playlist with a sliding window and only 10 segments advertised, you will only be able to trim within the range of those 10 segments. It is recommended that this feature be used on VOD or EVENT Playlists where the full segment archive is available. For Live playlist, you can increase the size of your retention window so that the sliding window can hold a longer range of segments. 

### Timestamp
The epoch timestamps provided should be relative to the Program Date Time advertised in your Variant Playlists.

### Program Date Time
The Program Date Time must be enabled for every segment that is advertised in the manifest. 



