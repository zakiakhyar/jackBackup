package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"
)

var (
	dbType string
)

func main() {
	// Load variabel dari file .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Membaca variabel konfigurasi dari .env
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	backupDir := os.Getenv("BACKUP_DIR")
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")
	dbType = os.Getenv("DB_VERSION")

	// Membaca opsi command line
	flag.StringVar(&dbType, "db", dbType, "Jenis database (postgres atau mysql)")
	flag.Parse()

	// Memilih perintah dump yang sesuai berdasarkan jenis database
	var dumpCommand string
	switch dbType {
	case "postgres":
		dumpCommand = "pg_dump"
	case "mysql":
		dumpCommand = "mysqldump"
	default:
		log.Fatalf("Jenis database '%s' tidak didukung.", dbType)
	}

	// Koneksi ke database
	dbPortOption := ""
	if dbPort != "" {
		dbPortOption = fmt.Sprintf("port=%s", dbPort)
	}
	dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s", dbUser, dbPassword, dbName, dbHost, dbPortOption)
	db, err := sql.Open(dbType, dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Nama file cadangan dengan format "nama_database-tanggal_waktu.sql"
	backupFileName := fmt.Sprintf("%s-%s.sql", dbName, time.Now().Format("2006-01-02_15-04-05"))

	// Path lengkap ke file cadangan
	backupFilePath := fmt.Sprintf("%s/%s", backupDir, backupFileName)

	// Perintah pg_dump atau mysqldump untuk membuat cadangan
	cmd := exec.Command(dumpCommand, "-u", dbUser, "--password="+dbPassword, dbName)
	outfile, err := os.Create(backupFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Mengirim notifikasi ke Telegram setelah selesai backup
	if telegramBotToken != "" && telegramChatID != "" {
		message := fmt.Sprintf("Backup database '%s' (%s) telah selesai.", dbName, dbType)
		sendTelegramNotification(telegramBotToken, telegramChatID, message)
	}
}

func sendTelegramNotification(botToken, chatID, message string) {
	// Mengirim pesan notifikasi ke Telegram menggunakan bot
	cmd := exec.Command("curl", "-s", "-X", "POST",
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
		"-d", fmt.Sprintf("chat_id=%s", chatID),
		"-d", fmt.Sprintf("text=%s", message))
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
