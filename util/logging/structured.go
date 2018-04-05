package logging

import (
	"github.com/opentable/sous/util/logging/constants"
	"github.com/pborman/uuid"
)

// LogMessage records a message to one or more structured logs
func (ls LogSet) LogMessage(lvl Level, msg LogMessage) {
	logto := ls.logrus.WithField("severity", lvl.String())

	ls.eachField(func(name constants.FieldName, value interface{}) {
		logto = logto.WithField(string(name), value)
	})

	msg.EachField(func(name constants.FieldName, value interface{}) {
		enforceSchema(name, value)
		logto = logto.WithField(string(name), value)
	})

	logto.Message = msg.Message()
	err := ls.dumpBundle.sendToKafka(lvl, logto)
	if _, isKafkaSend := msg.(*kafkaSendErrorMessage); err != nil && !isKafkaSend {
		reportKafkaSendError(ls, err)
	}

	switch lvl {
	default:
		logto.Printf("unknown Level: %d - %q", lvl, msg.Message())
	case CriticalLevel:
		logto.Error(msg.Message())
	case WarningLevel:
		logto.Warn(msg.Message())
	case InformationLevel:
		logto.Info(msg.Message())
	case DebugLevel:
		logto.Debug(msg.Message())
	case ExtraDebug1Level:
		logto.Debug(msg.Message())
	}
}
func (ls LogSet) eachField(f FieldReportFn) {
	if ls.appRole != "" {
		f("component-id", "sous-"+ls.appRole)
	} else {
		f("component-id", "sous")
	}
	f("logger-name", ls.name)
	f("@uuid", uuid.New())

	ls.appIdent.EachField(f)
}

func enforceSchema(name constants.FieldName, val interface{}) {
	if false {
		panic("bad logging")
	}
}
