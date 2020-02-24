## Filters

* [Codec](###Codec) 
* [Stream Type](###Stream-Type) 
* [Bandwidth](###Bandwidth)
* [Caption Type](###Caption-Type) 

### Codec
Values in this filter define a whitelist of the codecs you want to <b>include</b> in the modifed manifest. Video and Audio Filters are defined seperately. Passing an empty value for either video or audio filter will remove all.

#### Protocol Support

hls | dash |
----|------|
yes  | yes  |

#### Supported Values

| codec         | values | example |
|---------------|--------|---------|
| AVC           | avc    | v(avc)  |
| HEVC          | hvc    | v(hvc)  |
| Dolby         | dvh    | v(dvh)  |
| AAC           | mp4a   | a(mp4a) |
| AC-3          | ac-3   | a(ac-3) |
| Enhanced AC-3 | ec-3   | a(ec-3) |

You can add mutliple codec values to each video and audio filter respecitively. EX: `v(avc, hvc)`

### Stream Type

Values in this filter define stream types you wish to <b>remove</b> from your manifest. The filter in this example will filter out all audio streams from the modified manifest.

#### Protocol Support

hls | dash |
----|------|
no  | yes  |

#### Supported Values

| stream type | values | example   |
|-------------|--------|-----------|
| video       | video  | fs(video) |
| audio       | audio  | fs(audio) |
| text        | text   | fs(text)  |
| image       | image  | fs(image) |

You can add mutliple values. EX: `fs(audio, text)`


### Bandwidth
An inclusive range of variant bitrates to <b>include</b> in the modified manifest, variants outside this range will be filtered out. If a single value is provided, it will define the minimum bitrate desired in the modified manifest

#### Protocol Support

hls | dash |
----|------|
yes | yes  |

#### Supported Values

| values (bps) | example   |
|---------------|-----------|
| (min, max)    | b(0,1000) |
| (min)         | b(1000)   |

### Caption Type
Values in this filter define a whitelist of the caption types you want <b>include</b> in the modifed manifest. Passing an empty value for this filter will remove all caption types from the manifest.

#### Protocol Support

hls | dash |
----|------|
yes  | yes  |

#### Supported Values

| codec      | values | example    |
|------------|--------|------------|
| Subtitles  | "stpp" | ct("stpp") |
| WebVTT     | "wvtt" | ct("wvtt") |


## Help

You can find the source code for Bakery at GitHub:
[bakery][bakery]

[bakery]: https://github.com/cbsinteractive/bakery

If you have any questions regarding Bakery, please reach out in the [#i-vidtech-mediahub](slack://channel?team={cbs}&id={i-vidtech-mediahub}) channel.
