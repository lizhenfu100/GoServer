package player

func (self *TPlayer) UpdateOnRecvClientData() {
	self._HandleAsyncNotify()
	self.Mail.SendSvrMailAll()
}
