package handlers

import (
	"kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
)

func (pph PageHandler) GetProfilePage(ctx *gin.Context) {
	user_id_uint64 := ctx.GetUint64("user_id")
	user, err := data.GetUserById(pph.Database.Pool, user_id_uint64)
	if err != nil {
		ctx.String(500, "Unable to get user from database")
		return
	}
	roles, err := data.GetAllRoles(pph.Database.Pool)
	if err != nil {
		ctx.String(500, "Unable to get roles from database")
		return
	}
	ctx.HTML(200, "pages/profile", gin.H{
		"user":  user,
		"roles": roles,
	})
}
