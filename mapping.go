package main

import (
	"log"
)

//@staticmethod
//def _to_can_message(frame: bytes) -> Message:
//try:
//frame = frame.strip(b'\x07')
//
//type = frame[0:1].decode("ascii")
//assert type.lower() in ["t", "r"], "message type must be T or R, not '{}'".format(type)
//
//id = frame[1:9].decode("ascii")
//length = int(frame[9:10], 16)
//
//index = 10
//data = []
//
//for item in range(0, length):
//data.append(int(frame[index:index + 2], 16))
//index += 2
//
//return Message(type, id, length, data, frame)
//except Exception as e:
//print("Could not parse can frame '{}', error was: {}".format(frame, e))

func transformFloat(s string) string {
	return s
}

func transformTemperature(s string) string {
	return s
}

func transformAirVolume(s string) string {
	return s
}

type Type struct {
	name, unit     string
	transformation func(string) string
}

var mapping = map[string]Type{
	"00148041": {
		name:           "unknown_decreasing_number",
		unit:           "unknown",
		transformation: transformFloat,
	},
	"00454041": {
		name:           "temperature_inlet_before_recuperator",
		unit:           "°C",
		transformation: transformTemperature,
	},
	"00458041": {
		name:           "temperature_inlet_after_recuperator",
		unit:           "°C",
		transformation: transformTemperature,
	},
	"001E0041": {
		name:           "air_volume_input_ventilator",
		unit:           "m3",
		transformation: transformAirVolume,
	},
	"001DC041": {
		name:           "air_volume_output_ventilator",
		unit:           "m3",
		transformation: transformAirVolume,
	},
	"001E8041": {
		name:           "speed_input_ventilator",
		unit:           "rpm",
		transformation: transformAirVolume,
	},
	"001E4041": {
		name:           "speed_output_ventilator",
		unit:           "rpm",
		transformation: transformAirVolume,
	},
	"00488041": {
		name:           "air_humidity_outlet_before_recuperator",
		unit:           "%",
		transformation: transformFloat,
	},
	"0048C041": {
		name:           "air_humidity_inlet_before_recuperator",
		unit:           "%",
		transformation: transformFloat,
	},
	"00200041": {
		name:           "total_power_consumption",
		unit:           "W",
		transformation: transformFloat,
	},
	"001D4041": {
		name:           "power_percent_output_ventilator",
		unit:           "%",
		transformation: transformFloat,
	},
	"001D8041": {
		name:           "power_percent_input_ventilator",
		unit:           "%",
		transformation: transformFloat,
	},
	"0082C042": {
		name:           "0082C042",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"004C4041": {
		name:           "004C4041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00384041": {
		name:           "00384041",
		unit:           "unknown",
		transformation: transformFloat,
	},
	"00144041": {
		name:           "remaining_s_in_currend_ventilation_mode",
		unit:           "s",
		transformation: transformAirVolume,
	},
	"00824042": {
		name:           "00824042",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00810042": {
		name:           "00810042",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00208041": {
		name:           "total_power_consumption",
		unit:           "kWh",
		transformation: transformAirVolume,
	},
	"00344041": {
		name:           "00344041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00370041": {
		name:           "00370041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00300041": {
		name:           "days_until_next_filter_change",
		unit:           "days",
		transformation: transformAirVolume,
	},
	"00044041": {
		name:           "00044041",
		unit:           "unknown",
		transformation: transformFloat,
	},
	"00204041": {
		name:           "total_power_consumption_this_year",
		unit:           "kWh",
		transformation: transformAirVolume,
	},
	"00084041": {
		name:           "00084041",
		unit:           "unknown",
		transformation: transformFloat,
	},
	"00804042": {
		name:           "00804042",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00644041": {
		name:           "00644041",
		unit:           "unknown",
		transformation: transformFloat,
	},
	"00354041": {
		name:           "00354041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00390041": {
		name:           "frost_disbalance",
		unit:           "%",
		transformation: transformFloat,
	},

	"0035C041": {
		name:           "total_power_savings",
		unit:           "kWh",
		transformation: transformAirVolume,
	},

	"0044C041": {
		name:           "0044C041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},

	"0080C042": {
		name:           "0080C042",
		unit:           "unknown",
		transformation: transformAirVolume,
	},

	"000E0041": {
		name:           "000E0041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00604041": {
		name:           "00604041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00450041": {
		name:           "00450041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},

	"00378041": {
		name:           "00378041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00818042": {
		name:           "00818042",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00820042": {
		name:           "00820042",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"001D0041": {
		name:           "001D0041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00490041": {
		name:           "00490041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00350041": {
		name:           "00350041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},

	"0081C042": {
		name:           "0081C042",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00448041": {
		name:           "00448041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00560041": {
		name:           "00560041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00374041": {
		name:           "00374041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},

	"00808042": {
		name:           "00808042",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00040041": {
		name:           "00040041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"10040001": {
		name:           "10040001",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00120041": {
		name:           "00120041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00688041": {
		name:           "00688041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00358041": {
		name:           "total_power_savings_this_year",
		unit:           "kWh",
		transformation: transformAirVolume,
	},

	"ventilation_level": {
		name:           "00104041",
		unit:           "ventilation_level",
		transformation: transformFloat,
	},

	"00544041": {
		name:           "00544041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00498041": {
		name:           "00498041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00814042": {
		name:           "00814042",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"000C4041": {
		name:           "000C4041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00828042": {
		name:           "00828042",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"00494041": {
		name:           "00494041",
		unit:           "unknown",
		transformation: transformFloat,
	},

	"004C8041": {
		name:           "004C8041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00388041": {
		name:           "00388041",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00188041": {
		name:           "bypass_a_status",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00184041": {
		name:           "bypass_b_status",
		unit:           "unknown",
		transformation: transformAirVolume,
	},
	"00108041": {
		name:           "bypass_state",
		unit:           "0=auto,1=open,2=close",
		transformation: transformFloat,
	},
	"0038C041": {
		name:           "bypass_open",
		unit:           "%",
		transformation: transformFloat,
	},
}

func toType(value string) Type {
	prefix := string(value[0])
	address := value[1:9]
	length := string(value[9])
	data := string(value[10:])
	//
	log.Printf("PARSED: %s | %s | %s | %s\r", prefix, address, length, data)

	return Type{}
}
