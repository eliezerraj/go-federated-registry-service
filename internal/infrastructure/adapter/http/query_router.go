package http

import (
	"net/http"
	"encoding/json"	

	"github.com/go-federated-registry/internal/domain/model"
	"github.com/go-federated-registry/shared/erro"

	"go.opentelemetry.io/otel/codes"
)

// About add product sale fact
func (h *HttpRouters) ConvertQueryToVector(rw http.ResponseWriter, req *http.Request) error {
	ctx, cancel, span := h.withContext(req, "ConvertQueryToVector")
	defer cancel()
	defer span.End()
	
	// decode payload		
	query := model.Query{}
	defer req.Body.Close()
	
	err := json.NewDecoder(req.Body).Decode(&query)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())		
		return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
	}

	// call service
	res, err := h.workerService.ConvertQueryToVector(ctx, &query)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusCreated, res)
}

func (h *HttpRouters) ExecuteVectorSimilarity(rw http.ResponseWriter, req *http.Request) error {
	ctx, cancel, span := h.withContext(req, "ExecuteVectorSimilarity")
	defer cancel()
	defer span.End()
	
	// decode payload		
	query := model.Query{}
	defer req.Body.Close()
	
	err := json.NewDecoder(req.Body).Decode(&query)
	if err != nil {
		span.RecordError(err) 
		span.SetStatus(codes.Error, err.Error())		
		return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
	}
	// call service
	res, err := h.workerService.ExecuteVectorSimilarity(ctx, &query)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusOK, res)
}
	