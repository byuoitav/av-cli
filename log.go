package avcli

type Logger interface {
	Debugf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

func (s *Server) debugf(format string, a ...interface{}) {
	if s.Logger != nil {
		s.Logger.Debugf(format, a...)
	}
}

func (s *Server) infof(format string, a ...interface{}) {
	if s.Logger != nil {
		s.Logger.Infof(format, a...)
	}
}

func (s *Server) warnf(format string, a ...interface{}) {
	if s.Logger != nil {
		s.Logger.Warnf(format, a...)
	}
}

func (s *Server) errorf(format string, a ...interface{}) {
	if s.Logger != nil {
		s.Logger.Errorf(format, a...)
	}
}
