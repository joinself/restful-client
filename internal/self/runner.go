package self

import (
	"github.com/joinself/restful-client/pkg/log"
)

// RunService executes the listeners specified on the Service.
func RunService(service Service, logger log.Logger) {
	service.Run()
}
