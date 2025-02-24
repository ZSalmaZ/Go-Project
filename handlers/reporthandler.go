// reporthandler.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	s "project.com/myproject/stores"
)

type ReportHandler struct {
	ReportStore *s.PostgresReportStore
}

func NewReportHandler(rs *s.PostgresReportStore) *ReportHandler {
	return &ReportHandler{
		ReportStore: rs,
	}
}

// HandleReports handles GET /reports?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
func (rh *Handler) HandleReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format", http.StatusBadRequest)
		return
	}
	if endDate.Before(startDate) {
		http.Error(w, "end_date must be after start_date", http.StatusBadRequest)
		return
	}

	report, err := rh.Store.GetSalesReport(ctx, startDate, endDate)
	if err != nil {
		http.Error(w, "Error generating report", http.StatusInternalServerError)
		log.Println("Error generating sales report:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
