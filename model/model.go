package model

type TotalEnergy struct {
	ID          int    `db:"id"`
	UserID      string `db:"user_id"`
	TotalEnergy int    `db:"total_energy"`
}

type ToCollectEnergy struct {
	ID              int    `db:"id"`
	UserID          string `db:"user_id"`
	Status          string `db:"status"`
	ToCollectEnergy int    `db:"to_collect_energy"`
}
