package model

import "duels-api/pkg/repository"

type OptsReq struct {
	Opts repository.Options `json:"opts" query:"opts"`
}
