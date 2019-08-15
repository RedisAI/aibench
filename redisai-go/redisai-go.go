package redisai_go

func Generate_AI_TensorSet_Args(tensorName string, datatype string, dimensions []int, datafmt string, values []float32) []interface{} {

	args := make([]interface{}, (4 + len(dimensions) + len(values)))
	args[0] = "AI.TENSORSET"
	args[1] = tensorName
	args[2] = datatype
	for i := range dimensions {
		args[3+i] = dimensions[i]
	}
	padding := 3 + len(dimensions)
	args[padding] = datafmt
	padding = padding + 1
	for i := range values {
		args[padding+i] = values[i]
	}
	return args
}


func Generate_AI_TensorSetBLOB_Args(tensorName string, datatype string, dimensions []int, datafmt string, blob []byte) []interface{} {

	args := make([]interface{}, (4 + len(dimensions) + 1))
	args[0] = "AI.TENSORSET"
	args[1] = tensorName
	args[2] = datatype
	for i := range dimensions {
		args[3+i] = dimensions[i]
	}
	padding := 3 + len(dimensions)
	args[padding] = datafmt
	padding = padding + 1
	args[padding] = blob
	return args
}

func Generate_AI_TensorGet_Args(tensorName string, datafmt string) []interface{} {
	tensorArgs := make([]interface{}, 3)
	tensorArgs[0] = "AI.TENSORGET"
	tensorArgs[1] = tensorName
	tensorArgs[2] = datafmt
	return tensorArgs
}

func Generate_AI_ModelRun_Args(modelName string, inputs []string, outputs []string) []interface{} {
	args := make([]interface{}, (4 + len(inputs) + len(outputs)))
	args[0] = "AI.MODELRUN"
	args[1] = modelName
	args[2] = "INPUTS"
	for i := range inputs {
		args[3+i] = inputs[i]
	}
	padding := 3 + len(inputs)
	args[padding] = "OUTPUTS"
	padding = padding + 1
	for i := range outputs {
		args[padding+i] = outputs[i]
	}
	return args
}
