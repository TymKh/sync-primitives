# Sync primitives

## Spike Manager

The spike manager is a simple tool to manage the synchronization of multiple requests under the load-spikes.
It is supposed to work with the data, that is valid for some time (e.g 5 milliseconds or 1 hour) as well as static data that 
is valid for the lifetime of service. 

Spike manager is a good fit for you if you need to store data locally (in-memory for your service) to better/faster serve user requests.
If it is the case, and performance of external requests is an issue in your project (e.g you can't afford call SQL database on each user request) then you should try it out.

Spike manager guarantees that within timeframe only one (presumably first) external request (say time-consuming or resource-consuming task)
will be executed and the result will be cached for configured time. All other requests that fall within this timeframe (during spike) will be served
with response cached in the spike manager from first execution. Once data is expired, the next request will be executed and the result will be cached again.

This makes `spike.Manager` quite versatile, you should check tests in this repository (both in `spike` package and `exapmles` package). 

## Usage

Common use case would be fetching some data like `prices`, `currency rates`, static info like `zip codes`.
One interesting use case is to use it as a `LazyLoader` for config-like information. You can catch data that you need on the fly
and cache it for the service lifetime. That could be useful if space of data is possibly very big, but you need only small sample of it on each microservice
instance. Not prefetching anything, but reading only that data that you need right now could work pretty good for some load characters. Example
of this can be seen in `examples/lazy_loader_test.go` file.

### Manager vs Custom manager
You should default to using `Manager`, since it uses well-known cache library and is easy to configure on client-side 
(only requires one closure function that is used to read data from external source)

Custom manager should be used if you want to integrate caches that are used in your project for uniformity. You can also
augment default flow with some pre-fetching logic on a time basis (say every 5 minutes you want to fetch hot data since you know it will eventually be requested, and you have
another batch-endpoint to fetch all data at once), for this to happen - you could use your cache in some cron-like service as well as share it with `spike.Manager`



