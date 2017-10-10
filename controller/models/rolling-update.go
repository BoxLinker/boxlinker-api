package models

func RollingUpdateTables() []interface{} {
	var tables []interface{}
	tables = append(tables, new(RollingUpdate))
	return tables
}

type RollingUpdate struct {

}