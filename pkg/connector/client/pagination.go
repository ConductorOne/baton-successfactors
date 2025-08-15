package client

import "strconv"

const maxPageSize = "100"

func calculateNextPageStartIndex(response *Response) string {
	if response.StartIndex+response.ItemsPerPage <= response.TotalResults {
		return strconv.Itoa(response.StartIndex + response.ItemsPerPage)
	}

	return ""
}
