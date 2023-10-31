package main

import (
	"io/ioutil"
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

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	// Read in the environmental variable for INTERVAL
	interval, err := time.ParseDuration(SetDefaultString("5s", os.Getenv("INTERVAL")))
	if err != nil {
		logger.Fatal(err.Error())
	}
	// Read in the environmental variable for MESSAGE_PATH
	messagePath := SetDefaultString("./messages", os.Getenv("MESSAGE_PATH"))

	// Find all the files in the MESSAGE_PATH directory
	// that match a .msg extension
	files, err := ioutil.ReadDir(messagePath)
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
				// Remove the .msg extension from the file name
				fileBaseName := strings.TrimSuffix(file.Name(), ".msg")

				// Split the filename at the hyphen
				fileNameParts := strings.Split(fileBaseName, "-")

				// Set the log info variables
				//logType := fileNameParts[0]
				logLevel := fileNameParts[1]

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

				//fmt.Println(file.Name())
			}
		}

		// Sleep for INTERVAL seconds
		time.Sleep(time.Duration(interval))
	}

}
