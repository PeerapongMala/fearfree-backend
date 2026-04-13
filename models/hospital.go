package models // 👈 บรรทัดนี้สำคัญมาก!

type Hospital struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Type    string `json:"type"`
}

type UserHospital struct {
	ID         uint `json:"id" gorm:"primaryKey"`
	UserID     uint `json:"user_id" gorm:"uniqueIndex"`
	HospitalID uint `json:"hospital_id" gorm:"index"`

	Hospital *Hospital `json:"hospital" gorm:"foreignKey:HospitalID"`
}

func (Hospital) TableName() string     { return "hospital" }
func (UserHospital) TableName() string { return "user_hospital" }
