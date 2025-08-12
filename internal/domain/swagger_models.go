package domain

type ErrorResponse struct {
	Error   string                 `json:"error" example:"validation_failed" description:"에러 코드"`
	Message string                 `json:"message" example:"Invalid request body" description:"에러 메시지"`
	Details map[string]interface{} `json:"details,omitempty" description:"추가 에러 상세 정보"`
}

type SuccessResponse struct {
	Message string      `json:"message" example:"Operation completed successfully" description:"성공 메시지"`
	Data    interface{} `json:"data,omitempty" description:"응답 데이터"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok" description:"서버 상태"`
}