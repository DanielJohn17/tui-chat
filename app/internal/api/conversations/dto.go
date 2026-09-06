package conversations

type GetOrCreateDirectConvType struct {
	UserIDOne int64 `json:"user_id_one" validate:"required,gte=1"`
	UserIDTwo int64 `json:"user_id_two" validate:"required,gte=1"`
}

type GetConvParticipantType struct {
	ConvID   int64  `json:"conv_id"`
	UserID   int64  `json:"user_id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}
