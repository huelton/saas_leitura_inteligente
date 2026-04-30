package service

// SplitTextIntoPages splits rune-wise into chunks of pageSize characters (logical pages).
func SplitTextIntoPages(text string, pageSize int) []string {
	if pageSize <= 0 {
		pageSize = 1000
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return []string{}
	}
	var pages []string
	for i := 0; i < len(runes); i += pageSize {
		end := i + pageSize
		if end > len(runes) {
			end = len(runes)
		}
		pages = append(pages, string(runes[i:end]))
	}
	return pages
}
