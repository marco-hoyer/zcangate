package can

type Command struct {
	Fragmentation int
	Code          string
}

var Commands = map[string]Command{
	"auto_mode": {
		Fragmentation: 0x0,
		Code:          "85150801",
	},
	"manual_mode": {
		Fragmentation: 0x1,
		Code:          "84150801000000000100000001",
	},
	"ventilation_level_3": {
		Fragmentation: 0x1,
		Code:          "8415010100000000FFFFFFFF03",
	},
	"ventilation_level_2": {
		Fragmentation: 0x1,
		Code:          "8415010100000000FFFFFFFF02",
	},
	"ventilation_level_1": {
		Fragmentation: 0x1,
		Code:          "8415010100000000FFFFFFFF01",
	},
	"ventilation_level_0": {
		Fragmentation: 0x1,
		Code:          "8415010100000000FFFFFFFF00",
	},
	"ventilation_mode_supply_only_1h": {
		Fragmentation: 0x1,
		Code:          "8415060100000000100E000001",
	},
	"ventilation_mode_outlet_only_1h": {
		Fragmentation: 0x1,
		Code:          "8415070100000000100E000001", //
	},
	"ventilation_mode_balanced": {
		Fragmentation: 0x0,
		Code:          "85150601",
	},
	"temperature_profile_normal": {
		Fragmentation: 0x1,
		Code:          "8415030100000000ffffffff00",
	},
	"temperature_profile_cool": {
		Fragmentation: 0x1,
		Code:          "8415030100000000ffffffff01",
	},
	"temperature_profile_warm": {
		Fragmentation: 0x1,
		Code:          "8415030100000000ffffffff02",
	},
	"bypass_activated_1h": {
		Fragmentation: 0x1,
		Code:          "8415020100000000100e000001",
	},
	"bypass_deactivated_1h": {
		Fragmentation: 0x1,
		Code:          "8415020100000000100e000002",
	},
	"bypass_auto": {
		Fragmentation: 0x0,
		Code:          "85150201",
	},
	"passive_temperature_control_off": {
		Fragmentation: 0x0,
		Code:          "031d010400",
	},
	"passive_temperature_control_auto_only": {
		Fragmentation: 0x0,
		Code:          "031d010401",
	},
	"passive_temperature_control_on": {
		Fragmentation: 0x0,
		Code:          "031d010402",
	},
	"passive_humidity_control_off": {
		Fragmentation: 0x0,
		Code:          "031d010600",
	},
	"passive_humidity_control_auto_only": {
		Fragmentation: 0x0,
		Code:          "031d010601",
	},
	"passive_humidity_control_on": {
		Fragmentation: 0x0,
		Code:          "031d010602",
	},
	"humidity_protection_off": {
		Fragmentation: 0x0,
		Code:          "031d010700",
	},
	"humidity_protection_auto_only": {
		Fragmentation: 0x0,
		Code:          "031d010701",
	},
	"humidity_protection_on": {
		Fragmentation: 0x0,
		Code:          "031d010702",
	},
}
