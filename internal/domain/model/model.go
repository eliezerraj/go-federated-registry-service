package model

import (
	"time"

	go_core_db_pg 		"github.com/eliezerraj/go-core/v2/database/postgre"
	go_core_otel_trace "github.com/eliezerraj/go-core/v2/otel/trace"
)

type AppServer struct {
	Application 	*Application	 				`json:"application"`
	Server     		*Server     					`json:"server"`
	EnvTrace		*go_core_otel_trace.EnvTrace	`json:"env_trace"`
	DatabaseConfig	*go_core_db_pg.DatabaseConfig  	`json:"database_config"`
	Endpoint 		*[]Endpoint						`json:"endpoints,omitempty"`
}

type MessageRouter struct {
	Message			string `json:"message"`
}

type Application struct {
	Name				string 	`json:"name"`
	Version				string 	`json:"version"`
	Account				string 	`json:"account,omitempty"`
	OsPid				string 	`json:"os_pid"`
	IPAddress			string 	`json:"ip_address"`
	Env					string 	`json:"enviroment,omitempty"`
	LogLevel			string 	`json:"log_level,omitempty"`
	OtelTraces			bool   	`json:"otel_traces"`
	OtelMetrics			bool   	`json:"otel_metrics"`
	OtelLogs			bool   	`json:"otel_logs"`
	StdOutLogGroup 		bool   	`json:"stdout_log_group"`
	LogGroup			string 	`json:"log_group,omitempty"`
}

type Server struct {
	Port 			int `json:"port"`
	ReadTimeout		int `json:"readTimeout"`
	WriteTimeout	int `json:"writeTimeout"`
	IdleTimeout		int `json:"idleTimeout"`
	CtxTimeout		int `json:"ctxTimeout"`
}

type Endpoint struct {
	Name			string `json:"name_service"`
	Url				string `json:"url"`
	XApigwApiId		string `json:"x-apigw-api-id,omitempty"` //just in case to call APIGW private via vpce
	HostName		string `json:"host_name"`
	HttpTimeout		time.Duration `json:"httpTimeout"`
}

type APIError struct {
	StatusCode	int    `json:"statusCode"`
	Msg			string `json:"message"`
	TraceId		string `json:"x-request-id,omitempty"`
}

type Query struct {
	Query		string		`json:"query,omitempty"`
	Threshold	float64		`json:"threshold,omitempty"`
	Limit		int			`json:"limit,omitempty"`
	Vector		[]float64	`json:"vector,omitempty"`
}

type EmbedQuery struct {
	Inputs		string		`json:"inputs,omitempty"`
	Normalize	bool		`json:"normalize,omitempty"`
	Truncate	string		`json:"truncate,omitempty"`
	TruncationDirection	string	`json:"truncation_direction,omitempty"`
}

type EmbedResponse struct {
	ServiceId	string		`json:"service_id,omitempty"`
	ServiceType	string		`json:"service_type,omitempty"`
	Endpoint	string		`json:"endpoint,omitempty"`
	Transport	string		`json:"transport,omitempty"`
}