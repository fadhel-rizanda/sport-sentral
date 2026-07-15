package mapper_test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	commonv1 "microservice-golang/gen/common/v1"
	"microservice-golang/services/gateway/internal/mapper"
)

func TestFormatTimestamp(t *testing.T) {
	t.Run("nil_timestamp", func(t *testing.T) {
		res := mapper.FormatTimestamp(nil)
		if res != "" {
			t.Errorf("expected empty string, got %s", res)
		}
	})

	t.Run("valid_timestamp", func(t *testing.T) {
		now := time.Date(2026, 7, 14, 15, 0, 0, 0, time.UTC)
		pb := timestamppb.New(now)
		res := mapper.FormatTimestamp(pb)
		expected := "2026-07-14T15:00:00Z"
		if res != expected {
			t.Errorf("expected %s, got %s", expected, res)
		}
	})
}

func TestFormatTimestampPtr(t *testing.T) {
	t.Run("nil_timestamp", func(t *testing.T) {
		res := mapper.FormatTimestampPtr(nil)
		if res != nil {
			t.Errorf("expected nil, got %v", res)
		}
	})

	t.Run("valid_timestamp", func(t *testing.T) {
		now := time.Date(2026, 7, 14, 15, 0, 0, 0, time.UTC)
		pb := timestamppb.New(now)
		res := mapper.FormatTimestampPtr(pb)
		if res == nil {
			t.Fatal("expected pointer, got nil")
		}
		expected := "2026-07-14T15:00:00Z"
		if *res != expected {
			t.Errorf("expected %s, got %s", expected, *res)
		}
	})
}

func TestToCountrySimpleResponse(t *testing.T) {
	t.Run("nil_country", func(t *testing.T) {
		res := mapper.ToCountrySimpleResponse(nil)
		if res != nil {
			t.Errorf("expected nil, got %v", res)
		}
	})

	t.Run("valid_country", func(t *testing.T) {
		c := &commonv1.CountrySimple{
			Id:           "c1",
			Name:         "Indonesia",
			IsoAlpha_2:   "ID",
			IsoAlpha_3:   "IDN",
			PhoneCode:    "62",
			CurrencyCode: "IDR",
		}
		res := mapper.ToCountrySimpleResponse(c)
		if res == nil {
			t.Fatal("expected response, got nil")
		}
		if res.Name != "Indonesia" {
			t.Errorf("expected Indonesia, got %s", res.Name)
		}
		if res.ISOAlpha2 != "ID" {
			t.Errorf("expected ID, got %s", res.ISOAlpha2)
		}
	})
}
