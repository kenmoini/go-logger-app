package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SetDefaultString will return either the default string or an overriden value
func SetDefaultString(defaultVal string, overrideVal string) string {
	if len(strings.TrimSpace(overrideVal)) > 0 {
		return overrideVal
	}
	return defaultVal
}

func isDebug() bool {
	inputDebug := os.Getenv("DEBUG")
	defaultDebug := false
	if inputDebug == "" {
		return defaultDebug
	} else {
		if inputDebug == "true" {
			return true
		} else {
			return false
		}
	}
}

func main() {
	rawJSON := []byte(`{
		"level": "debug",
		"encoding": "json",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
	  }`)
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger, err := cfg.Build()
	//logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	// Read in the environmental variable for INTERVAL
	interval, err := time.ParseDuration(SetDefaultString("5s", os.Getenv("INTERVAL")))
	if err != nil {
		logger.Fatal(err.Error())
	}
	if isDebug() {
		logger.Info("Interval", zap.String("interval", interval.String()))
	}

	// Read in the environmental variable for MESSAGE_PATH
	messagePath := SetDefaultString("./messages", os.Getenv("MESSAGE_PATH"))
	if isDebug() {
		logger.Info("Message Path", zap.String("messagePath", messagePath))
	}

	// Find all the files in the MESSAGE_PATH directory
	// that match a .msg extension
	files, err := os.ReadDir(messagePath)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Start an interval loop
	// that runs every INTERVAL seconds
	// and reads in the files in the MESSAGE_PATH directory
	// that match a .msg extension
	// and logs the message
	// and then deletes the file
	// and then sleeps for INTERVAL seconds
	// and then repeats
	for true == true {

		// Loop through the files
		for _, file := range files {
			// Make sure this is not a directory
			if !file.IsDir() {
				if isDebug() {
					logger.Info("Processing message file", zap.String("file", file.Name()))
				}

				// Remove the .msg extension from the file name
				fileBaseName := strings.TrimSuffix(file.Name(), ".msg")
				if isDebug() {
					logger.Info("Processing message file", zap.String("file", fileBaseName))
				}

				// Split the filename at the hyphen
				fileNameParts := strings.Split(fileBaseName, "-")
				if isDebug() {
					logger.Info("File name parts", zap.Strings("fileNameParts", fileNameParts))
				}

				if len(fileNameParts) > 1 {
					// Set the log info variables
					//logType := fileNameParts[0]
					logLevel := fileNameParts[1]
					if isDebug() {
						logger.Info("Log level", zap.String("logLevel", logLevel))
					}

					// Read in the file data
					fileData, err := os.ReadFile(messagePath + "/" + file.Name())
					if err != nil {
						log.Fatal(err)
					}

					// Log the message
					switch logLevel {
					case "debug":
						logger.Debug(string(fileData))
					case "info":
						logger.Info(string(fileData))
					case "warn":
						logger.Warn(string(fileData))
					case "error":
					case "err":
						logger.Error(string(fileData))
					}
				}

				//fmt.Println(file.Name())
			}
		}

		// Sleep for INTERVAL seconds
		time.Sleep(time.Duration(interval))
	}

}
