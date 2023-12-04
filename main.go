package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

var (
	macFilter string
	debugMode bool
	hciID     int

	temp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "temp",
		Help: "Current temperature",
	}, []string{"mac"})

	humidity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "humidity",
		Help: "Current humidity %",
	}, []string{"mac"})

	battery = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "battery",
		Help: "Current battery %",
	}, []string{"mac"})

	payloadsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "payloads_received",
		Help: "Count of payloads received by this mac",
	}, []string{"mac"})

	lastPayloadReceivedTs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "last_payload_received_ts",
		Help: "Payload last received timestamp in unix",
	}, []string{"mac"})
)

type BLEPayload struct {
	Temp     float64
	Humidity float64
	Battery  float64
	Time     time.Time
	Mac      string
}

func main() {
	flag.BoolVar(&debugMode, "debugMode", false, "display debugging log messages")
	flag.StringVar(&macFilter, "macFilter", "", "filter to specific mac addresses")
	flag.IntVar(&hciID, "hciID", 0, "define a different hci id")

	flag.Parse()

	if debugMode {
		log.SetLevel(log.DebugLevel)
	}

	setupCloseHandler()

	log.Printf("Starting")
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), time.Second*30))

	var device ble.Device
	var err error

	go startHttpServer()

	device, err = linux.NewDevice(ble.OptDeviceID(hciID))
	if err != nil {
		log.Panic(err)
	}

	ble.SetDefaultDevice(device)

	var filterFunc func(a ble.Advertisement) bool

	if macFilter != "" {
		macs := strings.Split(macFilter, ",")
		filterFunc = func(a ble.Advertisement) bool {
			for _, m := range macs {
				if strings.TrimSpace(m) == a.Addr().String() {
					return true
				}
			}

			log.Debugf("mac broadcast not whitelisted: %s", a.Addr().String())
			return false
		}
	}

	for {
		_ = ble.Scan(ctx, false, func(a ble.Advertisement) {
			logFields := log.Fields{
				"mac": a.Addr().String(),
			}

			if len(a.ServiceData()) != 1 {
				log.WithFields(logFields).Debugf("Unexpected device - no servicedata")

				return
			}

			sData := a.ServiceData()[0]
			logFields["uuid"] = sData.UUID.String()

			if sData.UUID.String() != "181a" {
				log.WithFields(logFields).Debugf("Unexpected device - incorrect uuid")

				return
			}

			var macParts []string

			for _, m := range sData.Data[:6] {
				s := strconv.FormatInt(int64(m), 16)
				macParts = append(macParts, s)
			}

			measurement := BLEPayload{
				Mac:      strings.Join(macParts, ":"),
				Temp:     float64(sData.Data[7]),
				Humidity: float64(sData.Data[8]),
				Battery:  float64(sData.Data[9]),
				Time:     time.Now(),
			}

			// ok good, we got the measurement so send it away in a
			// coroutine - failures wont block collection
			go submitMeasurement(logFields, &measurement)
		}, filterFunc)
	}
}

func startHttpServer() {
	http.Handle("/metrics", promhttp.Handler())
	_ = http.ListenAndServe("0.0.0.0:2112", nil)

	log.Printf("starting http server")
	log.Printf("host: http://0.0.0.0:2112\nmetrics: http://0.0.0.0:2112/metrics")
}

func setupCloseHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}

func submitMeasurement(fields log.Fields, p *BLEPayload) {
	if p == nil {
		return
	}

	temp.WithLabelValues(p.Mac).Set(p.Temp)
	humidity.WithLabelValues(p.Mac).Set(p.Humidity)
	battery.WithLabelValues(p.Mac).Set(p.Battery)

	payloadsReceived.WithLabelValues(p.Mac).Inc()
	lastPayloadReceivedTs.WithLabelValues(p.Mac).Set(float64(p.Time.Unix()))

	log.WithFields(fields).Printf(
		"Time: %s - Temp: %2.2fc - Humidity: %2.2f%% - Battery: %2.2f%%",
		p.Time.Format("15:04:05"),
		p.Temp / 10,
		p.Humidity,
		p.Battery,
	)
}
