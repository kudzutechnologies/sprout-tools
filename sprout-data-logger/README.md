# Data Logger

> A utility that can be used as a sprout HTTP target for collecting data locally.

## Installation

If you have `go` installed in your system, simply use:

```
go get github.com/kudzutechnologies/sprout-tools/sprout-data-logger
```

## Usage

```
Usage of ./sprout-data-logger:
  -file string
    	the .csv file where to log the received data (default "data-202010261227.csv")
  -port int
    	the port whre to listen for data (default 8090)
```

This utility starts a local HTTP server that can be used as a target from your sprout to receive arbitrary data points from sprout.

### Sprout Configuration

For configure your sprout to use the data logger do:

* On the _WiFi Uplink_ module:
  * Set **Framing** to `HTTP(S) POST with raw body`
  * Set **Security** to `Disabled`
  * Set **Host** to your machine's IP address. The tool will report it when started.
  * Set **Port** to `8090`

* On the _Sensor Hub_ module:
  * Set **Encoder** to `JSON String`

## Example Script

The following script can be used in the javascript IDE to push data 

```js
// When activated, schedule the system to take samples every second
function activate() {
  setInterval(takeSample, 1000);
}

// The sampling function will ask all the sensors in the system to
// provide their values and then flush the buffer immediately.
function takeSample() {
  sampleAllSensors();
  samplesFlush();
}

// Here you can define your own sensor values
onEvent("sensors.sample", function() {
  // ... 
  sampleAdd("my-sensor", {
    humidity: read_humidity(),
    temp: read_temperature()
  });
  // ... 
});
```