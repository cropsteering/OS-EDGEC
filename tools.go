package main

import (
	"encoding/json"
	"os"
)

/**
* Cache string array to JSON
* Write to file
*
 */
func Cache_Array(data_array []string, file_path string) error {
	existing_data, err := Read_Array(file_path)
	if err != nil {
		if os.IsNotExist(err) {
			existing_data = data_array
		} else {
			return err
		}
	} else {
		existing_data = append(existing_data, data_array...)
	}

	json_data, err := json.Marshal(existing_data)
	if err != nil {
		return err
	}
	err = os.WriteFile(file_path, json_data, 0644)
	if err != nil {
		return err
	}

	return nil
}

/**
* Read JSON from file
*
 */
func Read_Array(file_path string) ([]string, error) {
	json_data, err := os.ReadFile(file_path)
	if err != nil {
		return nil, err
	}

	var cached_array []string
	err = json.Unmarshal(json_data, &cached_array)
	if err != nil {
		return nil, err
	}

	return cached_array, nil
}
