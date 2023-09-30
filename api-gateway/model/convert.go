package model

// Convert2Entrust 转换成entrust
//func (c *BindParam) Convert2Entrust() *Entrust {
//	var entrustStatus int64
//	// 市价委托,则设置成已成交；限价委托则设置成未成交
//	if c.EntrustProp == EntrustPropTypeMarketPrice {
//		entrustStatus = EntrustStatusTypeDeal
//	} else {
//		entrustStatus = EntrustStatusTypeUnDeal
//	}
//
//	var fee float64
//	if c.EntrustBs == EntrustBsTypeBuy { // 买入手续费
//		fee = c.Price * float64(c.Amount) * c.SysParam.BuyFee
//		if c.SysParam.MiniChargeFee > 0 && fee < c.SysParam.MiniChargeFee {
//			fee = c.SysParam.MiniChargeFee
//		}
//	} else if c.EntrustBs == EntrustBsTypeSell { // 卖出手续费
//		fee = c.Price * float64(c.Amount) * c.SysParam.SellFee
//		if c.SysParam.MiniChargeFee > 0 && fee < c.SysParam.MiniChargeFee {
//			fee = c.SysParam.MiniChargeFee
//		}
//	}
//	var positionID int64
//	if c.EntrustBs == EntrustBsTypeSell {
//		positionID = c.Position.ID
//	}
//	return &Entrust{
//		//ID:          id_gen.GetNextID(),
//		UID:         c.UserID,
//		ContractID:  c.Contract.ID,
//		OrderTime:   time.Now(),
//		StockCode:   c.StockData.Code,
//		StockName:   c.StockData.Name,
//		Amount:      c.Amount,
//		Price:       c.Price,
//		Balance:     float64(c.Amount) * c.Price,
//		Status:      entrustStatus, // 委托状态:1未成交 2成交 3已撤单
//		EntrustBS:   c.EntrustBs,   // 交易类型:1买入 2卖出
//		EntrustProp: c.EntrustProp, // 委托类型:1限价 2市价
//		PositionID:  positionID,    // 持仓id
//		Fee:         fee,           // 总交易费用
//	}
//}

//
//// EntrustConvert2Buy Entrust结构体转Buy
//func EntrustConvert2Buy(entrust *Entrust, position *Position) *Buy {
//	entrust.PositionID = entrust.ID
//	if position.ID != 0 {
//		entrust.PositionID = position.ID
//	}
//	return &Buy{
//		EntrustID:   entrust.ID,
//		UID:         entrust.UID,
//		ContractID:  entrust.ContractID,
//		OrderTime:   entrust.OrderTime,
//		StockCode:   entrust.StockCode,
//		StockName:   entrust.StockName,
//		Price:       entrust.Price,
//		Amount:      entrust.Amount,
//		Balance:     entrust.Balance,
//		EntrustBS:   entrust.EntrustBS,
//		EntrustProp: entrust.EntrustProp,
//		Fee:         entrust.Fee,
//		PositionID:  entrust.PositionID,
//	}
//}

//
//// EntrustConvert2Sell Entrust结构体转sell
//func EntrustConvert2Sell(entrust *Entrust, position *Position) *Sell {
//	return &Sell{
//		EntrustID:     entrust.ID,
//		UID:           entrust.UID,
//		ContractID:    entrust.ContractID,
//		OrderTime:     entrust.OrderTime,
//		StockCode:     entrust.StockCode,
//		StockName:     entrust.StockName,
//		Price:         entrust.Price,
//		Amount:        entrust.Amount,
//		Balance:       entrust.Price * float64(entrust.Amount),
//		PositionPrice: position.Price,
//		EntrustBS:     entrust.EntrustBS,
//		Profit:        (entrust.Price - position.Price) * float64(entrust.Amount),
//		EntrustProp:   entrust.EntrustProp,
//		Fee:           entrust.Fee,
//		PositionID:    position.ID,
//		Mode:          entrust.Mode,
//		Reason:        entrust.Reason,
//	}
//}

func EntrustConvert2Position(entrust *Entrust) *Position {
	return &Position{
		UID:          entrust.UID,
		ContractID:   entrust.ContractID,
		OrderTime:    entrust.OrderTime,
		StockCode:    entrust.StockCode,
		StockName:    entrust.StockName,
		Price:        entrust.Price,
		Amount:       entrust.Amount,
		Balance:      entrust.Balance,
		FreezeAmount: entrust.Amount,
	}
}
