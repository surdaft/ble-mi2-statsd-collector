# Psychedelic Flamingo

> I couldn't think of a good name

Pick up BLE broadcasts from the [Xiaomi Temperature Sensor](https://www.amazon.co.uk/Mi-Temperature-Humidity-Monitor-2/dp/B08C7KVDJW/ref=sr_1_5?keywords=xiaomi+temperature+sensor&qid=1655562676&sprefix=xiaomi+temp%2Caps%2C125&sr=8-5) flashed with the [custom firmware](https://github.com/atc1441/ATC_MiThermometer) 

## Flags

| name           | value required | default | description                                                                           |
|----------------|----------------|---------|---------------------------------------------------------------------------------------|
| -debugMode     | no             | false   | Enable extra debugging info, good for identifying which device you want to filter for |
| -macFilter     | yes            | empty   | filter for the broadcasts from a specific mac device                                  |
| -statsdHost    | yes            | empty   | provide a statsd host to ping info to                                                 |
| -statsdPrefix  | yes            | empty   | provide a prefix for stat metrics, a good example would be room name                  |

## Statsd

There is a parameter defined to allow you to send these metrics to your own statsd server.

Just make sure to divide the temperature by 10 before using it. This is because in golang the gauge method requires an int64. To do this for Grafana you can:

1. add a transformation 
2. select the transformation type to be "Add field from calculation"
3. select "Binary operation"
4. input `averageSeries(...)` / `10` 
5. select "replace all fields" checkbox
6. update the "Standard options" within the sidebar to specify the unit as Celsius or Fahrenheit