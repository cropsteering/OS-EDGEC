package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
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
*
 */
func Append_Map(data map[string]interface{}, file_path string) error {
	json_data, err := os.ReadFile(file_path)
	if err != nil {
		if os.IsNotExist(err) {
			return Cache_Map(data, file_path)
		}
		return err
	}

	ejson_map := orderedmap.New[string, interface{}]()

	if err := json.Unmarshal(json_data, &ejson_map); err != nil {
		R_LOG("Error: " + err.Error())
		return err
	}

	for key, value := range data {
		ejson_map.Set(key, value)
	}

	json_data, err = json.MarshalIndent(ejson_map, "", "  ")
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
		R_LOG("Input is not a map[string]interface{}")
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
		R_LOG("Input is not a []interface{}")
	}
	return temp
}

func Is_Float(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

func Is_StringEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

func String_Exists(value string, slice []string) bool {
	for _, element := range slice {
		if element == value {
			return true
		}
	}
	return false
}

func Int_Exists(value int, slice []int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func String_Delete(value string, slice []string) []string {
	index := -1
	for i, v := range slice {
		if v == value {
			index = i
			break
		}
	}

	if index != -1 {
		return append(slice[:index], slice[index+1:]...)
	}

	return slice
}

func Key_Exists_String(key string, m map[string]string) bool {
	_, exists := m[key]
	return exists
}

func Key_Exists_Int(key string, m map[string]int) bool {
	_, exists := m[key]
	return exists
}

func R_LOG(msg string) {
	if DEBUG {
		log.Println(msg)
	}
}
