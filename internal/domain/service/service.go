package service

import (
	"fmt"
	"encoding/json"
	"context"

	"github.com/rs/zerolog"

	"github.com/go-federated-registry/shared/erro"
	"github.com/go-federated-registry/internal/domain/model"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
	
	database "github.com/go-federated-registry/internal/infrastructure/repo/database"

	go_core_http "github.com/eliezerraj/go-core/v2/http"
	go_core_db_pg "github.com/eliezerraj/go-core/v2/database/postgre"
	go_core_otel_trace "github.com/eliezerraj/go-core/v2/otel/trace"
	go_core_midleware "github.com/eliezerraj/go-core/v2/middleware"
)

type WorkerService struct {
	workerRepository *database.WorkerRepository
	logger 			*zerolog.Logger
	tracerProvider 	*go_core_otel_trace.TracerProvider
	httpService		*go_core_http.HttpService
	endpoint		*[]model.Endpoint
}

// Helper: Get service endpoint by index with error handling
func (s *WorkerService) getServiceEndpoint(index int) (*model.Endpoint, error) {
	if s.endpoint == nil || len(*s.endpoint) <= index {
		return nil, fmt.Errorf("service endpoint at index %d not found", index)
	}
	return &(*s.endpoint)[index], nil
}

// Helper: Build HTTP headers with request ID
func (s *WorkerService) buildHeaders(ctx context.Context) map[string]string {
	requestID := go_core_midleware.GetRequestID(ctx)
	return map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"x-request-id":  requestID,
	}
}

// about do http call 
func (s *WorkerService) doHttpCall(ctx context.Context,	httpClientParameter go_core_http.HttpClientParameter) (interface{}, error) {
	s.logger.Info().
			 Ctx(ctx).
			 Str("func","doHttpCall").Send()

	resPayload, statusCode, err := s.httpService.DoHttp(ctx, httpClientParameter)
	if err != nil {
		s.logger.Error().
			Ctx(ctx).
			Err(err).Send()
		return nil, err
	}

	s.logger.Debug().
		Interface("+++++++++++++++++> httpClientParameter.Url:",httpClientParameter.Url).
		Interface("+++++++++++++++++> resPayload:",resPayload).
		Interface("+++++++++++++++++> statusCode:",statusCode).
		Interface("+++++++++++++++++> err:", err).
		Send()

	switch (statusCode) {
		case 200:
			return resPayload, nil
		case 201:
			return resPayload, nil	
		case 400:
		case 401:
		case 403:
		case 404:
		case 500:
			newErr := fmt.Errorf("internal server error (status code %d) - (process: %s)", statusCode, httpClientParameter.Url)
			s.logger.Error().
				Ctx(ctx).
				Err(newErr).Send()
			return nil, newErr
		default:
	}

	// marshal response payload
	jsonString, err := json.Marshal(resPayload)
	if err != nil {
		s.logger.Error().
			Ctx(ctx).
			Err(err).Send()
		return nil, fmt.Errorf("FAILED to marshal http response: %w (process: %s)", err, httpClientParameter.Url)
	}

	// parse error message
	message := model.APIError{}
	if err := json.Unmarshal(jsonString, &message); err != nil {
		s.logger.Error().
			Ctx(ctx).
			Err(err).Send()
		return nil, fmt.Errorf("FAILED to unmarshal error response: %w (process: %s)", err, httpClientParameter.Url)
	}

	newErr := fmt.Errorf("%s - (status code %d) - (process: %s)", message.Msg,statusCode, httpClientParameter.Url)
	s.logger.Warn().
		Ctx(ctx).
		Err(newErr).Send()
		
	return nil, newErr
}

// About new worker service
func NewWorkerService(	workerRepository *database.WorkerRepository, 
						appLogger 		*zerolog.Logger,
						tracerProvider 	*go_core_otel_trace.TracerProvider,
						endpoint		*[]model.Endpoint	) *WorkerService{
							
	logger := appLogger.With().
						Str("package", "domain.service").
						Logger()
	logger.Info().
			Str("func","NewWorkerService").Send()

	httpService := go_core_http.NewHttpService(&logger)	

	return &WorkerService{
		workerRepository: workerRepository,
		logger: &logger,
		tracerProvider: tracerProvider,
		httpService: httpService,
		endpoint: endpoint,
	}
}

// Helper function for common repository read operations
func (s *WorkerService) callRepositoryRead(ctx context.Context, spanName string, 
	fn func(context.Context) (interface{}, error)) (interface{}, error) {
	s.logger.Info().
			Ctx(ctx).
			Str("func", spanName).Send()
	
	ctx, span := s.tracerProvider.SpanCtx(ctx, "service."+spanName, trace.SpanKindServer)
	defer span.End()
	
	return fn(ctx)
}

// About database stats
func (s *WorkerService) Stat(ctx context.Context) (go_core_db_pg.PoolStats){
	s.logger.Info().
			Ctx(ctx).
			Str("func","Stat").Send()

	return s.workerRepository.Stat(ctx)
}

// About check health service
func (s * WorkerService) HealthCheck(ctx context.Context) error{
	s.logger.Info().
			Ctx(ctx).
			Str("func","HealthCheck").Send()

	ctx, span := s.tracerProvider.SpanCtx(ctx, "service.HealthCheck", trace.SpanKindServer)
	defer span.End()

	// Check database health
	ctx, spanDB := s.tracerProvider.SpanCtx(ctx, "DatabasePG.Ping", trace.SpanKindClient)
	err := s.workerRepository.DatabasePG.Ping()
	spanDB.End()
	
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())

		s.logger.Error().
				Ctx(ctx).
				Err(err).Msg("*** Database HEALTH CHECK FAILED ***")
		return erro.ErrHealthCheck
	}

	s.logger.Info().
			Ctx(ctx).
			Msg("*** Database HEALTH CHECK SUCCESSFULL ***")

	// check service/dependencies 
	endpoint, err := s.getServiceEndpoint(0)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		s.logger.Error().
			Ctx(ctx).
			Err(err).Send()
		return erro.ErrHealthCheck
	}

	headers := s.buildHeaders(ctx)
	
	ctxService00, spanService00 := s.tracerProvider.SpanCtx(ctx, "health.service." + endpoint.HostName, trace.SpanKindServer)
	httpClientParameter := go_core_http.HttpClientParameter {
		Url:	fmt.Sprintf("%s%s", endpoint.Url, "/info"),
		Method:	"GET",
		Timeout: endpoint.HttpTimeout,
		Headers: &headers,
	}

	// call a service via http
	_, err = s.doHttpCall(ctxService00, 
						   httpClientParameter)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())		
		s.logger.Error().
			Ctx(ctxService00).
			Err(err).Msgf("*** Service %s HEALTH CHECK FAILED ***", endpoint.HostName)
		return erro.ErrHealthCheck
	}
	spanService00.End()

	s.logger.Info().
		Ctx(ctx).
		Msgf("*** Service %s HEALTH CHECK SUCCESSFULL ***", endpoint.HostName)

	return nil
}
