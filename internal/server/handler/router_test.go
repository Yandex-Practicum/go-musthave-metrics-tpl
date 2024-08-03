package handler

import (
	"html/template"

	"github.com/vova4o/yandexadv/internal/models"
)

// mockService представляет собой мок-реализацию интерфейса Servicer
type mockService struct {
	updateFunc      func(metric models.Metric) error
	MocGetValueServ func(metric models.Metric) (string, error)
	WebPageFunc     func() (*template.Template, map[string]interface{}, error)
}

func (m *mockService) UpdateServ(metric models.Metric) error {
	return m.updateFunc(metric)
}

func (m *mockService) GetValueServ(metric models.Metric) (string, error) {
	return m.MocGetValueServ(metric)
}

func (m *mockService) MetrixStatistic() (*template.Template, map[string]interface{}, error) {
	return m.WebPageFunc()
}
