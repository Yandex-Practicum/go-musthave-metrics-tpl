package server

//func TestAdd(t *testing.T) {
//	type fields struct {
//		repository MetricRepository
//	}
//	type args struct {
//		ctx context.Context
//		m   model.Metrics
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &MetricService{
//				repository: tt.fields.repository,
//			}
//			if err := s.Add(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
//				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func TestNewMetricService(t *testing.T) {
//	type args struct {
//		mr MetricRepository
//	}
//	tests := []struct {
//		name string
//		args args
//		want *MetricService
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := NewMetricService(tt.args.mr); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("NewMetricService() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
