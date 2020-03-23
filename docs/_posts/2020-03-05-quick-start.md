---
layout: post
title: Quick Start
category: quick-start
---

# Quick Start

This tutorial is meant to familiarize you with how Bakery works as a proxy to be able to filter your manifest.


## Setting up your Origin

Origin hosts are currently managed by the video-processing-team. If you would like to configure your origin to use Bakery as a proxy, reach out by simply hopping into our channel in <a href="https://cbs.slack.com/app_redirect?channel=i-vidtech-mediahub" target="_blank">Slack</a> and we'll get you setup!

In the meantime, check out the <a href="https://github.com/cbsinteractive/bakery">project repo</a> to run Bakery in your local environment!

Once we have configured Bakery to point to your origin, if you have the following playback URL `http://streaming.cbsi.video/star_trek_discovery/S01/E01.m3u8`

Then your `BAKERY_ORIGIN_HOST` was set to `http://streaming.cbsi.video` and your playback URL on the proxy will be `http://bakery.dev.cbsivideo.com/star_trek_discovery/S01/E01.m3u8`. 


## Working with Propeller as your Origin

Bakery can act as a proxy to serve your Propeller channels!

Propeller is a live-streaming platform that manages all of your cloud based resources. If you have a stream running out of Propeller, you can stream that channel via Bakery! 


To play a Propeller channel you can make a Bakery request with the following scheme:

    https://bakery.dev.cbsivideo.com/propeller/<org-id>/<channel-id>.m3u8

For more information on working with Propeller, check out the documentation 
<a href="https://cbsinteractive.github.io/propeller/">here</a> or reach out to the team on <a href="https://cbs.slack.com/app_redirect?channel=i-vidtech-propeller" target="_blank">Slack</a> to get all set up!


## Applying Filters

If you want to apply filters, they should be placed right after the Bakery hostname and before the path. Your requests should match the following schema:

    http://bakery.dev.cbsivideo.com/[filters]/path/to/master/manifest.m3u8

If working with a Propeller origin:

    http://bakery.dev.cbsivideo.com/[filters]/propeller/<org-id>/<channel-id>.m3u8


Following the examples above you can start applying filters like so:

1. **Single Filter**
    <br>To apply a single filter such as an audio codec filter where AC-3 audio is removed from the manifest, you can make a request to Bakery as so:

    ```
    http://bakery.dev.cbsivideo.com/a(ac-3)/star_trek_discovery/S01/E01.m3u8
    ```
    for a Propeller channel:
    ```
    http://bakery.dev.cbsivideo.com/a(ac-3)/propeller/<org-id>/<channel-id>.m3u8
    ```

2. **Multiple Values**
    <br>You can supply multiple values to each filter as you would like simply by using `,` as your delimiter for each value. 

    The following example will filter out AC-3 audio and Enhanced AC-3 audio from the manifest:

    ```
    http://bakery.dev.cbsivideo.com/a(ac-3,ec-3)/star_trek_discovery/S01/E01.m3u8
    ```
    for a Propeller channel:
    ```
    http://bakery.dev.cbsivideo.com/a(ac-3,ec-3)/propeller/<org-id>/<channel-id>.m3u8
    ```

3. **Multiple Filters**
    <br>Mutliple Filters can be passed in. All that is needed is the `/` delimiter in between each filter. For example, if you wanted to remove AVC (H.264) video and AAC (MPEG-4) audio, you could make the following request to Bakery:

    ```
    http://bakery.dev.cbsivideo.com/a(mp4a)/v(avc)/star_trek_discovery/S01/E01.m3u8
    ```
    for a Propeller channel:
    ```
    http://bakery.dev.cbsivideo.com/a(mp4a)/v(avc)/propeller/<org-id>/<channel-id>.m3u8
    ```

4. **Nested Filters**
    <br>You can nest codec and bitrate filters within audio and video filters. For example, if you wanted to remove AAC (MPEG-4) audio and filter the results within 500 Kbps and 1MB, you could make the following request to Bakery:

    ```
    http://bakery.dev.cbsivideo.com/a(co(mp4a),b(500,1000))/star_trek_discovery/S01/E01.m3u8
    ```
    for a Propeller channel:
    ```
    http://bakery.dev.cbsivideo.com/a(co(mp4a),b(500,1000))/propeller/<org-id>/<channel-id>.m3u8
    ```

For more specific details and usage examples on specific filters and the values accepted by each, check out our documentation for filters <a href="/filters">here</a>!


## What's Next?

Thank you for choosing Bakery! As we are just getting started on this brand new service, we will be sure to post more tutorials and various posts as more features become available. Stay tuned!

Stuck or confused? Want to say hi? Reach out to us in <a href="https://cbs.slack.com/app_redirect?channel=i-vidtech-mediahub" target="_blank">Slack</a> and we'll help you out!