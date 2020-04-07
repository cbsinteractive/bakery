<p align="center">
  <img width="250" src="http://flv.io/bakery-logo.svg">
</p>

[![CI Status](https://github.com/cbsinteractive/bakery/workflows/CI/badge.svg)](https://github.com/cbsinteractive/bakery/actions)
[![CD Status](https://github.com/cbsinteractive/bakery-deploy/workflows/CD/badge.svg)](https://github.com/cbsinteractive/bakery-deploy/actions)
[![codecov](https://codecov.io/gh/cbsinteractive/bakery/branch/master/graph/badge.svg)](https://codecov.io/gh/cbsinteractive/bakery)


Bakery is a proxy and filter for HLS and DASH manifests.

## Setting up environment for development

#### Clone this project:

    $ git clone https://github.com/cbsinteractive/bakery.git

#### Export the environment variables:

Please reach out to the [Propeller](https://cbsinteractive.github.io/propeller) team for configuring your access prior to working with propeller origin channels. 

    $ export BAKERY_CLIENT_TIMEOUT=5s 
    $ export BAKERY_HTTP_PORT=:8082
    $ export BAKERY_ORIGIN_HOST="https://streaming.cbs.com" 
    $ export BAKERY_HOSTNAME="localhost:8082" // include port for local developement
    $ export BAKERY_PROPELLER_HOST="http://propeller.com"
    $ export BAKERY_PROPELLER_CREDS="usr:pws"

Note that `BAKERY_ORIGIN_HOST` will be the base URL of your manifest files.

#### Run the API:

    $ make run

The API will be available on http://localhost[:BAKERY_HTTP_PORT]

## Run Tests

    $ make  test

## Help

You can find the source code for Bakery at GitHub:
[bakery][bakery]

[bakery]: https://github.com/cbsinteractive/bakery

If you have any questions regarding Bakery, please reach out in the [#i-vidtech-mediahub](slack://channel?team={cbs}&id={i-vidtech-mediahub}) channel.
