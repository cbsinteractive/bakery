<p align="center">
  <img width="250" src="http://flv.io/bakery-logo.svg">
</p>


Bakery is a proxy and filter for HLS and DASH manifests.

## Setting up environment for development

#### Clone this project:

    $ git clone https://github.com/cbsinteractive/bakery.git

#### Export the environment variables:

    $ export BAKERY_CLIENT_TIMEOUT=5s 
    $ export BAKERY_HTTP_PORT=:8082
    $ export BAKERY_ORIGIN_HOST="https://streaming.cbs.com" 

Note that `BAKERY_ORIGIN_HOST` will be the base URL of your manifest files.

##### Usage Example:

If your playback URL is `http://streaming.cbsi.video/star_trek_discovery/S01/E01.m3u8`, your `BAKERY_ORIGIN_HOST` should be set to `http://streaming.cbsi.video`. The playback URL on the proxy will be `http://bakery.host.here/star_trek_discovery/S01/E01.m3u8`. 

If you want to apply filters, they should be placed right after the Bakery host. Following the example above, if you want to filter out all the levels that are outside of a given bitrate range, the playback URL should be: `http://bakery.host.here/b(1000,4000)/star_trek_discovery/S01/E01.m3u8`, where 1000Kbps and 4000Kbps are the lower and higher boundaries.

Mutliple Filters can be passed in. Following the example above, if you wanted to filter a specific audio, video codec and the bitrate range, you can do so the following way:
`http://bakery.host.here/v(avc)/a(mp4a)/b(1000,4000)/star_trek_discovery/S01/E01.m3u8`


##### Supported Filters: 
Supported Filters for each protocol can be found [here](api.md)

#### Run the API:

    $ make run

The API will be available on http://localhost[:BAKERY_HTTP_PORT]

## Run Tests

    $ make  test
