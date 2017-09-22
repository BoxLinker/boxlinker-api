package application

func Tables() []interface{} {
	var tables []interface{}
	tables = append(tables, new(HardwareConfigure))
	return tables
}


