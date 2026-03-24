package new_object_builder_service

// CRUD operation types for the AI chat CRUD flow.
// These complement the protobuf-generated types when proto regeneration is not available.

type GetProjectTablesSchemaRequest struct {
	ResourceEnvId string `json:"resource_env_id"`
}

func (x *GetProjectTablesSchemaRequest) GetResourceEnvId() string {
	if x != nil {
		return x.ResourceEnvId
	}
	return ""
}

func (x *GetProjectTablesSchemaRequest) Reset()         {}
func (x *GetProjectTablesSchemaRequest) String() string { return x.ResourceEnvId }
func (x *GetProjectTablesSchemaRequest) ProtoMessage()  {}

type DBColumn struct {
	ColumnName string `json:"column_name" protobuf:"bytes,1,opt,name=column_name"`
	DataType   string `json:"data_type" protobuf:"bytes,2,opt,name=data_type"`
	IsNullable string `json:"is_nullable" protobuf:"bytes,3,opt,name=is_nullable"`
}

func (x *DBColumn) GetColumnName() string {
	if x != nil {
		return x.ColumnName
	}
	return ""
}

func (x *DBColumn) GetDataType() string {
	if x != nil {
		return x.DataType
	}
	return ""
}

func (x *DBColumn) GetIsNullable() string {
	if x != nil {
		return x.IsNullable
	}
	return ""
}

func (x *DBColumn) Reset()         {}
func (x *DBColumn) String() string { return x.ColumnName }
func (x *DBColumn) ProtoMessage()  {}

type DBTableSchema struct {
	TableName string      `json:"table_name" protobuf:"bytes,1,opt,name=table_name"`
	Columns   []*DBColumn `json:"columns" protobuf:"bytes,2,rep,name=columns"`
}

func (x *DBTableSchema) GetTableName() string {
	if x != nil {
		return x.TableName
	}
	return ""
}

func (x *DBTableSchema) GetColumns() []*DBColumn {
	if x != nil {
		return x.Columns
	}
	return nil
}

func (x *DBTableSchema) Reset()         {}
func (x *DBTableSchema) String() string { return x.TableName }
func (x *DBTableSchema) ProtoMessage()  {}

type GetProjectTablesSchemaResponse struct {
	Tables []*DBTableSchema `json:"tables" protobuf:"bytes,1,rep,name=tables"`
}

func (x *GetProjectTablesSchemaResponse) GetTables() []*DBTableSchema {
	if x != nil {
		return x.Tables
	}
	return nil
}

func (x *GetProjectTablesSchemaResponse) Reset()         {}
func (x *GetProjectTablesSchemaResponse) String() string { return "" }
func (x *GetProjectTablesSchemaResponse) ProtoMessage()  {}

type ExecuteCrudOperationRequest struct {
	ResourceEnvId string `json:"resource_env_id" protobuf:"bytes,1,opt,name=resource_env_id"`
	Operation     string `json:"operation" protobuf:"bytes,2,opt,name=operation"`
	Table         string `json:"table" protobuf:"bytes,3,opt,name=table"`
	DataJson      string `json:"data_json" protobuf:"bytes,4,opt,name=data_json"`
	WhereJson     string `json:"where_json" protobuf:"bytes,5,opt,name=where_json"`
}

func (x *ExecuteCrudOperationRequest) GetResourceEnvId() string {
	if x != nil {
		return x.ResourceEnvId
	}
	return ""
}

func (x *ExecuteCrudOperationRequest) GetOperation() string {
	if x != nil {
		return x.Operation
	}
	return ""
}

func (x *ExecuteCrudOperationRequest) GetTable() string {
	if x != nil {
		return x.Table
	}
	return ""
}

func (x *ExecuteCrudOperationRequest) GetDataJson() string {
	if x != nil {
		return x.DataJson
	}
	return ""
}

func (x *ExecuteCrudOperationRequest) GetWhereJson() string {
	if x != nil {
		return x.WhereJson
	}
	return ""
}

func (x *ExecuteCrudOperationRequest) Reset()         {}
func (x *ExecuteCrudOperationRequest) String() string { return x.Operation }
func (x *ExecuteCrudOperationRequest) ProtoMessage()  {}

type ExecuteCrudOperationResponse struct {
	ResultJson   string `json:"result_json" protobuf:"bytes,1,opt,name=result_json"`
	RowsAffected int32  `json:"rows_affected" protobuf:"varint,2,opt,name=rows_affected"`
}

func (x *ExecuteCrudOperationResponse) GetResultJson() string {
	if x != nil {
		return x.ResultJson
	}
	return ""
}

func (x *ExecuteCrudOperationResponse) GetRowsAffected() int32 {
	if x != nil {
		return x.RowsAffected
	}
	return 0
}

func (x *ExecuteCrudOperationResponse) Reset()         {}
func (x *ExecuteCrudOperationResponse) String() string { return "" }
func (x *ExecuteCrudOperationResponse) ProtoMessage()  {}
