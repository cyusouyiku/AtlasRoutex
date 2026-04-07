package metrics

type PrometheusExporter struct {
recorder *Recorder
}

func NewPrometheusExporter(rec *Recorder) *PrometheusExporter {
return &PrometheusExporter{recorder: rec}
}

func (p *PrometheusExporter) Recorder() *Recorder { return p.recorder }
