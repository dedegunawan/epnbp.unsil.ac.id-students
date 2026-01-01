package mahasiswa_manager

import "github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"

func (mahasiswa *Mahasiswa) LoadTagihan() {

}

func (mahasiswa *Mahasiswa) IsAllPaid() bool {
	return false
}

func (mahasiswa *Mahasiswa) IsTagihanGenerated() bool {
	return true
}

func (mahasiswa *Mahasiswa) TagihanHarusDibayar() []entity.StudentBill {
	var studentBills []entity.StudentBill
	return studentBills
}
func (mahasiswa *Mahasiswa) HistoryTagihan() []entity.StudentBill {
	var studentBills []entity.StudentBill
	return studentBills
}
