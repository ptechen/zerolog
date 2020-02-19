package zerolog

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"time"
)

// output Example
//func output() {
//	if LogFilePath != ""{
//		f, err := os.OpenFile(LogFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
//		if err != nil {
//			panic("create log file failed")
//		}
//		w := diode.NewWriter(f, 10000, 10*time.Millisecond, func(missed int) {
//			GlobalsLogger.Warn().Msgf("Logger Dropped %d messages", missed)
//		})
//		*GlobalsLogger = (GlobalsLogger.Output(w))
//	}
//}

// Monitor is a method for monitoring log files.
func Monitor(output func()) {
	// Create a monitoring object.
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		GlobalsLogger.Fatal().Err(err).Send()
	}
	defer watch.Close()

	// Add objects, files to be monitored.
	err = watch.Add(LogFilePath)
	if err != nil {
		GlobalsLogger.Fatal().Err(err).Send()
	}
	ticker := time.NewTicker(MonitoringFrequency)

	go func() {
		for {
			select {
			case <- ticker.C:
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
						GlobalsLogger.Info().Msgf("Create a file: %s", ev.Name)

						// Redirect the output destination of the log file.
						output()
					}

					// Write to file
					if ev.Op&fsnotify.Write == fsnotify.Write {
						GlobalsLogger.Info().Msgf("Write to file: %s", ev.Name)
					}

					// Delete file
					if ev.Op&fsnotify.Remove == fsnotify.Remove {
						GlobalsLogger.Info().Msgf("Delete file: %s", ev.Name)
					}

					// Rename file
					if ev.Op&fsnotify.Rename == fsnotify.Rename {
						GlobalsLogger.Info().Msgf("Rename file: %s", ev.Name)
						rename2File()
						watch.Events <- fsnotify.Event {
							Name:LogFilePath,
							Op: 1,
						}
					}

					// Modify file permissions
					if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
						GlobalsLogger.Info().Msgf("Modify file permissions: %s", ev.Name)
					}
				}
			case err := <-watch.Errors:
				{
					log.Println("error : ", err)
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
	now := time.Now()
	_ = os.Rename(LogFilePath, fmt.Sprintf("%s-%d-%d-%d-%d-%d",
		LogFilePath, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()))
}