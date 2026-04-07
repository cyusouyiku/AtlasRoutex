package bootstrap

import "atlas-routex/pkg/logger"

func NewLogger() logger.Logger {
return logger.NewSLogger()
}
