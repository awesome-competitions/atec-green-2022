package handler

import (
	"bytes"
	"energy/config"
	"energy/db"
	"energy/log"
	"energy/model"
	"energy/server"
	"energy/util"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TableToCollectEnergy        = "to_collect_energy"
	TableTotalEnergy            = "total_energy"
	NotCollected         uint64 = 0
	AllCollected         uint64 = 1
	CollectedByOther     uint64 = 2
)

type Handler struct {
	sync.Mutex
	DB                     *db.DB
	totalEnergies          []*totalEnergy
	toCollectEnergies      []*toCollectEnergy
	toCollectEnergyChanges map[int]bool
	totalEnergyChanges     map[int]int
	counter                uint64
	startAt                int64
	sqlBuffer              bytes.Buffer
}

type toCollectEnergy struct {
	sync.Mutex

	ID     int
	UserID int
	Status string
	Total  int
}

type totalEnergy struct {
	sync.Mutex

	ID    int
	Total int
}

func New(d *db.DB) *Handler {
	return &Handler{
		DB:                     d,
		totalEnergies:          make([]*totalEnergy, 10_0001),
		toCollectEnergies:      make([]*toCollectEnergy, 100_0001),
		toCollectEnergyChanges: map[int]bool{},
		totalEnergyChanges:     map[int]int{},
		sqlBuffer:              bytes.Buffer{},
	}
}

func (h *Handler) Init() error {
	toCollectEnergies := make([]*model.ToCollectEnergy, 0)
	err := h.DB.Collection(TableToCollectEnergy).Find().All(&toCollectEnergies)
	if err != nil {
		return err
	}
	totalEnergies := make([]*model.TotalEnergy, 0)
	err = h.DB.Collection(TableTotalEnergy).Find().All(&totalEnergies)
	if err != nil {
		return err
	}

	for _, tce := range toCollectEnergies {
		userId, _ := strconv.Atoi(tce.UserID)
		h.toCollectEnergies[tce.ID] = &toCollectEnergy{
			ID:     tce.ID,
			UserID: userId,
			Total:  tce.ToCollectEnergy,
			Status: tce.Status,
		}
	}
	for _, te := range totalEnergies {
		userId, _ := strconv.Atoi(te.UserID)
		h.totalEnergies[userId] = &totalEnergy{
			ID:    te.ID,
			Total: te.TotalEnergy,
		}
	}
	log.Infof("to_collect_energy size: %d", len(h.toCollectEnergies))
	log.Infof("total_energy size: %d", len(h.totalEnergies))
	return nil
}

func (h *Handler) Handle(hc *server.HttpCodec, body []byte) {
	userId, toCollectEnergyId, err := util.ParseUrl(string(hc.Path()))
	if err != nil {
		hc.Fail()
		return
	}

	// business handle
	query := h.handle(userId, toCollectEnergyId)
	if query != "" {
		_, err = h.DB.SQL().Exec(query)
		if err != nil {
			log.Infof("sync fail: %v, sql: %s", err, query)
			hc.Fail()
			return
		}
	}

	// resp
	hc.Suc()
}

func (h *Handler) handle(userId, toCollectEnergyId int) string {
	h.Lock()
	defer h.Unlock()

	if tce := h.toCollectEnergies[toCollectEnergyId]; tce != nil {
		if tce.Status == "all_collected" {
			goto build
		}
		if te := h.totalEnergies[userId]; te != nil {
			if tce.UserID == userId {
				te.Total += tce.Total
				h.totalEnergyChanges[userId] += tce.Total
				tce.Total = 0
				tce.Status = "all_collected"
			} else {
				if tce.Status == "collected_by_other" {
					goto build
				}
				count := int(math.Floor(float64(tce.Total) * 0.3))
				te.Total += count
				tce.Total -= count
				tce.Status = "collected_by_other"
				h.totalEnergyChanges[userId] += count
			}
			h.toCollectEnergyChanges[tce.ID] = true
		}
	}
build:
	// sync db
	count := atomic.AddUint64(&h.counter, 1)
	query := ""
	if count == 1 {
		h.startAt = time.Now().UnixNano()
	}
	if count%config.InsertBatchSize == 0 {
		query = h.buildSql()
	}
	if int(count) == 100_0000 {
		log.Infof("all costs: %d ms", (time.Now().UnixNano()-h.startAt)/1e6)
	}
	return query
}

func (h *Handler) buildSql() string {
	h.sqlBuffer.Reset()
	// to collect energy
	h.sqlBuffer.WriteString("update `" + TableToCollectEnergy + "` set `to_collect_energy` = `to_collect_energy` - FLOOR(`to_collect_energy`*0.3),`status` = 'collected_by_other' where id in (")
	for id := range h.toCollectEnergyChanges {
		h.sqlBuffer.WriteString("" + strconv.Itoa(id) + ",")
	}
	h.sqlBuffer.Truncate(h.sqlBuffer.Len() - 1)
	h.sqlBuffer.WriteString(");")
	// total energy
	for userId, total := range h.totalEnergyChanges {
		id := h.totalEnergies[userId].ID
		h.sqlBuffer.WriteString("update `" + TableTotalEnergy + "` set `total_energy` = `total_energy` + " + strconv.Itoa(total) + " where id = " + strconv.Itoa(id) + ";")
	}
	// clear
	h.toCollectEnergyChanges = map[int]bool{}
	h.totalEnergyChanges = map[int]int{}
	return h.sqlBuffer.String()
}
