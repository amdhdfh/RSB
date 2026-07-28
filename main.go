package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func initDatabase() {
	dsn := "root:@tcp(127.0.0.1:3306)/rsb?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi database:", err)
	}

	DB.AutoMigrate(
		&Pasien{},
		&Spesialis{},
		&Dokter{},
		&Ruangan{},
		&Pendaftaran{},
		&Laboratorium{},
		&Farmasi{},
	)
}

func main() {
	initDatabase()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// ================= GATEWAY UTAMA =================
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "landing.html", nil)
	})

	// ================= ROLE: PORTAL PASIEN MANDIRI =================
	r.GET("/pasien/daftar", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pasien_daftar.html", nil)
	})

	r.POST("/pasien/daftar", func(c *gin.Context) {
		pasien := Pasien{
			NIK:          c.PostForm("nik"),
			NamaPasien:   c.PostForm("nama_pasien"),
			TempatLahir:  c.PostForm("tempat_lahir"),
			TanggalLahir: c.PostForm("tanggal_lahir"),
			JenisKelamin: c.PostForm("jenis_kelamin"),
			Agama:        c.PostForm("agama"),
			StatusKawin:  c.PostForm("status_kawin"),
			GolDarah:     c.PostForm("gol_darah"),
			Alamat:       c.PostForm("alamat"),
			NoWhatsapp:   c.PostForm("no_whatsapp"),
			Email:        c.PostForm("email"),
		}

		if err := DB.Create(&pasien).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				c.String(http.StatusBadRequest, "Gagal Pendaftaran: Nomor NIK (%s) sudah terdaftar di sistem kami.", pasien.NIK)
				return
			}
			c.String(http.StatusInternalServerError, "Gagal registrasi online: %v", err)
			return
		}

		c.HTML(http.StatusOK, "pasien_sukses.html", gin.H{"pasien": pasien})
	})

	// ================= ROLE: PORTAL DOKTER (AUTHENTICATION & DASHBOARD) =================

	// 1. Halaman Register Dokter
	r.GET("/dokter/register", func(c *gin.Context) {
		var spesialis []Spesialis
		DB.Find(&spesialis)
		c.HTML(http.StatusOK, "dokter_register.html", gin.H{"spesialis": spesialis})
	})

	// 2. Aksi Proses Register Dokter
	r.POST("/dokter/register", func(c *gin.Context) {
		idSpesialisRaw, _ := strconv.ParseUint(c.PostForm("spesialis"), 10, 32)

		dokter := Dokter{
			NamaDokter:  c.PostForm("nama"),
			Username:    c.PostForm("username"),
			Password:    c.PostForm("password"), // Catatan: Pada aplikasi produksi asli, gunakan hashing bcrypt!
			IDSpesialis: uint(idSpesialisRaw),
			NoTelp:      c.PostForm("telp"),
		}

		if err := DB.Create(&dokter).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				c.String(http.StatusBadRequest, "Gagal Registrasi: Username sudah digunakan.")
				return
			}
			c.String(http.StatusInternalServerError, "Gagal membuat akun dokter: %v", err)
			return
		}
		c.Redirect(http.StatusFound, "/dokter/login")
	})

	// 3. Halaman Login Dokter
	r.GET("/dokter/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dokter_login.html", nil)
	})

	// 4. Aksi Proses Login Dokter (Set Cookie Sesi)
	r.POST("/dokter/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		var dokter Dokter
		err := DB.Where("username = ? AND password = ?", username, password).First(&dokter).Error
		if err != nil {
			c.HTML(http.StatusUnauthorized, "dokter_login.html", gin.H{"error": "Username atau Password salah!"})
			return
		}

		// Simpan ID Dokter di cookie browser selama 2 jam
		c.SetCookie("dokter_session", strconv.Itoa(int(dokter.IDDokter)), 7200, "/", "", false, true)
		c.Redirect(http.StatusFound, "/dokter/dashboard")
	})

	// 5. Halaman Dashboard Utama Dokter (Secure via Cookie)
	r.GET("/dokter/dashboard", func(c *gin.Context) {
		cookie, err := c.Cookie("dokter_session")
		if err != nil {
			c.Redirect(http.StatusFound, "/dokter/login")
			return
		}

		idDokterRaw, _ := strconv.ParseUint(cookie, 10, 32)
		dokterID := uint(idDokterRaw)

		var selectedDokter Dokter
		if err := DB.Preload("Spesialis").First(&selectedDokter, dokterID).Error; err != nil {
			c.Redirect(http.StatusFound, "/dokter/login")
			return
		}

		var pendaftarans []Pendaftaran
		today := time.Now().Format("2006-01-02")
		DB.Preload("Pasien").Preload("Ruangan").
			Where("id_dokter = ? AND DATE(tanggal_pendaftaran) = ?", dokterID, today).
			Order("nomor_antrean ASC").Find(&pendaftarans)

		c.HTML(http.StatusOK, "dokter_dashboard.html", gin.H{
			"selectedDokter": selectedDokter,
			"pendaftarans":   pendaftarans,
		})
	})

	// 6. Aksi Merubah Status Kunjungan Pasien
	r.POST("/dokter/ubah-status", func(c *gin.Context) {
		cookie, err := c.Cookie("dokter_session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		nomorAntrean := c.PostForm("nomor_antrean")
		statusBaru := c.PostForm("status")
		today := time.Now().Format("2006-01-02")

		DB.Model(&Pendaftaran{}).
			Where("id_dokter = ? AND nomor_antrean = ? AND DATE(tanggal_pendaftaran) = ?", cookie, nomorAntrean, today).
			Update("status_antrean", statusBaru)

		c.Redirect(http.StatusFound, "/dokter/dashboard")
	})

	// 7. Aksi Logout Dokter (Hapus Cookie Sesi)
	r.GET("/dokter/logout", func(c *gin.Context) {
		c.SetCookie("dokter_session", "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/dokter/login")
	})

	// ================= ROLE: PORTAL ADMIN (MENGELOLA DATA) =================
	admin := r.Group("/admin")
	{
		admin.GET("/pasien", func(c *gin.Context) {
			var pasiens []Pasien
			DB.Order("id_pasien DESC").Find(&pasiens)
			c.HTML(http.StatusOK, "index.html", gin.H{"pasiens": pasiens})
		})

		admin.GET("/pasien/tambah", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/admin/pasien")
		})

		admin.POST("/pasien/tambah", func(c *gin.Context) {
			pasien := Pasien{
				NIK:          c.PostForm("nik"),
				NamaPasien:   c.PostForm("nama_pasien"),
				TempatLahir:  c.PostForm("tempat_lahir"),
				TanggalLahir: c.PostForm("tanggal_lahir"),
				JenisKelamin: c.PostForm("jenis_kelamin"),
				NoWhatsapp:   c.PostForm("no_whatsapp"),
				Email:        c.PostForm("email"),
				Alamat:       c.PostForm("alamat"),
			}
			if err := DB.Create(&pasien).Error; err != nil {
				if strings.Contains(err.Error(), "Duplicate entry") {
					c.String(http.StatusBadRequest, "Gagal Input Manual: Nomor NIK sudah terdaftar.")
					return
				}
				c.String(http.StatusInternalServerError, "Gagal simpan database: %v", err)
				return
			}
			c.Redirect(http.StatusFound, "/admin/pasien")
		})

		admin.POST("/pasien/hapus/:id", func(c *gin.Context) {
			id := c.Param("id")
			DB.Delete(&Pasien{}, id)
			c.Redirect(http.StatusFound, "/admin/pasien")
		})

		admin.GET("/dokter", func(c *gin.Context) {
			var dokters []Dokter
			var spesialis []Spesialis
			DB.Preload("Spesialis").Find(&dokters)
			DB.Find(&spesialis)
			c.HTML(http.StatusOK, "dokter.html", gin.H{"dokters": dokters, "spesialis": spesialis})
		})

		// Aksi Tambah Dokter via Admin (agar singkron dengan templates/dokter.html)
		admin.POST("/dokter/tambah", func(c *gin.Context) {
			idSpesialisRaw, _ := strconv.ParseUint(c.PostForm("spesialis"), 10, 32)
			namaDokter := c.PostForm("nama")

			// Generate username default berbasis nama tanpa spasi untuk memudahkan admin
			defaultUser := strings.ToLower(strings.ReplaceAll(namaDokter, " ", ""))

			dokter := Dokter{
				NamaDokter:  namaDokter,
				Username:    defaultUser,
				Password:    "dokter123", // password bawaan default
				IDSpesialis: uint(idSpesialisRaw),
				NoTelp:      c.PostForm("telp"),
			}
			DB.Create(&dokter)
			c.Redirect(http.StatusFound, "/admin/dokter")
		})

		admin.GET("/pendaftaran", func(c *gin.Context) {
			var pasiens []Pasien
			var dokters []Dokter
			var ruangans []Ruangan
			var pendaftarans []Pendaftaran

			DB.Find(&pasiens)
			DB.Preload("Spesialis").Find(&dokters)
			DB.Find(&ruangans)
			DB.Preload("Pasien").Preload("Dokter.Spesialis").Preload("Ruangan").Find(&pendaftarans)

			c.HTML(http.StatusOK, "pendaftaran.html", gin.H{
				"pasiens":      pasiens,
				"dokters":      dokters,
				"ruangans":     ruangans,
				"pendaftarans": pendaftarans,
			})
		})

		admin.POST("/pendaftaran/tambah", func(c *gin.Context) {
			var totalAntrean int64
			today := time.Now().Format("2006-01-02")
			DB.Model(&Pendaftaran{}).Where("DATE(tanggal_pendaftaran) = ?", today).Count(&totalAntrean)
			nomorAntrean := int(totalAntrean) + 1

			idPasienRaw, _ := strconv.ParseUint(c.PostForm("id_pasien"), 10, 32)
			idDokterRaw, _ := strconv.ParseUint(c.PostForm("id_dokter"), 10, 32)
			idRuanganRaw, _ := strconv.ParseUint(c.PostForm("id_ruangan"), 10, 32)

			pendaftaran := Pendaftaran{
				IDPasien:           uint(idPasienRaw),
				IDDokter:           uint(idDokterRaw),
				IDRuangan:          uint(idRuanganRaw),
				NomorAntrean:       nomorAntrean,
				JenisLayanan:       c.PostForm("jenis_layanan"),
				KeluhanAwal:        c.PostForm("keluhan"),
				NamaRegistrator:    c.PostForm("registrator"),
				StatusAntrean:      "Menunggu",
				TanggalPendaftaran: time.Now(),
			}

			DB.Omit("Pasien", "Dokter", "Ruangan").Create(&pendaftaran)
			c.Redirect(http.StatusFound, "/admin/pendaftaran")
		})

		admin.GET("/layanan", func(c *gin.Context) {
			var labs []Laboratorium
			var farmasis []Farmasi

			DB.Preload("Pendaftaran.Pasien").Find(&labs)
			DB.Preload("Pendaftaran.Pasien").Find(&farmasis)

			c.HTML(http.StatusOK, "layanan.html", gin.H{
				"labs":     labs,
				"farmasis": farmasis,
			})
		})
	}

	log.Println("Aplikasi SIRS Berjalan Eksklusif di http://localhost:8085")
	r.Run(":8085")
}
