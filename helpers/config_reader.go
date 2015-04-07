package helpers

import (
	"encoding/json"
	"errors"
	"os"
)

// DecodeFile is a more generic JSON parser
func DecodeFile(fileName string, object interface{}) error {

	//Open the config file
	file, err := os.Open(fileName)

	if err != nil {
		return errors.New("Could not open file: " + err.Error())
	}

	jsonParser := json.NewDecoder(file)
	err = jsonParser.Decode(object)

	if err != nil {
		return errors.New("Could not parse file: " + err.Error())
	} else {
		return nil
	}

}
