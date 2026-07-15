package database

import (
	"context"
	"fmt"

	"github.com/go-federated-registry/internal/domain/model"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

func (w *WorkerRepository) ExecuteVectorSimilarity(ctx context.Context,
													query *model.Query) (*[]model.EmbedResponse, error){
	w.logger.Info().
			Ctx(ctx).
			Str("func","ExecuteVectorSimilarity").Send()

	// Trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.ExecuteVectorSimilarity", trace.SpanKindInternal)
	defer span.End()

	// db connection
	conn, err := w.DatabasePG.Acquire(ctx)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		w.logger.Error().
			  	 Ctx(ctx).
				 Err(err).Send()
		return nil, fmt.Errorf("FAILED to acquire connection: %w", err)
	}
	defer w.DatabasePG.Release(conn)

	// Convert vector to string for SQL query
	strVector := "["
	for i, v := range query.Vector {
		strVector += fmt.Sprintf("%f", v)
		if i < len(query.Vector)-1 {
			strVector += ","
		}
	}
	strVector += "]"

	sqlQuery := `select sr.service_id,
						sr.base_transport,
						se.url
				from 	service_registry sr,
						service_endpoints se,
						service_vectors sv
				where 	se.fk_service_registry_id = sr.id
				and 	sv.fk_service_registry_id = sr.id
				and (sv.search_vector <=> $1) < $2
				order by (sv.search_vector <=> $1) asc
				limit $3;`

	rows, err := conn.Query(ctx,
							sqlQuery,
							strVector,
							query.Threshold,
							query.Limit,
						)
	if err != nil {
		span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, fmt.Errorf("FAILED to query service_vectors: %w", err)
	}
	defer rows.Close()

    if err := rows.Err(); err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		w.logger.Error().
				Ctx(ctx).
				Err(err).Msg("error iterating cartitem rows")
        return nil, fmt.Errorf("error iterating cartitem rows: %w", err)
    }

	embedResponses := []model.EmbedResponse{}
	embedResponse := model.EmbedResponse{}
	
	for rows.Next() {
		err = rows.Scan(&embedResponse.ServiceId,
						&embedResponse.Transport,
						&embedResponse.Endpoint,
		)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			w.logger.Error().
					Ctx(ctx).
					Err(err).Send()
			return nil, fmt.Errorf("FAILED to iterate over service_vectors: %w", err)
		}
		embedResponses = append(embedResponses, embedResponse)
	}

	return &embedResponses, nil
}