package player

func (self *TPlayer) UpdateOnRecvClientData() {
	self.Mail.SendSvrMailAll()
}
