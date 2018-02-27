package sheets

func GetTimeData(project string, sheetID string, startTime string, endTime string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"Project": "iReflect",
			"TaskID":  "IR-15",
			"Hours":   0.6,
		},
		{
			"Project": "iReflect",
			"TaskID":  "IR-16",
			"Hours":   0.7,
		},
		{
			"Project": "iReflect",
			"TaskID":  "IR-17",
			"Hours":   7.6,
		},
		{
			"Project": "iReflect",
			"TaskID":  "IR-18",
			"Hours":   0.8,
		},
		{
			"Project": "iReflect",
			"TaskID":  "IR-19",
			"Hours":   0.9,
		},
		{
			"Project": "iReflect",
			"TaskID":  "IR-20",
			"Hours":   4.6,
		},
		{
			"Project": "iReflect",
			"TaskID":  "IR-21",
			"Hours":   2.2,
		},
		{
			"Project": "iReflect",
			"TaskID":  "IR-22",
			"Hours":   1.2,
		},
	}
}
