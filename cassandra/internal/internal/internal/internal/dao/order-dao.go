package dao

import (
	"github.com/gocql/gocql"
	"github.com/mychewcents/ddbms-project/cassandra/internal/common"
	"github.com/mychewcents/ddbms-project/cassandra/internal/internal/internal/internal/datamodel/table"
	"github.com/mychewcents/ddbms-project/cassandra/internal/internal/internal/internal/datamodel/view"
	"log"
	"time"
)

type OrderDao interface {
	InsertOrder(ot *table.OrderTab, chComplete chan bool)
	GetOldestUnDeliveredOrder(oWId int, oDId int) *view.OrderByCarrierView
	UpdateOrderCAS(oWId int, oDId int, oId gocql.UUID, oCarrierId int) bool
}

type orderDaoImpl struct {
	cassandraSession *common.CassandraSession
}

func NewOrderDao(cassandraSession *common.CassandraSession) OrderDao {
	return &orderDaoImpl{cassandraSession: cassandraSession}
}

func (o *orderDaoImpl) InsertOrder(ot *table.OrderTab, chComplete chan bool) {
	stmt := "INSERT INTO " +
		"order_tab (o_w_id, o_d_id, o_id, o_c_id, o_c_name, o_carrier_id, ol_delivery_d, o_ol_count, o_ol_total_amount, o_all_local, o_entry_d) " +
		"VALUES (?,?,?,?,?,?,?,?,?,?,?)"

	query := o.cassandraSession.WriteSession.Query(stmt, ot.OWId, ot.ODId, ot.OId, ot.OCId, ot.OCName.GetNameString(), ot.OCarrierId, ot.OlDeliveryD,
		ot.OOlCount, ot.OOlTotalAmount, ot.OAllLocal, ot.OEntryD)

	err := query.Exec()
	if err != nil {
		log.Fatalf("InsertOrder. ot=%v, err%v", ot, err)
	}

	chComplete <- true
}

func (o *orderDaoImpl) GetOldestUnDeliveredOrder(oWId int, oDId int) *view.OrderByCarrierView {
	query := o.cassandraSession.ReadSession.Query("SELECT * "+
		"from order_by_customer_view "+
		"where o_w_id=? AND o_d_id=? LIMIT 1", oWId, oDId)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		log.Fatalf("ERROR GetOldestUndeliveredOrder error in query execution. oWId=%v, oDId=%v, err=%v\n", oWId, oDId, err)
	}

	ov, err := view.MakeOrderByCarrierView(result)
	if err != nil {
		log.Fatalf("ERROR GetOldestUndeliveredOrder error making customer. oWId=%v, oDId=%v, err=%v\n", oWId, oDId, err)
	}

	return ov
}

func (o *orderDaoImpl) UpdateOrderCAS(oWId int, oDId int, oId gocql.UUID, oCarrierId int) bool {
	query := o.cassandraSession.WriteSession.Query("UPDATE order_tab "+
		"SET o_carrier_id=?, ol_delivery_d=? "+
		"WHERE o_w_id=? and o_d_id=? AND o_id=? "+
		"IF o_carrier_id=-1=? AND ol_delivery_d=null", oCarrierId, time.Now(),
		oWId, oDId, oId)

	applied, err := query.ScanCAS()
	if err != nil {
		log.Fatalf("ERROR UpdateOrderCAS quering. err=%v\n", err)
		return false
	}

	return applied
}
