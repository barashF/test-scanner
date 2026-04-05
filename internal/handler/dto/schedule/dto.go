package schedule

type CreateRequest struct {
	DaysOfWeek []int  `json:"days_of_week"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
}

type CreateResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"текст ошибки или причина отказа"`
}
