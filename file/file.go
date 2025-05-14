package file

import (
	"encoding/json"
	"fmt"
	"os"
)

type User struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Sex     int    `json:"sex"`
}

func DataReader(filename string) (output []User, err error) {
	f, err := os.ReadFile(filename)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	err = json.Unmarshal(f, &output)

	if err != nil {
		fmt.Println("Could not unpack the json data")
		return nil, err
	}

	return output, nil
}
