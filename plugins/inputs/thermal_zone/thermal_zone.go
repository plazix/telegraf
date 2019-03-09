package thermal_zone

import (
	"fmt"
	"github.com/influxdata/telegraf/plugins/parsers"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ThermalZoneItem struct {
	atype     string
	temp_path string
}

type ThermalZone struct {
	zones map[string]ThermalZoneItem

	sync.Mutex
}

func NewThermalZone() *ThermalZone {
	return &ThermalZone{}
}

const sampleConfig = `
  ## The dataformat to be read from files
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (tz *ThermalZone) SampleConfig() string {
	return sampleConfig
}

func (tz *ThermalZone) Description() string {
	return "Read metrics about temperature from /sys/class/thermal/thermal_zone"
}

func (tz *ThermalZone) Gather(acc telegraf.Accumulator) error {
	tz.Lock()
	defer tz.Unlock()

	for zone, info := range tz.zones {
		tempContents, err := ioutil.ReadFile(info.temp_path)
		if err != nil {
			return fmt.Errorf("E! Error file: %v could not be read, %s", info.temp_path, err)
		}

		temp, err := strconv.Atoi(strings.TrimSpace(string(tempContents)))
		if err != nil {
			return fmt.Errorf("E! Error convert thermal_zone temp: %v could not be convert to int, %s",
				tempContents, err)
		}

		acc.AddFields(
			"thermal_zone",
			map[string]interface{}{"value": temp},
			map[string]string{"zone": zone, "type": info.atype})
	}

	return nil
}

func (tz *ThermalZone) Start(acc telegraf.Accumulator) error {
	tz.Lock()
	defer tz.Unlock()

	tz.zones = make(map[string]ThermalZoneItem)

	err := filepath.Walk("/sys/class/thermal", func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "thermal_zone") {
			var type_filename = filepath.Join(path, "type")

			atype, err := ioutil.ReadFile(type_filename)
			if err != nil {
				return fmt.Errorf("E! Error file: %v could not be read, %s", type_filename, err)
			}

			tz.zones[info.Name()] = ThermalZoneItem{
				atype: strings.TrimSpace(string(atype)),
				temp_path: filepath.Join(path, "temp"),
			}
		}

		return nil
	})

	return err
}

func (tz *ThermalZone) Stop() {
	tz.Lock()
	defer tz.Unlock()

	tz.zones = nil
}

func (tz *ThermalZone) SetParserFunc(fn parsers.ParserFunc) {
}

func init() {
	inputs.Add("thermal_zone", func() telegraf.Input {
		return NewThermalZone()
	})
}
