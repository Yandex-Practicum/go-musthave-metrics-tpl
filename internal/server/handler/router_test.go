package handler

import (
	"html/template"

	"github.com/vova4o/yandexadv/internal/models"
)

// mockService представляет собой мок-реализацию интерфейса Servicer
type mockService struct {
	updateFuncJSON      func(metric *models.Metrics) error
	updateFunc          func(metric models.Metric) error
	MocGetValueServ     func(metric models.Metrics) (string, error)
	WebPageFunc         func() (*template.Template, map[string]models.Metrics, error)
	getValueFuncJSON    func(metric models.Metrics) (*models.Metrics, error)
	MocGetValueServJSON func(metric models.Metrics) (*models.Metrics, error)
}

func (m *mockService) GetValueServJSON(metric models.Metrics) (*models.Metrics, error) {
	return m.MocGetValueServJSON(metric)
}

func (m *mockService) UpdateServJSON(metric *models.Metrics) error {
	if m.updateFuncJSON == nil {
		return m.updateFuncJSON(metric)
	}
	return nil
}

func (m *mockService) UpdateServ(metric models.Metric) error {
	return m.updateFunc(metric)
}

func (m *mockService) GetValueServ(metric models.Metrics) (string, error) {
	return m.MocGetValueServ(metric)
}

func (m *mockService) MetrixStatistic() (*template.Template, map[string]models.Metrics, error) {
	return m.WebPageFunc()
}

func (m *mockService) GetValueFuncJSON(metric models.Metrics) (*models.Metrics, error) {
	return m.getValueFuncJSON(metric)
}
