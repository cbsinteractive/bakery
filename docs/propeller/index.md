---
title: Propeller
nav_order: 3
has_children: true
has_toc: false
---

# Propeller

If you haven't had the chance, we suggest getting started with our <a href="/bakery/quick-start/2020/03/05/quick-start.html">Quick Start</a> guide before trying to proxy and filter your channels via Bakery. For help on
managing playback of your Propeller channels via Bakery, check out the documentation below. 

## Playback

Bakery can be used to manage your Propeller playback URLs. 

### Channels

To request a Propeller channel via Bakery:

    https://bakery.dev.cbsivideo.com/propeller/<org-id>/<channel-id>.m3u8

Bakery will then set the Playback URL with the following priority, depending on your channel settings:

1. Startfruit
2. DAI
3. Carrier
4. Playback URL
5. Archive
{: .lh-tight }

As long as you're channel was set to archive, Bakery will automatically proxy the archive stream when your Propeller channel has ended. 

### Clips

To request a Propeller clip via Bakery:

    https://bakery.dev.cbsivideo.com/propeller/<org-id>/clip/<clip-id>.m3u8

