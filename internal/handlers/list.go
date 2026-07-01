package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
	"github.com/rs/zerolog/log"
)

// listPage — HTML-шаблон страницы со списком метрик. html/template
// автоматически экранирует значения, что безопаснее ручной сборки строк.
const listPage = `<!doctype html>
<html>
<head><meta charset="utf-8"><title>Metrics</title></head>
<body>
<ul>
{{range .Gauges}}<li>gauge {{.Name}} = {{.Value}}</li>
{{end}}{{range .Counters}}<li>counter {{.Name}} = {{.Value}}</li>
{{end}}</ul>
</body>
</html>`

// listTemplate компилируется один раз при загрузке пакета и переиспользуется
// в ListHandler на каждый запрос — исключает повторный парсинг шаблона.
var listTemplate = template.Must(template.New("list").Parse(listPage))

// listMetric — одна строка списка. Поле Value заранее сформатировано,
// чтобы шаблон не занимался форматированием чисел.
type listMetric struct {
	Name  string
	Value string
}

// listView — данные для шаблона listPage: отдельно gauge- и counter-метрики.
type listView struct {
	Gauges   []listMetric
	Counters []listMetric
}

// ListHandler обрабатывает GET / — возвращает HTML-страницу со списком всех
// gauge- и counter-метрик, известных хранилищу. HTML собирается через
// html/template с автоэкранированием значений.
func ListHandler(srv service.MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gauges := srv.GetAllGauges(r.Context())
		counters := srv.GetAllCounters(r.Context())

		view := listView{
			Gauges:   make([]listMetric, 0, len(gauges)),
			Counters: make([]listMetric, 0, len(counters)),
		}
		for name, value := range gauges {
			view.Gauges = append(view.Gauges, listMetric{
				Name:  name,
				Value: strconv.FormatFloat(value, 'g', -1, 64),
			})
		}
		for name, value := range counters {
			view.Counters = append(view.Counters, listMetric{
				Name:  name,
				Value: strconv.FormatInt(value, 10),
			})
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := listTemplate.Execute(w, view); err != nil {
			log.Error().Err(err).Msg("ошибка рендеринга списка метрик")
		}
	}
}
