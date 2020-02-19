package zerolog

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/ptechen/zerolog/diode"
	"os"
	"time"
)
var newLogFileName string

// output Example
func output(newLogFileName string, logger *Logger) {
	if newLogFileName != "" {
		f, err := os.OpenFile(newLogFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			panic("create log file failed")
		}
		w := diode.NewWriter(f, 10000, 10*time.Millisecond, func(missed int) {
			logger.Warn().Msgf("Logger Dropped %d messages", missed)
		})
		*logger = logger.Output(w)
		newLogFileName = ""

	} else if LogFilePath != ""{
		f, err := os.OpenFile(LogFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			panic("create log file failed")
		}
		w := diode.NewWriter(f, 10000, 10*time.Millisecond, func(missed int) {
			logger.Warn().Msgf("Logger Dropped %d messages", missed)
		})
		*logger = logger.Output(w)
	}
}


// Monitor is a method for monitoring log files.
func Monitor(logger *Logger) {
	logger.Info().Msg("Monitor log file")
	// Create a monitoring object.
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer watch.Close()

	// Add objects, files to be monitored.
	err = watch.Add(LogFilePath)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	go func() {
		ticker := time.NewTicker(MonitoringFrequency)
		for {
			select {
			case <- ticker.C:
				logger.Info().Msg("check file size")
				if fileSize(LogFilePath) > LogFileSize {
					watch.Events <- fsnotify.Event{
						Name:LogFilePath,
						Op: 4,
					}
				}

			case ev := <-watch.Events:
				{
					// Create a file
					if ev.Op&fsnotify.Create == fsnotify.Create {
						logger.Info().Msgf("Create a file: %s", ev.Name)
						// Redirect the output destination of the log file.
						//output()
					}

					// Write to file
					if ev.Op&fsnotify.Write == fsnotify.Write {
						logger.Info().Msgf("Write to file: %s", ev.Name)
					}

					// Delete file
					if ev.Op&fsnotify.Remove == fsnotify.Remove {
						logger.Info().Msgf("Delete file: %s", ev.Name)
					}

					// Rename file
					if ev.Op&fsnotify.Rename == fsnotify.Rename {
						now := time.Now()
						newLogFileName = fmt.Sprintf("%s-%d-%d-%d-%d-%d",
							LogFilePath, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
						output(newLogFileName, logger)
						logger.Info().Msgf("Rename file: %s", ev.Name)
						rename2File()
						watch.Events <- fsnotify.Event {
							Name:LogFilePath,
							Op: 1,
						}
					}

					// Modify file permissions
					if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
						logger.Info().Msgf("Modify file permissions: %s", ev.Name)
					}
				}
			case err := <-watch.Errors:
				{
					logger.Err(err).Send()
					return
				}
			}
		}
	}()
}





func fileSize(file string) int64 {
	f, err := os.Stat(file)
	if err != nil {
		return 0
	}
	return f.Size()
}

func rename2File() {
	_ = os.Rename(LogFilePath, newLogFileName)
}