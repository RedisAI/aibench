package inference

type TestResult struct {

	// Test Configs
	ResultFormatVersion  string `json:"ResultFormatVersion"`
	Limit                uint64 `json:"Limit"`
	MetadataAutobatching int64  `json:"MetadataAutobatching"`
	TensorBatchSize      uint64 `json:"TensorBatchSize"`
	Workers              uint   `json:"Workers"`
	MaxRps               uint64 `json:"MaxRps"`

	// Test Description
	TestDescription  string `json:"TestDescription"`

	// DB Spefic Configs
	DBSpecificConfigs map[string]interface{} `json:"DBSpecificConfigs"`

	StartTime      int64 `json:"StartTime`
	EndTime        int64 `json:"EndTime"`
	DurationMillis int64 `json:"DurationMillis"`

	// Totals
	Totals map[string]interface{} `json:"Totals"`

	// Overall Rates
	OverallRates map[string]interface{} `json:"OverallRates"`

	// Overall Quantiles
	OverallQuantiles map[string]interface{} `json:"OverallQuantiles"`

	RunTimeStats map[int64]interface{} `json:"RunTimeStats"`

}
