package driver

type BleCommand string

const (
	ATRESET        BleCommand = "AT+QRST\r\n"
	ATVERSION      BleCommand = "AT+QVERSION\r\n"
	ATINIT_2       BleCommand = "AT+QBLEINIT=2\r\n"
	ATADV          BleCommand = "AT+QBLEADVPARAM=150,150\r\n"
	ATGATTSSRV     BleCommand = "AT+QBLEGATTSSRV=fff1\r\n"
	ATGATTSCHAR    BleCommand = "AT+QBLEGATTSCHAR=fff2\r\n"
	ATGATTSSRVDONE BleCommand = "AT+QBLEGATTSSRVDONE\r\n"
	ATNAME         BleCommand = "AT+QBLENAME=QuecHCM111Z\r\n"
	ATADDR         BleCommand = "AT+QBLEADDR?\r\n"
	ATADVSTART     BleCommand = "AT+QBLEADVSTART\r\n"
)
