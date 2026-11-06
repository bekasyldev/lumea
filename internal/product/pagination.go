package product

import database "github.com/bekasyldev/lumea/internal/database"

const (
	DirectionNext string = "next"
	DirectionPrev string = "prev"
)

type Pagination struct {
	Data    []database.Product `json:"data"`
	Cursor  int64              `json:"cursor"`
	HasMore bool               `json:"has_more"`
}
