package driver

import (
	"fmt"

	"quark-mobile/internal/driver/mobile"
	"quark-mobile/internal/driver/openlist"
	"quark-mobile/internal/driver/quark"
	"quark-mobile/internal/model"
)

var drivers = map[model.DriverType]Driver{}
var openlistClient *openlist.Client

func InitDrivers(client *openlist.Client) {
	openlistClient = client
	drivers[model.DriverQuark] = quark.NewQuarkDriver(client)
	drivers[model.DriverMobile] = mobile.NewMobileDriver(client)
}

func GetDriver(d model.DriverType) (Driver, error) {
	drv, ok := drivers[d]
	if !ok {
		return nil, fmt.Errorf("unknown driver: %s", d)
	}
	return drv, nil
}

func GetOpenListClient() *openlist.Client {
	return openlistClient
}

func RegisteredDrivers() []model.DriverType {
	result := make([]model.DriverType, 0, len(drivers))
	for k := range drivers {
		result = append(result, k)
	}
	return result
}
