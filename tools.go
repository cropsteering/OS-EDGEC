package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
)

/**
* Cache string array to JSON
* Write to file
*
 */
func Cache_Map(data interface{}, file_path string) error {
	json_data, err := json.MarshalIndent(data, "", "  ")
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
* Append string array to JSON
* Write to file
* TODO: Check to see if logic exists, duplicates
*
 */
func Append_Map(data map[string]interface{}, file_path string) error {
	jsonData, err := os.ReadFile(file_path)
	if err != nil {
		c_err := Cache_Map(data, file_path)
		if c_err != nil {
			return c_err
		}
		return err
	} else {
		var data_map map[string]interface{}
		jerr := json.Unmarshal(jsonData, &data_map)
		if jerr != nil {
			return jerr
		} else {
			AppendMaps(data_map, data)
			json_data, jerr2 := json.MarshalIndent(data_map, "", "  ")
			if jerr2 != nil {
				return jerr2
			} else {
				ferr := os.WriteFile(file_path, json_data, 0644)
				if ferr != nil {
					return ferr
				}
			}
		}
	}
	return nil
}

func AppendMaps(map1 map[string]interface{}, map2 map[string]interface{}) {
	for key, value := range map2 {
		map1[key] = value
	}
}

/**
* Read JSON from file
*
 */
func Read_Interface(file_path string) (interface{}, error) {
	json_data, err := os.ReadFile(file_path)
	if err != nil {
		return nil, err
	}

	var cached_array interface{}
	err = json.Unmarshal(json_data, &cached_array)
	if err != nil {
		return nil, err
	}

	return cached_array, nil
}

func Read_Map(file_path string) (map[string]interface{}, error) {
	json_data, err := os.ReadFile(file_path)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(json_data, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

/**
* Count how many arrays in our JSON
*
 */
func Array_Count(json_data []byte) (int, error) {
	var data map[string]interface{}

	if err := json.Unmarshal([]byte(json_data), &data); err != nil {
		return 0, err
	}

	count := 0

	for _, v := range data {
		if arr, ok := v.([]interface{}); ok {
			_ = arr
			count++
		}
	}

	return count, nil
}

/**
* Cache string array to JSON
* Write to file
*
 */
func Cache_Array(file_path string, new_data []string) error {
	existing_data, err := Read_Array(file_path)
	if err != nil {
		return err
	}

	updated_data := append(existing_data, new_data...)

	json_data, err := json.MarshalIndent(updated_data, "", "  ")
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
* Cache string array to JSON
* Write to file
*
 */
func Append_String(file_path string, new_data string) error {
	existing_data, err := Read_Array(file_path)
	if err != nil {
		return err
	}

	updated_data := append(existing_data, new_data)

	json_data, err := json.MarshalIndent(updated_data, "", "  ")
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
	file_content, err := os.ReadFile(file_path)

	if os.IsNotExist(err) {
		return []string{}, nil
	}

	if err != nil {
		return nil, err
	}

	var existing_data []string
	err = json.Unmarshal(file_content, &existing_data)
	if err != nil {
		return nil, err
	}

	return existing_data, nil
}

func Iterate_Map(input interface{}) (keys []string, values []interface{}) {
	switch data := input.(type) {
	case map[string]interface{}:
		for key := range data {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for _, key := range keys {
			values = append(values, data[key])
		}
	default:
		log.Println("Input is not a map[string]interface{}")
	}

	return keys, values
}

func Iterate_Interface(data interface{}) []string {
	var temp []string
	switch slice := data.(type) {
	case []interface{}:
		for _, value := range slice {
			temp_string := fmt.Sprintf("%v", value)
			temp = append(temp, temp_string)
		}
	default:
		log.Println("Input is not a []interface{}")
	}
	return temp
}
