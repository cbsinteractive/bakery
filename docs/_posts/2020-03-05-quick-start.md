---
layout: post
title: Quick Start
category: quick-start
---

# Quick Start

This tutorial is meant to familiarize you with how bakery works as a proxy to be able to filter your manifest.


## Setting up your Origin

Origin hosts are currently managed by the video-processing-team. If you would like to configure your origin to use bakery as a proxy, reach out by simply hopping into our channel in <a href="https://cbs.slack.com/app_redirect?channel=i-vidtech-mediahub" target="_blank">Slack</a> and we'll get you setup!

In the meantime, check out the <a href="https://github.com/cbsinteractive/bakery">project repo</a> to run bakery in your local environment!

Once we have configured bakery to point to your origin, if you have the following playback URL `http://streaming.cbsi.video/star_trek_discovery/S01/E01.m3u8`

Then your `BAKERY_ORIGIN_HOST` was set to `http://streaming.cbsi.video` and your playback URL on the proxy will be `http://bakery.dev.cbsivideo.com/star_trek_discovery/S01/E01.m3u8`. 

## Applying Filters

If you want to apply filters, they should be placed right after the Bakery origin host. Following the example above you can start applying filters like so:

1. **Single Filter**
    <br>To apply a single filter such as an audio codec filter, you would have a url in the form of `http://bakery.dev.cbsivideo.com/a(ac-3)/star_trek_discovery/S01/E01.m3u8` where AC-3 audio is removed from the manifest.

2. **Multiple Values**
    <br>You can supply multiple values to each filter as you would like simply by using `,` as your delimiter for each value. The url `http://bakery.dev.cbsivideo.com/a(ac-3,ec-3)/star_trek_discovery/S01/E01.m3u8` will filter out AC-3 audio and Enhanced AC-3 audio from the manifest.

3. **Multiple Filters**
    <br>Mutliple Filters can be passed in. All that is needed is the `/` delimiter in between each filter. For example, if you wanted to filter a specific audio and video codec, you can do so with the following url, `http://bakery.dev.cbsivideo.com/a(mp4a)/v(avc)/star_trek_discovery/S01/E01.m3u8`, which will remove AVC (H.264) video and AAC (MPEG-4) audio


For more specific details and usage examples on specific filters and the values accepted by each, check out our documentation for filters <a href="/filters">here</a>!


## What's Next?

Thank you for choosing Bakery! As we are just getting started on this brand new service, we will be sure to post more tutorials and various posts as more features become available. Stay tuned!

Stuck or confused? Want to say hi? Reach out to us in <a href="https://cbs.slack.com/app_redirect?channel=i-vidtech-mediahub" target="_blank">Slack</a> and we'll help you out!