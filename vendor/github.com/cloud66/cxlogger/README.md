# cxlogger
Cloud 66 Go Wrapper for Logging

Usage:
```
import github.com/cloud66/cxlogger

func main() {
    
    out := "STDOUT"  // (NONE|STDOUT|file_path)
    level := "debug" // (debug|info|warn|error|crit)
    cxlogger.Initialize(out, level)
  
    string_value := "some message"
    cxlogger.Debug(string_value)
    cxlogger.Info(string_value)
    cxlogger.Warn(string_value)
    cxlogger.Error(string_value)
    cxlogger.Crit(string_value)
    
    // format_value := "some message with %s"
    // cxlogger.Debugf(format_value, params)
    // cxlogger.Infof(format_value, params)
    // cxlogger.Warnf(format_value, params)
    // cxlogger.Errorf(format_value, params)
    // cxlogger.Critf(format_value, params)
    
    error_value := errors.New("Sample error")
    cxlogger.Debug(error_value)
    cxlogger.Info(error_value)
    cxlogger.Warn(error_value)
    cxlogger.Error(error_value)
    cxlogger.Crit(error_value)

	// Add Indentation
	cxlogger.Debug("Beginning of my function") // The output is:
	cxlogger.IncreaseIndentation()             // 
	cxlogger.Debug("blah blah blah")           // Beginning of my function
	cxlogger.DecreaseIndentation()             //   blah blah blah
	cxlogger.Debug("End of my function")       // End of my function

	// Manual indentation
	cxlogger.DebugIndent(1, "This message will have one level of indentation")

	// It is possible to use different contexts
	cxlogger.Context = "Thread1"
	// or to create a new logger for each context
	mylogger := cxlogger.NewWithContext("Thread2", out, level)

}
```

