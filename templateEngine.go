package zactem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func RenderTemplate(template string, data interface{}) (string, error) {
	jsonData, _ := json.Marshal(data)
	var parsedData interface{}
	json.Unmarshal(jsonData, &parsedData)
	data = parsedData
	template = RemoveComments(template)
	imports := findImportStatements(template)
	for _, importStmt := range imports {
		moduleName, importPath := extractImportPath(importStmt)
		importedData, err := importYAML(importPath)
		moduleMap := make(map[string]interface{})
		if len(moduleName) > 0 {
			moduleMap[moduleName] = importedData
		} else {
			moduleMap = importedData
		}
		if err != nil {
			return "", err
		}
		data = mergeData(data, moduleMap)
	}

	if len(imports) == 0 {
		data = mergeData(data, make(map[string]interface{}))
	}

	result, _ := processTemplate(template, data)

	return result, nil
}

func findImportStatements(template string) []string {
	importPattern := regexp.MustCompile(`{{import\s*\*\s*as\s*\w+\s*from\s*"([^"]+)"}}`)
	return importPattern.FindAllString(template, -1)
}

func extractImportPath(importStmt string) (string, string) {
	importPattern := regexp.MustCompile(`{{import\s*\*\s*as\s*(\w+)\s*from\s*"([^"]+)"}}`)
	matches := importPattern.FindStringSubmatch(importStmt)
	if len(matches) >= 3 {
		moduleName := matches[1]
		importPath := matches[2]
		return moduleName, importPath
	}
	return "", ""
}

func isMap(value interface{}) bool {
	_, ok := value.(map[string]interface{})
	return ok
}

func processNestedMap(data map[string]interface{}) map[string]interface{} {
	for key, value := range data {
		if isMap(value) {
			nestedMap := value.(map[string]interface{})
			processNestedMap(nestedMap)
			// newMap := make(map[string]interface{})
			// newMap[key] = nestedMap
			data[key] = nestedMap
		}
	}
	return data
}

func importYAML(importPath string) (map[string]interface{}, error) {
	absPath, err := resolveImportPath(importPath)
	if err != nil {
		return nil, err
	}

	yamlData, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var importedData map[string]interface{}
	err = yaml.Unmarshal(yamlData, &importedData)
	if err != nil {
		return nil, err
	}

	return importedData, nil
}

func resolveImportPath(importPath string) (string, error) {
	baseDir, _ := os.Getwd()
	absPath := filepath.Join(baseDir, importPath)

	return absPath, nil
}

func mergeData(data interface{}, importedData interface{}) map[string]interface{} {
	var dataMap map[string]interface{}
	mergedData := make(map[string]interface{})
	dataMap = convertToMap(data)
	// }
	for key, value := range dataMap {
		mergedData[key] = value
	}

	// Merge the imported data into the mergedData
	if importedDataMap, ok := importedData.(map[string]interface{}); ok {
		for key, value := range importedDataMap {
			mergedData[key] = value
		}
	}

	return mergedData
}

func convertToMap(data interface{}) map[string]interface{} {
	if resultMap, ok := data.(map[string]interface{}); ok {
		return resultMap
	}

	result := make(map[string]interface{})

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Struct {
		typeOfS := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			result[typeOfS.Field(i).Name] = field.Interface()
		}
	}

	return result
}

func processTemplate(template string, data interface{}) (string, error) {
	result := RemoveImportLines(template)
	processData(&result, "", data, "")
	return result, nil
}

func processData(result *string, key string, value interface{}, placeholder string) {
	reflectType := reflect.ValueOf(value)
	switch reflectType.Kind() {
	case reflect.String:
		processString(result, placeholder, value.(string))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		processNumber(result, placeholder, fmt.Sprintf("%v", value.(int)))
	case reflect.Float32:
		processNumber(result, placeholder, fmt.Sprintf("%f", value.(float32)))
	case reflect.Float64:
		processNumber(result, placeholder, fmt.Sprintf("%f", value.(float64)))
	case reflect.Bool:
		processBool(result, placeholder, value.(bool))
	case reflect.Map:
		processMap(result, placeholder, key, value.(map[string]interface{}))
	case reflect.Slice, reflect.Array:
		processSlice(result, placeholder, reflectType)
	}
}

func processMap(result *string, placeholder string, parentKey string, data map[string]interface{}) {
	if regexExists(placeholder, *result) && len(data) > 0 && len(placeholder) > 0 {
		convertedYml, _ := FormatMapAsYAML(data)
		*result = strings.ReplaceAll(*result, placeholder, convertedYml)
	}

	// var str string
	for key, value := range data {
		keyString := fmt.Sprintf("%v%v", func() string {
			if len(parentKey) > 0 {
				return parentKey + "."
			}
			return ""
		}(), key)
		placeholder := fmt.Sprintf("{{%s}}", keyString)
		processData(result, keyString, value, placeholder)
	}
}

func processString(result *string, placeholder string, value string) {
	*result = strings.ReplaceAll(*result, placeholder, value)
}

func processNumber(result *string, placeholder string, value string) {
	*result = strings.ReplaceAll(*result, placeholder, value)
}

func processBool(result *string, placeholder string, value bool) {
	strValue := fmt.Sprintf("%v", value)
	*result = strings.ReplaceAll(*result, placeholder, strValue)
}

func processSlice(result *string, placeholder string, value reflect.Value) {
	for i := 0; i < value.Len(); i++ {
		processData(result, "", value.Index(i), placeholder)
		*result = strings.ReplaceAll(*result, placeholder, *result)
	}
}

func FormatMapAsYAML(data map[string]interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

func regexExists(regex, str string) bool {
	match, err := regexp.MatchString(regex, str)
	if err != nil {
		// Handle the error, if any
		fmt.Println("Error:", err)
	}
	return match
}

func RemoveComments(input string) string {
	commentPattern := regexp.MustCompile(`(?m)^#.*$`)
	result := commentPattern.ReplaceAllString(input, "")
	result = strings.TrimSpace(result)
	return result
}

func RemoveImportLines(input string) string {
	importPattern := regexp.MustCompile(`(?m).*{{import.*}}.*\n?`)
	result := importPattern.ReplaceAllString(input, "")
	result = strings.TrimSpace(result)
	return result
}
