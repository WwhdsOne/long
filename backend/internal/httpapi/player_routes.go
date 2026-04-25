package httpapi

import "github.com/cloudwego/hertz/pkg/route"

func registerPlayerActionRoutes(router route.IRouter, options Options) {
	registerButtonClickRoutes(router, options)
	registerAutoClickRoutes(router, options)
	registerEquipmentRoutes(router, options)
}
