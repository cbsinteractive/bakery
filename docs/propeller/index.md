---
title: Propeller
nav_order: 3
has_children: true
has_toc: false
---

# Propeller
Propeller is the ViacomCBS live-streaming platform. Propeller can orchestrate and manage all of your cloud based resources needed for a live streaming environment. If you have a stream running out of Propeller, you can stream that channel via Bakery! 

For help on managing playback of your Propeller channels via Bakery, check out the documentation below. 

## Playback

Bakery can be used to manage your Propeller playback URLs. 

### Channels

To request a Propeller channel via Bakery:

    https://bakery.dev.cbsi.video/propeller/<org-id>/<channel-id>.m3u8

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

    https://bakery.dev.cbsi.video/propeller/<org-id>/clip/<clip-id>.m3u8



## Help

For more information on Propeller, check out the documentation <a href="https://cbsinteractive.github.io/propeller/">here</a> or reach out to the Propeller team on <a href="https://cbs.slack.com/app_redirect?channel=i-vidtech-propeller" target="_blank">Slack</a> to get all set up!