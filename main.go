package main

import (
	"encoding/json"
	"fmt"
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

type jsonMessage struct {
	Host      string `json:"host"`
	Message   string `json:"message"`
	PID       int    `json:"pid"`
	TID       int    `json:"tid"`
	Timestamp string `json:"timestamp"`
}
type jsonObjectMarshaler struct {
	obj any
}

func (j *jsonObjectMarshaler) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(j.obj)
	// bytes, err := protojson.Marshal(j.obj)
	if err != nil {
		return nil, fmt.Errorf("json marshaling failed: %w", err)
	}
	return bytes, nil
}

func ZapJsonable(key string, obj any) zap.Field {
	return zap.Reflect(key, &jsonObjectMarshaler{obj: obj})
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
		  "levelEncoder": "lowercase",
		  "timeKey": "time",
		  "timeEncoder": "iso8601"
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

	// Setup some sugar
	sugar := logger.Sugar()

	// Read in the environmental variable for INTERVAL
	interval, err := time.ParseDuration(SetDefaultString("5s", os.Getenv("INTERVAL")))
	if err != nil {
		sugar.Fatal(err.Error())
	}
	if isDebug() {
		sugar.Infow("Interval", "interval", interval.String())
	}

	// Read in the environmental variable for MESSAGE_PATH
	messagePath := SetDefaultString("./messages", os.Getenv("MESSAGE_PATH"))
	if isDebug() {
		sugar.Infow("Message Path", "messagePath", messagePath)
	}

	// Find all the files in the MESSAGE_PATH directory
	// that match a .msg extension
	files, err := os.ReadDir(messagePath)
	if err != nil {
		sugar.Fatal(err.Error())
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
					sugar.Infow("Processing message file", "file", file.Name())
				}

				// Remove the .msg extension from the file name
				fileBaseName := strings.TrimSuffix(file.Name(), ".msg")
				if isDebug() {
					sugar.Infow("Processing message file", "file", fileBaseName)
				}

				// Split the filename at the hyphen
				fileNameParts := strings.Split(fileBaseName, "-")
				if isDebug() {
					sugar.Infow("File name parts", "fileNameParts", fileNameParts)
				}

				if len(fileNameParts) > 1 {
					// Set the log info variables
					logType := fileNameParts[0]
					if isDebug() {
						sugar.Infow("Log type", "logType", logType)
					}
					// Split the log type by underscores
					logTypeParts := strings.Split(logType, "_")
					if isDebug() {
						sugar.Infow("Log type parts", "logTypeParts", logTypeParts)
					}

					// Set the log level
					logLevel := fileNameParts[1]
					if isDebug() {
						sugar.Infow("Log level", "logLevel", logLevel)
					}

					// Read in the file data
					fileData, err := os.ReadFile(messagePath + "/" + file.Name())
					if err != nil {
						log.Fatal(err)
					}

					if len(logTypeParts) > 1 {
						switch logTypeParts[0] {
						case "json":
							jsonData := jsonMessage{}
							err = json.Unmarshal(fileData, &jsonData)

							if err != nil {
								sugar.Fatal(err.Error())
							} else {
								switch logLevel {
								case "debug":
									sugar.Debugw(file.Name(), ZapJsonable("event", jsonData))
								case "info":
									sugar.Infow(file.Name(), ZapJsonable("event", jsonData))
								case "warn":
									sugar.Warnw(file.Name(), ZapJsonable("event", jsonData))
								case "error":
								case "err":
									sugar.Errorw(file.Name(), ZapJsonable("event", jsonData))
								}
							}
						case "text":
							switch logLevel {
							case "debug":
								sugar.Debug(string(fileData))
							case "info":
								sugar.Info(string(fileData))
							case "warn":
								sugar.Warn(string(fileData))
							case "error":
							case "err":
								sugar.Error(string(fileData))
							}
						}
					} else {
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
				}

				//fmt.Println(file.Name())
			}
		}

		// Sleep for INTERVAL seconds
		time.Sleep(time.Duration(interval))
	}

}
