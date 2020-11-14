package dao

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mychewcents/tpcc-benchmarks/cockroachdb/internal/internal/internal/models"
)

// StockDao creates the Dao object for the Stock table
type StockDao interface {
	GetStockDetails(tx *sql.Tx, districtID int, orderLineItems map[int]*models.NewOrderOrderLineItem) (totalAmount float64, err error)
	UpdateStockDetails(tx *sql.Tx, orderLineItems map[int]*models.NewOrderOrderLineItem) error
}

type stockDaoImpl struct {
	db *sql.DB
}

// CreateStockDao creates the new StockDao object
func CreateStockDao(db *sql.DB) StockDao {
	return &stockDaoImpl{db: db}
}

// GetStockDetails gets the stock details
func (s *stockDaoImpl) GetStockDetails(tx *sql.Tx, districtID int, orderLineItems map[int]*models.NewOrderOrderLineItem) (totalAmount float64, err error) {
	var itemsWhereClause strings.Builder

	for key, value := range orderLineItems {
		itemsWhereClause.WriteString(fmt.Sprintf("(%d, %d),", value.SupplierWarehouseID, key))
	}

	itemsWhereClauseString := itemsWhereClause.String()
	itemsWhereClauseString = itemsWhereClauseString[:len(itemsWhereClauseString)-1]

	sqlStatement := fmt.Sprintf("SELECT S_I_ID, S_I_NAME, S_I_PRICE, S_QUANTITY, S_YTD, S_ORDER_CNT, S_DIST_%02d FROM STOCK WHERE (S_W_ID, S_I_ID) IN %s",
		districtID, itemsWhereClauseString)
	rows, err := tx.Query(sqlStatement)
	if err == sql.ErrNoRows {
		return 0.0, fmt.Errorf("no rows found for the items ids passed")
	}
	if err != nil {
		return 0.0, fmt.Errorf("error in getting the stock details for the items. \nquery: %s. \nErr: %v", sqlStatement, err)
	}

	var name, data string
	var price, currYTD float64
	var id, startStock, currOrderCnt int

	for rows.Next() {
		if err := rows.Scan(&id, &name, &price, &startStock, &currYTD, &currOrderCnt, &data); err != nil {
			return 0.0, fmt.Errorf("error in scanning the results for the items. Err: %v", err)
		}

		if value, ok := orderLineItems[id]; ok {
			value.Name = name
			value.Price = price
			value.StartStock = startStock
			value.CurrYTD = currYTD
			value.CurrOrderCnt = currOrderCnt
			value.Data = data

			adjustedQty := startStock - value.Quantity
			if adjustedQty < 10 {
				adjustedQty += 100
			}
			value.FinalStock = adjustedQty

			value.Amount = price * float64(value.Quantity)
			totalAmount += value.Amount
		}
	}

	return
}

func (s *stockDaoImpl) UpdateStockDetails(tx *sql.Tx, orderLineItems map[int]*models.NewOrderOrderLineItem) error {
	var stockOrderItemIdentifiers, stockQuantityUpdates, stockYTDUpdates, stockOrderCountUpdates, stockRemoteCountUpdates strings.Builder

	var itemIdentifier string
	whenClauseFormat := "WHEN %s THEN %d "

	idx := 0
	for key, value := range orderLineItems {
		itemIdentifier = fmt.Sprintf("(%d, %d)", value.SupplierWarehouseID, key)

		stockOrderItemIdentifiers.WriteString(fmt.Sprintf("%s,", itemIdentifier))
		stockQuantityUpdates.WriteString(fmt.Sprintf(whenClauseFormat, itemIdentifier, value.FinalStock))
		stockYTDUpdates.WriteString(fmt.Sprintf(whenClauseFormat, itemIdentifier, int(value.CurrYTD)+value.Quantity))
		stockOrderCountUpdates.WriteString(fmt.Sprintf(whenClauseFormat, itemIdentifier, value.CurrOrderCnt+1))
		stockRemoteCountUpdates.WriteString(fmt.Sprintf(whenClauseFormat, itemIdentifier, value.IsRemote))
		idx++
	}

	stockOrderItemIdentifiersString := stockOrderItemIdentifiers.String()
	stockOrderItemIdentifiersString = stockOrderItemIdentifiersString[:len(stockOrderItemIdentifiersString)-1]

	stockUpdateStatement := fmt.Sprintf(`
			UPDATE STOCK 
				SET S_QUANTITY = CASE (S_W_ID, S_I_ID) %s END, 
				S_YTD = CASE (S_W_ID, S_I_ID) %s END, 
				S_ORDER_CNT = CASE (S_W_ID, S_I_ID) %s END, 
				S_REMOTE_CNT = CASE (S_W_ID, S_I_ID) %s END 
			WHERE (S_W_ID, S_I_ID) IN (%s)`,
		stockQuantityUpdates.String(),
		stockYTDUpdates.String(),
		stockOrderCountUpdates.String(),
		stockRemoteCountUpdates.String(),
		stockOrderItemIdentifiersString,
	)

	if _, err := tx.Exec(stockUpdateStatement); err != nil {
		return err
	}

	return nil
}
