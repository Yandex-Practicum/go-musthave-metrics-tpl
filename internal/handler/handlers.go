package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"yandex-go-musthave-metrics-tpl/internal/model"
	"yandex-go-musthave-metrics-tpl/internal/service/server"
)

var _ MetricService = (*server.MetricService)(nil)

type MetricService interface {
	Add(ctx context.Context, m model.Metrics) error
	GetValue(ctx context.Context, name string) (*float64, error)
	GetAll(ctx context.Context) ([]model.Metrics, error)
}

type MetricHandler struct {
	MetricService MetricService
}

func NewMetricHandler(ms MetricService) *MetricHandler {
	return &MetricHandler{
		MetricService: ms,
	}
}

func (h *MetricHandler) checkType(mType string, w http.ResponseWriter) bool {
	if mType != model.Gauge && mType != model.Counter {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return false
	}
	return true
}

func (h *MetricHandler) checkValue(mValue string, w http.ResponseWriter) (float64, error) {
	valFloat, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return 0, err
	}
	return valFloat, nil
}

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
	mType := r.PathValue("type")
	name := r.PathValue("name")
	value := r.PathValue("value")
	if !h.checkType(mType, w) {
		return
	}

	floatValue, err := h.checkValue(value, w)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.MetricService.Add(r.Context(), model.Metrics{
		ID:    name,
		MType: mType,
		Delta: nil,
		Value: &floatValue,
		Hash:  "",
	})
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) Get(w http.ResponseWriter, r *http.Request) {
	mType := r.PathValue("type")
	name := r.PathValue("name")
	if !h.checkType(mType, w) {
		return
	}
	metricValue, err := h.MetricService.GetValue(r.Context(), name)
	w.Header().Set("Content-Type", "text/plain")
	if err != nil || metricValue == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatFloat(*metricValue, 'f', -1, 64)))
}

func (h *MetricHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.MetricService.GetAll(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	resp, err := json.Marshal(metrics)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}
