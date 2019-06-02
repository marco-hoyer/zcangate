package common

import (
	"log"
	"strconv"
	"github.com/marco-hoyer/zcangate/can"
)

type Type struct {
	name, unit     string
	transformation func(string) float64
}

type Measurement struct {
	name, unit string
	value      float64
}

func transformSmallNumber(s string) float64 {
	v, _ := strconv.ParseInt(s[0:2], 16, 64)
	return float64(v)
}

func transformTemperature(s string) float64 {
	v1, _ := strconv.ParseInt(s[0:2], 16, 64)
	v2, _ := strconv.ParseInt(s[2:4], 16, 64)

	value := float64((v1 + v2*255) / 10)
	log.Println("Transformed temperature value '", s, "' into: ", value)
	return value
}

func transformAirVolume(s string) float64 {
	v1, _ := strconv.ParseInt(s[0:2], 16, 64)
	v2, _ := strconv.ParseInt(s[2:4], 16, 64)
	return float64(v1 + v2*255)
}

func transformAny(s string) float64 {
	word := uint8(0)
	for n := range s {
		word += s[n] << uint(n*8)
	}
	return float64(word)
}

var mapping = map[int]Type{
	16: {
		name:           "z_unknown_NwoNode",
		unit:           "",
		transformation: transformAny,
	},
	17: {
		name:           "z_unknown_NwoNode",
		unit:           "",
		transformation: transformAny,
	},
	18: {
		name:           "z_unknown_NwoNode",
		unit:           "",
		transformation: transformAny,
	},
	65: {
		name:           "ventilation_level",
		unit:           "level",
		transformation: transformSmallNumber,
	},
	66: {
		name:           "bypass_state",
		unit:           "0=auto,1=open,2=close",
		transformation: transformSmallNumber,
	},
	81: {
		name:           "Timer1",
		unit:           "s",
		transformation: transformAny,
	},
	82: {
		name:           "Timer2",
		unit:           "s",
		transformation: transformAny,
	},
	83: {
		name:           "Timer3",
		unit:           "s",
		transformation: transformAny,
	},
	84: {
		name:           "Timer4",
		unit:           "s",
		transformation: transformAny,
	},
	85: {
		name:           "Timer5",
		unit:           "s",
		transformation: transformAny,
	},
	86: {
		name:           "Timer6",
		unit:           "s",
		transformation: transformAny,
	},
	87: {
		name:           "Timer7",
		unit:           "s",
		transformation: transformAny,
	},
	88: {
		name:           "Timer8",
		unit:           "s",
		transformation: transformAny,
	},

	96: {
		name:           "bypass ??? ValveMsg",
		unit:           "unknown",
		transformation: transformAny,
	},
	97: {
		name:           "bypass_b_status",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	98: {
		name:           "bypass_a_status",
		unit:           "unknown",
		transformation: transformAirVolume,
	},

	115: {
		name:           "ventilator enabled output",
		unit:           "",
		transformation: transformAny,
	},
	116: {
		name:           "ventilator enabled input",
		unit:           "",
		transformation: transformAny,
	},
	117: {
		name:           "ventilator power_percent output",
		unit:           "%",
		transformation: transformSmallNumber,
	},
	118: {
		name:           "ventilator power_percent input",
		unit:           "%",
		transformation: transformSmallNumber,
	},
	119: {
		name:           "ventilator air_volume output",
		unit:           "m3",
		transformation: transformAirVolume,
	},
	120: {
		name:           "ventilator air_volume input",
		unit:           "m3",
		transformation: transformAirVolume,
	},
	121: {
		name:           "ventilator speed output",
		unit:           "rpm",
		transformation: transformAirVolume,
	},
	122: {
		name:           "ventilator speed input",
		unit:           "rpm",
		transformation: transformAirVolume,
	},
	128: {
		name:           "Power_consumption_actual",
		unit:           "W",
		transformation: transformSmallNumber,
	},
	129: {
		name:           "Power_consumption_this_year",
		unit:           "kWh",
		transformation: transformAirVolume,
	},
	130: {
		name:           "Power_consumption_lifetime",
		unit:           "kWh",
		transformation: transformAirVolume,
	},
	144: {
		name:           "Power PreHeater this year",
		unit:           "kWh",
		transformation: transformAny,
	},
	145: {
		name:           "Power PreHeater total",
		unit:           "kWh",
		transformation: transformAny,
	},
	146: {
		name:           "Power PreHeater actual",
		unit:           "W",
		transformation: transformAny,
	},
	192: {
		name:           "days_until_next_filter_change",
		unit:           "days",
		transformation: transformAirVolume,
	},

	208: {
		name:           "z_Unknown_TempHumConf",
		unit:           "",
		transformation: transformAny,
	},
	209: {
		name:           "RMOT",
		unit:           "°C",
		transformation: transformTemperature,
	},
	210: {
		name:           "z_Unknown_TempHumConf",
		unit:           "",
		transformation: transformAny,
	},
	211: {
		name:           "z_Unknown_TempHumConf",
		unit:           "",
		transformation: transformAny,
	},
	212: {
		name:           "Target_temperature",
		unit:           "°C",
		transformation: transformTemperature,
	},
	213: {
		name:           "Power_avoided_heating_actual",
		unit:           "W",
		transformation: transformAny,
	},
	214: {
		name:           "Power_avoided_heating_this_year",
		unit:           "kWh",
		transformation: transformAirVolume,
	},
	215: {
		name:           "Power_avoided_heating_lifetime",
		unit:           "kWh",
		transformation: transformAirVolume,
	},
	216: {
		name:           "Power_avoided_cooling_actual",
		unit:           "W",
		transformation: transformAny,
	},
	217: {
		name:           "Power_avoided_cooling_this_year",
		unit:           "kWh",
		transformation: transformAirVolume,
	},
	218: {
		name:           "Power_avoided_cooling_lifetime",
		unit:           "kWh",
		transformation: transformAirVolume,
	},
	219: {
		name:           "Power PreHeater Target",
		unit:           "W",
		transformation: transformAny,
	},
	220: {
		name:           "temperature_inlet_before_preheater",
		unit:           "°C",
		transformation: transformTemperature,
	},
	221: {
		name:           "temperature_inlet_after_recuperator",
		unit:           "°C",
		transformation: transformTemperature,
	},
	222: {
		name:           "z_Unknown_TempHumConf",
		unit:           "",
		transformation: transformAny,
	},
	224: {
		name:           "z_Unknown_VentConf",
		unit:           "",
		transformation: transformAny,
	},
	225: {
		name:           "z_Unknown_VentConf",
		unit:           "",
		transformation: transformAny,
	},
	226: {
		name:           "z_Unknown_VentConf",
		unit:           "",
		transformation: transformAny,
	},
	227: {
		name:           "bypass_open",
		unit:           "%",
		transformation: transformSmallNumber,
	},
	228: {
		name:           "frost_disbalance",
		unit:           "%",
		transformation: transformSmallNumber,
	},
	229: {
		name:           "z_Unknown_VentConf",
		unit:           "",
		transformation: transformAny,
	},
	230: {
		name:           "z_Unknown_VentConf",
		unit:           "",
		transformation: transformAny,
	},

	256: {
		name:           "z_Unknown_NodeConf",
		unit:           "unknown",
		transformation: transformAny,
	},
	257: {
		name:           "z_Unknown_NodeConf",
		unit:           "unknown",
		transformation: transformAny,
	},

	273: {
		name:           "temperature_something...",
		unit:           "°C",
		transformation: transformTemperature,
	},
	274: {
		name:           "temperature_outlet_before_recuperator",
		unit:           "°C",
		transformation: transformTemperature,
	},
	275: {
		name:           "temperature_outlet_after_recuperator",
		unit:           "°C",
		transformation: transformTemperature,
	},
	276: {
		name:           "temperature_inlet_before_preheater",
		unit:           "°C",
		transformation: transformTemperature,
	},
	277: {
		name:           "temperature_inlet_before_recuperator",
		unit:           "°C",
		transformation: transformTemperature,
	},
	278: {
		name:           "temperature_inlet_after_recuperator",
		unit:           "°C",
		transformation: transformTemperature,
	},

	289: {
		name:           "z_unknown_HumSens",
		unit:           "",
		transformation: transformAny,
	},
	290: {
		name:           "air_humidity_outlet_before_recuperator",
		unit:           "%",
		transformation: transformSmallNumber,
	},
	291: {
		name:           "air_humidity_outlet_after_recuperator",
		unit:           "%",
		transformation: transformSmallNumber,
	},
	292: {
		name:           "air_humidity_inlet_before_preheater",
		unit:           "%",
		transformation: transformSmallNumber,
	},
	293: {
		name:           "air_humidity_inlet_before_recuperator",
		unit:           "%",
		transformation: transformSmallNumber,
	},
	294: {
		name:           "air_humidity_inlet_after_recuperator",
		unit:           "%",
		transformation: transformSmallNumber,
	},

	305: {
		name:           "PresSens_exhaust",
		unit:           "Pa",
		transformation: transformAny,
	},
	306: {
		name:           "PresSens_inlet",
		unit:           "Pa",
		transformation: transformAny,
	},

	369: {
		name:           "z_Unknown_AnalogInput",
		unit:           "V?",
		transformation: transformAny,
	},
	370: {
		name:           "z_Unknown_AnalogInput",
		unit:           "V?",
		transformation: transformAny,
	},
	371: {
		name:           "z_Unknown_AnalogInput",
		unit:           "V?",
		transformation: transformAny,
	},
	372: {
		name:           "z_Unknown_AnalogInput",
		unit:           "V?",
		transformation: transformAny,
	},
	400: {
		name:           "z_Unknown_PostHeater_ActualPower",
		unit:           "W",
		transformation: transformAny,
	},
	401: {
		name:           "z_Unknown_PostHeater_ThisYear",
		unit:           "kWh",
		transformation: transformAny,
	},
	402: {
		name:           "z_Unknown_PostHeater_Total",
		unit:           "kWh",
		transformation: transformAny,
	},
}

func ToMeasurement(frame can.CanBusFrame) Measurement {
	dataType, found := mapping[frame.Pdu]
	if found {
		//log.Printf("Mapping found for: %s | %d | %s | %s\r", address, length, data, dataType.name)
		return Measurement{
			name:  dataType.name,
			unit:  dataType.unit,
			value: dataType.transformation(frame.Data),
		}
	} else {
		log.Printf("Unknown message: %s | %d | %s\r", frame.Id, frame.Length, frame.Data)
		return Measurement{}
	}

	return Measurement{}
}
