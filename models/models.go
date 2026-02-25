package models

import (
	"time"

	"yuyue-auth/config"
)

type Admin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Program struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Version     string    `gorm:"size:50" json:"version"`
	Language    string    `gorm:"size:50" json:"language"`
	SecretKey   string    `gorm:"size:100" json:"secret_key"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type License struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ProgramID   uint       `gorm:"index;not null" json:"program_id"`
	Program     Program    `gorm:"foreignKey:ProgramID" json:"program,omitempty"`
	LicenseKey  string     `gorm:"uniqueIndex;size:100;not null" json:"license_key"`
	Status      int        `gorm:"default:0" json:"status"`
	ActivatedAt *time.Time `json:"activated_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Duration    int        `gorm:"default:0" json:"duration"`
	DeviceInfo  string     `gorm:"size:500" json:"device_info"`
	Remark      string     `gorm:"size:255" json:"remark"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Order struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	OrderNo       string     `gorm:"uniqueIndex;size:64;not null" json:"order_no"`
	ProgramID     uint       `gorm:"index" json:"program_id"`
	Program       Program    `gorm:"foreignKey:ProgramID" json:"program,omitempty"`
	PackageID     *uint      `json:"package_id"`
	LicenseID     *uint      `json:"license_id"`
	License       *License   `gorm:"foreignKey:LicenseID" json:"license,omitempty"`
	Amount        float64    `gorm:"type:decimal(10,2)" json:"amount"`
	PaymentMethod string     `gorm:"size:20" json:"payment_method"`
	PaymentStatus int        `gorm:"default:0" json:"payment_status"`
	TradeNo       string     `gorm:"size:100" json:"trade_no"`
	BuyerEmail    string     `gorm:"size:100" json:"buyer_email"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Setting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Group     string    `gorm:"size:50;index;not null" json:"group"`
	Key       string    `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaymentConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PaymentType string    `gorm:"size:20;uniqueIndex;not null" json:"payment_type"`
	Enabled     bool      `gorm:"default:false" json:"enabled"`
	Config      string    `gorm:"type:text" json:"config"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Package struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProgramID   uint      `gorm:"index;not null" json:"program_id"`
	Program     Program   `gorm:"foreignKey:ProgramID" json:"program,omitempty"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Duration    int       `gorm:"default:0" json:"duration"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PiracyRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Domain    string    `gorm:"size:255;not null" json:"domain"`
	IP        string    `gorm:"size:50" json:"ip"`
	Message   string    `gorm:"size:500;not null" json:"message"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func AutoMigrate() error {
	return config.DB.AutoMigrate(
		&Admin{},
		&Program{},
		&License{},
		&Order{},
		&Setting{},
		&PaymentConfig{},
		&PiracyRecord{},
		&Package{},
	)
}

func InitDefaultSettings() error {
	defaults := []Setting{
		{Group: "system", Key: "site_title", Value: "鱼跃授权"},
		{Group: "system", Key: "site_description", Value: "专业的程序授权管理系统"},
		{Group: "system", Key: "site_keywords", Value: "授权系统,程序授权,激活码"},
		{Group: "system", Key: "footer_copyright", Value: "© 2026 鱼跃授权 All Rights Reserved"},
		{Group: "system", Key: "police_record", Value: ""},
		{Group: "system", Key: "site_icon", Value: ""},
		{Group: "system", Key: "site_favicon", Value: ""},
		{Group: "system", Key: "site_url", Value: ""},
		{Group: "system", Key: "unauth_page_html", Value: ""},
		{Group: "system", Key: "piracy_page_html", Value: ""},
	}

	for _, s := range defaults {
		config.DB.Where("`key` = ?", s.Key).FirstOrCreate(&s, Setting{Key: s.Key})
	}
	return nil
}

func GetSetting(key string) string {
	var setting Setting
	config.DB.Where("`key` = ?", key).First(&setting)
	return setting.Value
}

func SetSetting(key, value string) error {
	return config.DB.Model(&Setting{}).Where("`key` = ?", key).Update("value", value).Error
}
