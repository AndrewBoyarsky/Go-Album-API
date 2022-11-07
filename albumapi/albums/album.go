package albums

type Album struct {
	ID       string  `json:"id" binding:"required,min=1,max=12"`
	Title    string  `json:"title" binding:"required,min=4,max=30"`
	Artist   string  `json:"artist" binding:"required,min=5,max=40"`
	Price    float64 `json:"price" binding:"required,min=2"`
	UserName string  `json:"userName" bson:"userName"`
	Status   string  `json:"status" bson:"status"`
}
