# Psychedelic Flamingo

> I couldn't think of a good name

Pick up BLE broadcasts from the [Xiaomi Temperature Sensor](https://www.amazon.co.uk/Mi-Temperature-Humidity-Monitor-2/dp/B08C7KVDJW/ref=sr_1_5?keywords=xiaomi+temperature+sensor&qid=1655562676&sprefix=xiaomi+temp%2Caps%2C125&sr=8-5) flashed with the [custom firmware](https://github.com/atc1441/ATC_MiThermometer) 

## Flags

| name           | value required | default | description                                                                           |
|----------------|----------------|---------|---------------------------------------------------------------------------------------|
| -debugMode     | no             | false   | Enable extra debugging info, good for identifying which device you want to filter for |
| -macFilter     | yes            | empty   | filter for the broadcasts from a specific mac device                                  |
| -hciID         | yes            | 0       | which bluetooth device to use, `hcitool dev` to list devices                          |

## Prometheus

Prometheus exports on port 2112

Just make sure to divide the temperature by 10 before using it. This is because in golang the gauge method requires an int64.
To do this for Grafana you can just divide the field by 10 directly: `temp{mac="xx:xx:xx:xx:xx"} / 10`
