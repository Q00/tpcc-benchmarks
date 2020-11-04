CREATE TABLE IF NOT EXISTS defaultdb.ORDERS_WID_DID (
  O_W_ID int,
  O_D_ID int,
  O_ID int,
  O_C_ID int NULL,
  O_CARRIER_ID int DEFAULT NULL,
  O_OL_CNT decimal(2,0),
  O_ALL_LOCAL DECIMAL(1,0),
  O_ENTRY_D timestamp DEFAULT CURRENT_TIMESTAMP,
  O_TOTAL_AMOUNT decimal(12,2),
  O_DELIVERY_D timestamp DEFAULT NULL,
  INDEX (O_C_ID, O_ID DESC),
  INDEX (O_CARRIER_ID, O_ID),
  PRIMARY KEY (O_W_ID, O_D_ID, O_ID),
  CONSTRAINT FK_ORDERS FOREIGN KEY (O_W_ID, O_D_ID, O_C_ID) REFERENCES defaultdb.CUSTOMER (C_W_ID, C_D_ID, C_ID)
);

CREATE TABLE IF NOT EXISTS defaultdb.ORDER_LINE_WID_DID (
  OL_W_ID int,
  OL_D_ID int,
  OL_O_ID int,
  OL_NUMBER int,
  OL_I_ID int,
  OL_DELIVERY_D timestamp,
  OL_AMOUNT decimal(6,2),
  OL_SUPPLY_W_ID int,
  OL_QUANTITY decimal(2,0),
  OL_DIST_INFO char(24),
  INDEX (OL_O_ID),
  INDEX (OL_I_ID),
  PRIMARY KEY (OL_W_ID, OL_D_ID, OL_O_ID, OL_NUMBER),
  CONSTRAINT FK_ORDER_LINE_WID_DID FOREIGN KEY (OL_W_ID, OL_D_ID, OL_O_ID) REFERENCES defaultdb.ORDERS_WID_DID (O_W_ID, O_D_ID, O_ID)
);

CREATE TABLE IF NOT EXISTS defaultdb.ORDER_ITEMS_CUSTOMERS_WID_DID (
  IC_W_ID int,
  IC_D_ID int,
  IC_C_ID int,
  IC_I_1_ID int,
  IC_I_2_ID int,
  INDEX (IC_W_ID, IC_D_ID, IC_C_ID),
  INDEX (IC_I_1_ID, IC_I_2_ID),
  PRIMARY KEY (IC_W_ID, IC_D_ID, IC_C_ID, IC_I_1_ID, IC_I_2_ID),
  CONSTRAINT FK_ORDERS FOREIGN KEY (IC_W_ID, IC_D_ID, IC_C_ID) REFERENCES defaultdb.CUSTOMER (C_W_ID, C_D_ID, C_ID)
);
