package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

var Disco_HTML string = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Sensor Discovery</title>
	</head>
		<body>
			<h1>Sensor discovery mode</h1>
			<h1>Status: {{.}}</h1>
			<form method='POST'>
				<label for='Disco'> </label>
				<select name='Disco' id='Disco'>
					<option value='Enable'>Enable</option>
					<option value='Disable'>Disable</option>
				</select>
				<input type='submit' Value='Submit'>
			</form>
			<br />
			<a href="index.html"><button>Back to dashboard</button></a>
		</body>
	</html>
`

var Logic_HTML_1 string = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Logic Editor</title>
	<style>
		.hidden-option {
			display: none;
		}
		table {
            border: 1px solid black;
            border-collapse: collapse;
            width: 25%;
        }
		td, th {
            border: 1px solid black;
            padding: 8px;
            text-align: left;
        }
	</style>
</head>
<body>
	<h1>Logic Editor</h1>
	<form method='POST'>
		<label>Power Controller:</label><br />
		<input type='text' name='POWERC'><br /><br />
		<label for='SENSOR'>IF SENSOR </label>
			<select name='SENSOR' id='SENSOR' onchange='update()'>
`

var Logic_HTML_2 string = `
		<label for='STATUS'> </label>
		<select name='STATUS' id='STATUS'>
			<option value='on'>ON</option>
			<option value='off'>OFF</option>
		</select>
		<input type="checkbox" id="THEN" name="THEN" value="TRUE">
		<label for="checkbox">THEN</label>
`

var Logic_HTML_END = `
	<br /><input type='submit' name='submit' value='Add'><br />
	</form>
	<br /><a href='index.html'><button>Back to dashboard</button></a>
	<script>
	function update() {
		var reading_drop = document.getElementById("READING");
		var sensor_drop = document.getElementById("SENSOR");
		var selected = sensor_drop.value;
		for (var i = 0; i < reading_drop.options.length; i++) {
			var option = reading_drop.options[i];
			if (option.getAttribute('data-group') == selected) {
				option.style.display = 'block';
			} else {
				option.style.display = 'none';
			}
		}
	}
	window.onload = update;

	function validateInput(input) {
		let numericValue = input.value.replace(/[^0-9]/g, '');
	
		numericValue = Math.min(15000, Math.max(0, numericValue));
	
		input.value = numericValue;
	
		const placeholderText = "Enter numbers (0-15000)";
		input.placeholder = input.value.length === 0 ? placeholderText : "";
	  }
</script>
</body>
</html>
`

func Build_Logic() string {
	var s_builder strings.Builder
	s_builder.WriteString(Logic_HTML_1)

	values_array, err := Read_Interface("values.json")
	if err != nil {
		R_LOG(err.Error())
	} else {
		keys, _ := Iterate_Map(values_array)
		for _, key := range keys {
			s_builder.WriteString("				<option value='" + key + "'>" + key + "</option>\n")
		}
		s_builder.WriteString(`
				</select>
				<label for='READING'>READING </label>
					<select name='READING' id='READING'>
		`)
		v_keys, v_values := Iterate_Map(values_array)
		for v_index, v_name := range v_keys {
			temp_value := Iterate_Interface(v_values[v_index])
			for _, v_value := range temp_value {
				s_builder.WriteString("				<option value='" + v_value + "' class='hidden-option' data-group='" + v_name + "'>" + v_value + "</option>\n")
			}
		}
		s_builder.WriteString(`
				</select>
				<label for='EQU'>IS </label>
					<select name='EQU' id='EQU'>
						<option value='less'>Less than</option>
						<option value='lessequ'>Less than equal</option>
						<option value='great'>Greater than</option>
						<option value='greatequ'>Greater than equal</option>
						<option value='equa'>Equal to</option>
						<option value='nequa'>Not equal to</option>
					</select>
		`)
		s_builder.WriteString("				<input type='text' id='VALUE' name='VALUE' placeholder='Enter numbers (0-15000)' oninput='validateInput(this)'>\n")
		s_builder.WriteString(`
				<label for='PIN'>TURN PIN </label>
					<select name='PIN' id='PIN'>
		`)
		for x := 0; x <= 50; x++ {
			str := strconv.Itoa(x)
			s_builder.WriteString("				<option value='" + str + "'>" + str + "</option>\n")
		}
		s_builder.WriteString("			</select>")
	}

	s_builder.WriteString(Logic_HTML_2)

	s_builder.WriteString(`
			<label for='WEIGHT'>| WEIGHT </label>
				<select name='WEIGHT' id='WEIGHT'>
	`)
	for x := 0; x <= 100; x++ {
		str := strconv.Itoa(x)
		s_builder.WriteString("				<option value='" + str + "'>" + str + "</option>\n")
	}
	s_builder.WriteString("			</select>")

	s_builder.WriteString("	<br /><br />\n")

	json_data, err := os.ReadFile("logic.json")
	if err != nil {
		R_LOG("Error: " + err.Error())
	} else {
		s_builder.WriteString("	<table>\n")
		s_builder.WriteString(`
			<thead>
				<tr>
					<th>Sensor</th>
					<th>Sensor reading</th>
					<th>Compare</th>
					<th>Sensor value</th>
					<th>Pin</th>
					<th>Pin state</th>
					<th>Power Controller</th>
					<th>Then</th>
					<th>Weight</th>
					<th>Delete logic</th>
				</tr>
			</thead>
		`)
		s_builder.WriteString("	<tbody>\n")

		json_map := orderedmap.New[string, []string]()

		if err := json.Unmarshal(json_data, &json_map); err != nil {
			R_LOG("Error: " + err.Error())
		} else {
			for el := json_map.Oldest(); el != nil; el = el.Next() {
				s_builder.WriteString("			<tr>\n")
				uuid := el.Key
				logic_slice := el.Value
				for _, l_value := range logic_slice {
					s_builder.WriteString("				<td>" + l_value + "</td>\n")
				}
				s_builder.WriteString("				<td><form method='POST'><input type='hidden' name='uuid' value='" + uuid + "'><input type='submit' name='submit' value='Delete'></form></td>\n")
				s_builder.WriteString("			</tr>\n")
			}
			s_builder.WriteString("		</tbody>\n")
			s_builder.WriteString("	</table>\n")
		}
	}

	s_builder.WriteString(Logic_HTML_END)

	return s_builder.String()
}
