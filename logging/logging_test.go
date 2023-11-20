package logging

import (
	"log"
	"testing"
)

func TestSetup(t *testing.T) {
	if err := Setup(&Opts{
		LogFile: "/non/existing/path/to/a/log/file",
	}); err == nil {
		t.Errorf("Setup() with a non existing path = nil, want error")
	}
}

func TestLoggingFlags(t *testing.T) {
	for _, test := range []struct {
		logDate  bool
		logTime  bool
		logMsec  bool
		logUTC   bool
		wantFlag int
	}{
		{
			// Date only
			logDate:  true,
			wantFlag: log.Ldate,
		},
		{
			// Date and time
			logDate:  true,
			logTime:  true,
			wantFlag: log.Ldate | log.Ltime,
		},
		{
			// Date, time and msec (time is implied by msec)
			logDate:  true,
			logMsec:  true,
			wantFlag: log.Ldate | log.Ltime | log.Lmicroseconds,
		},
		{
			// Also UTC
			logDate:  true,
			logMsec:  true,
			logUTC:   true,
			wantFlag: log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC,
		},
	} {
		opts := &Opts{
			LogDate: test.logDate,
			LogTime: test.logTime,
			LogMsec: test.logMsec,
			LogUTC:  test.logUTC,
		}
		if gotFlag := loggingFlags(opts); gotFlag != test.wantFlag {
			t.Errorf("loggingFlags(%+v) = %v, want %v", opts, gotFlag, test.wantFlag)
		}
	}
}
