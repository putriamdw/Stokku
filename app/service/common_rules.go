package service

// CountTotalPages membulatkan ke atas tanpa bilangan pecahan
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
