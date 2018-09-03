package harbor

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/caicloud/nirvana/log"
)

const (
	// Harbor has constraint on page size, the maximum value is 500
	MaxPageSize     = 500
	RespHeaderTotal = "X-Total-Count"
)

func getTotalFromResp(resp *http.Response) (int, error) {
	totalStr := resp.Header.Get(RespHeaderTotal)
	if totalStr == "" {
		return 0, fmt.Errorf("response header %s is empty", RespHeaderTotal)
	}

	total, err := strconv.Atoi(totalStr)
	if err != nil {
		log.Errorf("strconv.Atoi error: %v, resp header %s is %s", RespHeaderTotal, totalStr)
		return 0, err
	}

	return total, nil
}
