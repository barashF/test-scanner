package schedule

type CreateRequest struct {
	DaysOfWeek []int  `json:"days_of_week"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
}
