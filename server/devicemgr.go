package server

import (
	"github.com/ghetzel/canibus/core"
)

func InitDrivers() {
	var conf = core.GetConfig()
	var drivers = conf.GetDrivers()
	for i := range drivers {
		drivers[i].Init()
	}
}
