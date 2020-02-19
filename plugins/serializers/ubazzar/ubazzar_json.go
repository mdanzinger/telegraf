package ubazzar

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"time"

	"github.com/influxdata/telegraf"
)

type serializer struct {
	TimestampUnits time.Duration
}

type event struct {
	EventID string `json:"event_id"`
	ServiceCustomerID string `json:"service_customer_id"`
	Service string `json:"service"`
	UnitOfMeasure string `json:"unit_of_measure"`
	Quantity float64 `json:"quantity"`
	StartTime string `json:"start_time"`
	EndTime string `json:"end_time"`
	MetaData map[string]string `json:"meta_data"`
}

func NewSerializer(timestampUnits time.Duration) (*serializer, error) {
	s := &serializer{
		TimestampUnits: truncateDuration(timestampUnits),
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	e := s.createObject(metric)
	serialized, err := json.Marshal(e)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	objects := make([]interface{}, 0, len(metrics))
	for _, metric := range metrics {
		e := s.createObject(metric)
		objects = append(objects, e)
	}

	obj := map[string]interface{}{
		"metrics": objects,
	}

	serialized, err := json.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}
	return serialized, nil
}

func (s *serializer) createObject(metric telegraf.Metric) *event {
	eventID, _ := uuid.NewV4()
	service, _ := metric.GetTag("service")
	customerID, _ := metric.GetTag("customer_id")
	unitOfMeasure, _ := metric.GetTag("unit_of_measure")
	startTime, ok := metric.GetField("start_time")
	if !ok {
		startTime = time.Now().Add(time.Second*-5).Format(time.RFC3339)
	}
	quantity, _ := metric.GetField("quantity")


	filteredTags := make(map[string]string)
	for k, v := range metric.Tags() {
		switch k {
		case "service", "customer_id", "unit_of_measure":
			continue
		}
		filteredTags[k] = v
	}


	e := &event{
		EventID:           eventID.String(),
		ServiceCustomerID: customerID,
		Service:           service,
		UnitOfMeasure:     unitOfMeasure,
		Quantity:          quantity.(float64),
		StartTime: startTime.(string),
		EndTime:           metric.Time().Format(time.RFC3339),
		MetaData: filteredTags,
	}

	return e
}

func truncateDuration(units time.Duration) time.Duration {
	// Default precision is 1s
	if units <= 0 {
		return time.Second
	}

	// Search for the power of ten less than the duration
	d := time.Nanosecond
	for {
		if d*10 > units {
			return d
		}
		d = d * 10
	}
}
