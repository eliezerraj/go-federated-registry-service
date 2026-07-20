package service

import (
	"context"
	"fmt"
	"encoding/json"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
	"github.com/go-federated-registry/internal/domain/model"
	go_core_http "github.com/eliezerraj/go-core/v2/http"
)

// About convert query to vector using the vector service
func (s *WorkerService) ConvertQueryToVector(ctx context.Context, 
											query *model.Query) (*model.Query, error){
	s.logger.Info().
			Ctx(ctx).
			Str("func","ConvertQueryToVector").Send()

	// trace
	ctx, span := s.tracerProvider.SpanCtx(ctx, "service.ConvertQueryToVector", trace.SpanKindServer)
	defer span.End()

	// Get service endpoint
	endpoint, err := s.getServiceEndpoint(0)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())	
		return nil, err
	}
	headers := s.buildHeaders(ctx)

	embedQuery := &model.EmbedQuery{
		Inputs: query.Query,
	}

	httpClientParameter := go_core_http.HttpClientParameter {
		Url:  fmt.Sprintf("%s%s", endpoint.Url, "/embed"),
		Method: "POST",
		Timeout: endpoint.HttpTimeout,
		Headers: &headers,
		Body: embedQuery,
	}

	resPayload, err := s.doHttpCall(ctx, httpClientParameter)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		s.logger.Error().
			Ctx(ctx).
			Err(err).Send()
		return nil, err
	}

	var embeddings [][]float64
	bytePayload, err := json.Marshal(resPayload)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	err = json.Unmarshal(bytePayload, &embeddings)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if len(embeddings) > 0 {
		query.Vector = embeddings[0]
	}

	return query, nil
}

func (s *WorkerService) ExecuteVectorSimilarity(ctx context.Context,
												query *model.Query) (*[]model.EmbedResponse, error){
	s.logger.Info().
			Ctx(ctx).
			Str("func","ExecuteVectorSimilarity").Send()
	// trace
	ctx, span := s.tracerProvider.SpanCtx(ctx, "service.ExecuteVectorSimilarity", trace.SpanKindServer)
	defer span.End()
	
	resQuery, err := s.ConvertQueryToVector(ctx, query)
	if err != nil {
		span.RecordError(err) 
		span.SetStatus(codes.Error, err.Error())
		s.logger.Error().
			Ctx(ctx).
			Err(err).Send()
		return nil, err
	}

	embedResponse, err := s.workerRepository.ExecuteVectorSimilarity(ctx, resQuery)
	if err != nil {
		span.RecordError(err) 
		span.SetStatus(codes.Error, err.Error())
		s.logger.Error().
			Ctx(ctx).
			Err(err).Send()
		return nil, err
	}

	return embedResponse, nil
}