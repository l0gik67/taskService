package transport

import (
	"core/model"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

const userIDFilterName = "user_id"

const boardIDFilterName = "board_id"

func taskFiltersFromRequest(r *http.Request) (model.TaskFilters, error) {
	filters := model.TaskFilters{}
	rawUserID := strings.TrimSpace(r.URL.Query().Get(userIDFilterName))
	if rawUserID == "" {
		return filters, model.ErrUserIDFilterRequired
	}

	userID, err := strconv.ParseInt(rawUserID, 10, 64)
	if err != nil {
		return filters, errors.New(`query parameter "user_id" must be a valid integer`)
	}
	filters.UserID = &userID
	rawBoardID := strings.TrimSpace(r.URL.Query().Get(boardIDFilterName))
	if rawBoardID != "" {
		boardID, err := strconv.ParseInt(rawBoardID, 10, 64)
		if err != nil {
			return filters, errors.New(`query parameter "board_id" must be a valid integer`)
		}
		filters.BoardID = &boardID
	}

	if err := filters.Validate(); err != nil {
		return filters, err
	}

	return filters, nil
}
