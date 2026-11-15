package model

// Envelope seragam untuk semua response API
type WebResponse struct {
	Success		bool				`json:"success"`
	Message		string				`json:"message"`
	Data 		any 				`json:"data,omitempty"`
	Meta 		*Meta 			 	`json:"meta,omitempty"`
	Errors		map[string]string  	`json:"errors,omitempty"`
}

// Meta berisi informasi paginasi untuk response list
type Meta struct {
	Page		int		`json:"page"`
	Limit		int		`json:"limit"`
	Total		int		`json:"total"`
	TotalPages	int		`json:"total_pages"`
}

// ListQuery adalah parameter umum untuk endpoint list
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}
