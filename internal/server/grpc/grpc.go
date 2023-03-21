package grpc

import (
	"context"
	"github.com/iddanilov/metricsAndAlerting/internal/db"
	"github.com/iddanilov/metricsAndAlerting/internal/server"
	"log"
	"strings"

	client "github.com/iddanilov/metricsAndAlerting/internal/models"
	// импортируем пакет со сгенерированными protobuf-файлами
	pb "github.com/iddanilov/metricsAndAlerting/proto"
)

// UsersServer поддерживает все необходимые методы сервера.
type MetricsAndAlertingServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	//pb.UnimplementedUsersServer

	// используем sync.Map для хранения пользователей
	//users sync.Map
	s  *server.Storage
	db *db.DB
}

func NewMetricsAndAlertingServer(storage *server.Storage, db *db.DB) *MetricsAndAlertingServer {
	return &MetricsAndAlertingServer{
		s:  storage,
		db: db,
	}
}

func (s *MetricsAndAlertingServer) SaveMetrics(ctx context.Context, request *pb.MetricRequest) (*pb.MetricResponse, error) {
	log.Println("request: ", request)
	if strings.ToLower(request.MType) == "gauge" {
		err := s.db.UpdateMetric(ctx, client.Metrics{
			ID:    request.ID,
			MType: strings.ToLower(request.MType),
			Value: &request.Value,
		})
		if err != nil {
			log.Println(err)
			return nil, err
		}
		s.s.SaveGaugeMetric(&client.Metrics{
			ID:    request.ID,
			MType: strings.ToLower(request.MType),
			Value: &request.Value,
		})
	} else if strings.ToLower(request.MType) == "counter" {
		err := s.db.UpdateMetric(ctx, client.Metrics{
			ID:    request.ID,
			MType: strings.ToLower(request.MType),
			Delta: &request.Delta,
			Value: nil,
		})
		if err != nil {
			log.Println(err)
			return nil, err
		}
		s.s.SaveCountMetric(client.Metrics{
			ID:    request.ID,
			MType: strings.ToLower(request.MType),
			Delta: &request.Delta,
		})
	}
	return nil, nil
}
