package main

import "time"

type Pasien struct {
	IDPasien     uint      `gorm:"primaryKey;column:id_pasien" json:"id_pasien"`
	NIK          string    `gorm:"column:nik;unique" json:"nik"`
	NamaPasien   string    `gorm:"column:nama_pasien" json:"nama_pasien"`
	TempatLahir  string    `gorm:"column:tempat_lahir" json:"tempat_lahir"`
	TanggalLahir string    `gorm:"column:tanggal_lahir" json:"tanggal_lahir"`
	JenisKelamin string    `gorm:"column:jenis_kelamin" json:"jenis_kelamin"`
	Agama        string    `gorm:"column:agama" json:"agama"`
	StatusKawin  string    `gorm:"column:status_kawin" json:"status_kawin"`
	GolDarah     string    `gorm:"column:gol_darah" json:"gol_darah"`
	Alamat       string    `gorm:"column:alamat" json:"alamat"`
	NoWhatsapp   string    `gorm:"column:no_whatsapp" json:"no_whatsapp"`
	Email        string    `gorm:"column:email" json:"email"`
	CreatedAt    time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Pasien) TableName() string { return "pasien" }

// Sisa struct (Spesialis, Dokter, Ruangan, Pendaftaran, dll.) tetap sama seperti kode sebelumnya
type Spesialis struct {
	IDSpesialis   uint   `gorm:"primaryKey;column:id_spesialis"`
	NamaSpesialis string `gorm:"column:nama_spesialis"`
}

func (Spesialis) TableName() string { return "spesialis" }

type Dokter struct {
	IDDokter    uint      `gorm:"primaryKey;autoIncrement"`
	NamaDokter  string    `gorm:"type:varchar(100);not null"`
	Username    string    `gorm:"type:varchar(50);unique;not null"` // Tambahkan ini
	Password    string    `gorm:"type:varchar(255);not null"`       // Tambahkan ini
	IDSpesialis uint      `gorm:"not null"`
	NoTelp      string    `gorm:"type:varchar(20)"`
	Spesialis   Spesialis `gorm:"foreignKey:IDSpesialis"`
}

func (Dokter) TableName() string { return "dokter" }

type Ruangan struct {
	IDRuangan       uint   `gorm:"primaryKey;column:id_ruangan"`
	NamaRuangan     string `gorm:"column:nama_ruangan"`
	KategoriRuangan string `gorm:"column:kategori_ruangan"`
}

func (Ruangan) TableName() string { return "ruangan" }

type Pendaftaran struct {
	IDPendaftaran      uint      `gorm:"primaryKey;column:id_pendaftaran"`
	IDPasien           uint      `gorm:"column:id_pasien"`
	IDDokter           uint      `gorm:"column:id_dokter"`
	IDRuangan          uint      `gorm:"column:id_ruangan"`
	TanggalPendaftaran time.Time `gorm:"column:tanggal_pendaftaran;default:CURRENT_TIMESTAMP"`
	NomorAntrean       int       `gorm:"column:nomor_antrean"`
	JenisLayanan       string    `gorm:"column:jenis_layanan"`
	KeluhanAwal        string    `gorm:"column:keluhan_awal"`
	NamaRegistrator    string    `gorm:"column:nama_registrator"`
	StatusAntrean      string    `gorm:"column:status_antrean;default:'Menunggu'"`
	Pasien             Pasien    `gorm:"foreignKey:IDPasien;references:IDPasien"`
	Dokter             Dokter    `gorm:"foreignKey:IDDokter;references:IDDokter"`
	Ruangan            Ruangan   `gorm:"foreignKey:IDRuangan;references:IDRuangan"`
}

func (Pendaftaran) TableName() string { return "pendaftaran" }

type Laboratorium struct {
	IDLab            uint        `gorm:"primaryKey;column:id_lab"`
	IDPendaftaran    uint        `gorm:"column:id_pendaftaran"`
	JenisPemeriksaan string      `gorm:"column:jenis_pemeriksaan"`
	HasilPemeriksaan string      `gorm:"column:hasil_pemeriksaan"`
	TanggalPeriksa   time.Time   `gorm:"column:tanggal_periksa"`
	Pendaftaran      Pendaftaran `gorm:"foreignKey:IDPendaftaran;references:IDPendaftaran"`
}

func (Laboratorium) TableName() string { return "laboratorium" }

type Farmasi struct {
	IDResep           uint        `gorm:"primaryKey;column:id_resep"`
	IDPendaftaran     uint        `gorm:"column:id_pendaftaran"`
	DetailObat        string      `gorm:"column:detail_obat"`
	StatusPengambilan string      `gorm:"column:status_pengambilan;default:'Belum Diambil'"`
	Pendaftaran       Pendaftaran `gorm:"foreignKey:IDPendaftaran;references:IDPendaftaran"`
}

func (Farmasi) TableName() string { return "farmasi" }
