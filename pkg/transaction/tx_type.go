package transaction

const (
	TxType_DEPOSIT  = "DEPOSIT"
	TxType_WITHDRAW = "WITHDRAW"
)

func GetTxType(txTypeCode int) string {
	switch txTypeCode {
	case 0:
		return TxType_DEPOSIT
	case 1:
		return TxType_WITHDRAW
	default:
		return ""
	}
}
