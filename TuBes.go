package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ==================== STRUCTS ====================

type Subscription struct {
	ID            int       `json:"id"`
	Nama          string    `json:"nama"`
	Kategori      string    `json:"kategori"`
	BiayaBulanan  float64   `json:"biaya_bulanan"`
	MetodePembayaran string `json:"metode_pembayaran"`
	TanggalJatuhTempo int   `json:"tanggal_jatuh_tempo"` // tanggal dalam bulan (1-31)
	Status        string    `json:"status"`               // aktif / nonaktif
	TanggalMulai  string    `json:"tanggal_mulai"`
	Catatan       string    `json:"catatan"`
}

type Database struct {
	Subscriptions []Subscription `json:"subscriptions"`
	NextID        int            `json:"next_id"`
}

// ==================== GLOBAL ====================

var db Database
var dbFile = "subscriptions.json"
var scanner = bufio.NewScanner(os.Stdin)

// ==================== WARNA TERMINAL ====================

const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	BgBlue  = "\033[44m"
	BgGreen = "\033[42m"
)

// ==================== UTILITAS ====================

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printLine(char string, n int) {
	fmt.Println(strings.Repeat(char, n))
}

func printHeader(title string) {
	clearScreen()
	printLine("═", 60)
	fmt.Printf("%s%s%s  %-54s%s%s\n", Bold, BgBlue, White, title, Reset, "")
	printLine("═", 60)
}

func input(prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func inputFloat(prompt string) float64 {
	for {
		s := input(prompt)
		v, err := strconv.ParseFloat(s, 64)
		if err == nil && v >= 0 {
			return v
		}
		fmt.Println(Red + "  Input tidak valid. Masukkan angka positif." + Reset)
	}
}

func inputInt(prompt string, min, max int) int {
	for {
		s := input(prompt)
		v, err := strconv.Atoi(s)
		if err == nil && v >= min && v <= max {
			return v
		}
		fmt.Printf(Red+"  Input tidak valid. Masukkan angka antara %d-%d.\n"+Reset, min, max)
	}
}

func formatRupiah(amount float64) string {
	intPart := int64(amount)
	s := strconv.FormatInt(intPart, 10)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return "Rp " + result
}

func tekanEnterLanjut() {
	fmt.Print("\n  " + Cyan + "Tekan ENTER untuk melanjutkan..." + Reset)
	scanner.Scan()
}

// ==================== DATABASE ====================

func loadDatabase() {
	file, err := os.ReadFile(dbFile)
	if err != nil {
		db = Database{NextID: 1}
		return
	}
	json.Unmarshal(file, &db)
	if db.NextID == 0 {
		db.NextID = 1
	}
}

func saveDatabase() {
	data, _ := json.MarshalIndent(db, "", "  ")
	os.WriteFile(dbFile, data, 0644)
}

func getActiveSubscriptions() []Subscription {
	var result []Subscription
	for _, s := range db.Subscriptions {
		if s.Status == "aktif" {
			result = append(result, s)
		}
	}
	return result
}

func findByID(id int) *Subscription {
	for i := range db.Subscriptions {
		if db.Subscriptions[i].ID == id {
			return &db.Subscriptions[i]
		}
	}
	return nil
}

func daysUntilDue(tanggalJT int) int {
	now := time.Now()
	thisMonth := time.Date(now.Year(), now.Month(), tanggalJT, 0, 0, 0, 0, time.Local)
	if now.Day() > tanggalJT {
		nextMonth := now.Month() + 1
		year := now.Year()
		if nextMonth > 12 {
			nextMonth = 1
			year++
		}
		thisMonth = time.Date(year, nextMonth, tanggalJT, 0, 0, 0, 0, time.Local)
	}
	diff := thisMonth.Sub(now)
	return int(math.Ceil(diff.Hours() / 24))
}

// ==================== TAMPILAN TABEL ====================

func printSubscriptionTable(subs []Subscription) {
	if len(subs) == 0 {
		fmt.Println(Yellow + "\n  Tidak ada data langganan." + Reset)
		return
	}
	fmt.Printf("\n%s%-4s %-22s %-14s %-12s %-6s %-10s %-8s%s\n",
		Bold+Cyan, "ID", "Nama", "Kategori", "Biaya/Bulan", "Tgl JT", "Pembayaran", "Status", Reset)
	printLine("─", 82)
	for _, s := range subs {
		statusColor := Green
		if s.Status == "nonaktif" {
			statusColor = Red
		}
		days := daysUntilDue(s.TanggalJatuhTempo)
		daysStr := fmt.Sprintf("%dd lagi", days)
		if days <= 3 && s.Status == "aktif" {
			daysStr = Red + Bold + daysStr + Reset
		} else if days <= 7 && s.Status == "aktif" {
			daysStr = Yellow + daysStr + Reset
		}
		fmt.Printf("%-4d %-22s %-14s %-12s %-6d %-10s %s%-8s%s\n",
			s.ID,
			truncate(s.Nama, 22),
			truncate(s.Kategori, 14),
			formatRupiah(s.BiayaBulanan),
			s.TanggalJatuhTempo,
			truncate(s.MetodePembayaran, 10),
			statusColor, s.Status, Reset,
		)
		_ = daysStr
	}
	printLine("─", 82)
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-2] + ".."
	}
	return s
}

// ==================== KATEGORI ====================

var kategoriList = []string{
	"Hiburan", "Musik", "Film & TV", "Game", "Produktivitas",
	"Kesehatan", "Pendidikan", "Cloud Storage", "Berita", "Lainnya",
}

var metodePembayaranList = []string{
	"Kartu Kredit", "Kartu Debit", "Transfer Bank", "OVO", "GoPay", "Dana", "ShopeePay", "Lainnya",
}

func pilihKategori() string {
	fmt.Println("\n  " + Bold + "Pilih Kategori:" + Reset)
	for i, k := range kategoriList {
		fmt.Printf("  %d. %s\n", i+1, k)
	}
	idx := inputInt("  Pilihan: ", 1, len(kategoriList))
	return kategoriList[idx-1]
}

func pilihMetodePembayaran() string {
	fmt.Println("\n  " + Bold + "Pilih Metode Pembayaran:" + Reset)
	for i, m := range metodePembayaranList {
		fmt.Printf("  %d. %s\n", i+1, m)
	}
	idx := inputInt("  Pilihan: ", 1, len(metodePembayaranList))
	return metodePembayaranList[idx-1]
}

// ==================== FITUR CRUD ====================

func tambahLangganan() {
	printHeader("📌 TAMBAH LANGGANAN BARU")
	fmt.Println()

	nama := input("  Nama Layanan   : ")
	if nama == "" {
		fmt.Println(Red + "  Nama tidak boleh kosong!" + Reset)
		tekanEnterLanjut()
		return
	}
	kategori := pilihKategori()
	biaya := inputFloat("  Biaya Bulanan  : Rp ")
	tanggalJT := inputInt("  Tanggal Jatuh Tempo (1-28): ", 1, 28)
	metode := pilihMetodePembayaran()
	catatan := input("  Catatan        : ")

	s := Subscription{
		ID:               db.NextID,
		Nama:             nama,
		Kategori:         kategori,
		BiayaBulanan:     biaya,
		MetodePembayaran: metode,
		TanggalJatuhTempo: tanggalJT,
		Status:           "aktif",
		TanggalMulai:     time.Now().Format("02-01-2006"),
		Catatan:          catatan,
	}

	db.Subscriptions = append(db.Subscriptions, s)
	db.NextID++
	saveDatabase()

	fmt.Println(Green + "\n  ✅ Langganan berhasil ditambahkan!" + Reset)
	tekanEnterLanjut()
}

func lihatSemuaLangganan() {
	printHeader("📋 DAFTAR SEMUA LANGGANAN")
	printSubscriptionTable(db.Subscriptions)

	// Hitung total aktif
	total := 0.0
	for _, s := range db.Subscriptions {
		if s.Status == "aktif" {
			total += s.BiayaBulanan
		}
	}
	fmt.Printf("\n  %sTotal Pengeluaran Aktif / Bulan: %s%s\n", Bold, formatRupiah(total), Reset)
	tekanEnterLanjut()
}

func ubahLangganan() {
	printHeader("✏️  UBAH LANGGANAN")
	printSubscriptionTable(db.Subscriptions)

	if len(db.Subscriptions) == 0 {
		tekanEnterLanjut()
		return
	}

	idStr := input("\n  Masukkan ID yang ingin diubah: ")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(Red + "  ID tidak valid!" + Reset)
		tekanEnterLanjut()
		return
	}

	s := findByID(id)
	if s == nil {
		fmt.Println(Red + "  ID tidak ditemukan!" + Reset)
		tekanEnterLanjut()
		return
	}

	fmt.Printf("\n  Data saat ini: %s%s%s\n", Bold, s.Nama, Reset)
	fmt.Println("  (Tekan ENTER untuk tidak mengubah)")

	nama := input(fmt.Sprintf("  Nama [%s]: ", s.Nama))
	if nama != "" {
		s.Nama = nama
	}

	fmt.Printf("  Kategori saat ini: %s\n", s.Kategori)
	gantiKat := input("  Ubah kategori? (y/n): ")
	if strings.ToLower(gantiKat) == "y" {
		s.Kategori = pilihKategori()
	}

	biayaStr := input(fmt.Sprintf("  Biaya [%s]: Rp ", formatRupiah(s.BiayaBulanan)))
	if biayaStr != "" {
		biaya, err := strconv.ParseFloat(biayaStr, 64)
		if err == nil && biaya >= 0 {
			s.BiayaBulanan = biaya
		}
	}

	tglStr := input(fmt.Sprintf("  Tanggal JT [%d]: ", s.TanggalJatuhTempo))
	if tglStr != "" {
		tgl, err := strconv.Atoi(tglStr)
		if err == nil && tgl >= 1 && tgl <= 28 {
			s.TanggalJatuhTempo = tgl
		}
	}

	fmt.Printf("  Metode saat ini: %s\n", s.MetodePembayaran)
	gantiMetode := input("  Ubah metode pembayaran? (y/n): ")
	if strings.ToLower(gantiMetode) == "y" {
		s.MetodePembayaran = pilihMetodePembayaran()
	}

	fmt.Println("\n  Status:")
	fmt.Println("  1. Aktif")
	fmt.Println("  2. Nonaktif")
	stStr := input("  Pilih [ENTER=skip]: ")
	if stStr == "1" {
		s.Status = "aktif"
	} else if stStr == "2" {
		s.Status = "nonaktif"
	}

	catatan := input(fmt.Sprintf("  Catatan [%s]: ", s.Catatan))
	if catatan != "" {
		s.Catatan = catatan
	}

	saveDatabase()
	fmt.Println(Green + "\n  ✅ Langganan berhasil diubah!" + Reset)
	tekanEnterLanjut()
}

func hapusLangganan() {
	printHeader("🗑️  HAPUS LANGGANAN")
	printSubscriptionTable(db.Subscriptions)

	if len(db.Subscriptions) == 0 {
		tekanEnterLanjut()
		return
	}

	idStr := input("\n  Masukkan ID yang ingin dihapus: ")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(Red + "  ID tidak valid!" + Reset)
		tekanEnterLanjut()
		return
	}

	s := findByID(id)
	if s == nil {
		fmt.Println(Red + "  ID tidak ditemukan!" + Reset)
		tekanEnterLanjut()
		return
	}

	konfirmasi := input(fmt.Sprintf("  Hapus \"%s\"? (y/n): ", s.Nama))
	if strings.ToLower(konfirmasi) != "y" {
		fmt.Println(Yellow + "  Dibatalkan." + Reset)
		tekanEnterLanjut()
		return
	}

	newSubs := []Subscription{}
	for _, sub := range db.Subscriptions {
		if sub.ID != id {
			newSubs = append(newSubs, sub)
		}
	}
	db.Subscriptions = newSubs
	saveDatabase()

	fmt.Println(Green + "\n  ✅ Langganan berhasil dihapus!" + Reset)
	tekanEnterLanjut()
}

// ==================== PENGINGAT ====================

func tampilkanPengingat() {
	printHeader("🔔 PENGINGAT JATUH TEMPO")
	fmt.Println()

	aktif := getActiveSubscriptions()
	ada := false

	// Urutkan berdasarkan hari terdekat
	sort.Slice(aktif, func(i, j int) bool {
		return daysUntilDue(aktif[i].TanggalJatuhTempo) < daysUntilDue(aktif[j].TanggalJatuhTempo)
	})

	fmt.Printf("  %-22s %-12s %-10s %s\n", "Nama", "Biaya", "Tgl JT", "Sisa Hari")
	printLine("─", 60)

	for _, s := range aktif {
		days := daysUntilDue(s.TanggalJatuhTempo)
		var ket string
		var color string

		if days == 0 {
			ket = "⚠️  HARI INI!"
			color = Red + Bold
		} else if days <= 3 {
			ket = fmt.Sprintf("🔴 %d hari lagi", days)
			color = Red
		} else if days <= 7 {
			ket = fmt.Sprintf("🟡 %d hari lagi", days)
			color = Yellow
		} else {
			ket = fmt.Sprintf("🟢 %d hari lagi", days)
			color = Green
		}

		if days <= 7 {
			ada = true
		}

		fmt.Printf("  %-22s %-12s %-10d %s%s%s\n",
			truncate(s.Nama, 22),
			formatRupiah(s.BiayaBulanan),
			s.TanggalJatuhTempo,
			color, ket, Reset,
		)
	}

	if !ada {
		fmt.Println("\n  " + Green + "✅ Tidak ada jatuh tempo dalam 7 hari ke depan." + Reset)
	}

	tekanEnterLanjut()
}

// ==================== PENCARIAN ====================

// Sequential Search — cari berdasarkan nama (case-insensitive)
func sequentialSearch(keyword string) []Subscription {
	keyword = strings.ToLower(keyword)
	var hasil []Subscription
	for _, s := range db.Subscriptions {
		if strings.Contains(strings.ToLower(s.Nama), keyword) ||
			strings.Contains(strings.ToLower(s.Catatan), keyword) {
			hasil = append(hasil, s)
		}
	}
	return hasil
}

// Binary Search — cari berdasarkan kategori (data harus terurut)
func binarySearchKategori(kategori string) []Subscription {
	// Salin & urutkan berdasarkan kategori
	sorted := make([]Subscription, len(db.Subscriptions))
	copy(sorted, db.Subscriptions)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Kategori) < strings.ToLower(sorted[j].Kategori)
	})

	kategori = strings.ToLower(kategori)
	lo, hi := 0, len(sorted)-1
	var hasil []Subscription

	// Temukan salah satu index yang cocok
	foundIdx := -1
	for lo <= hi {
		mid := (lo + hi) / 2
		midKat := strings.ToLower(sorted[mid].Kategori)
		if midKat == kategori {
			foundIdx = mid
			break
		} else if midKat < kategori {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	if foundIdx == -1 {
		return hasil
	}

	// Kumpulkan semua yang sama kategorinya di sekitar foundIdx
	left := foundIdx
	for left > 0 && strings.ToLower(sorted[left-1].Kategori) == kategori {
		left--
	}
	right := foundIdx
	for right < len(sorted)-1 && strings.ToLower(sorted[right+1].Kategori) == kategori {
		right++
	}
	return sorted[left : right+1]
}

func menuPencarian() {
	printHeader("🔍 PENCARIAN LANGGANAN")
	fmt.Println()
	fmt.Println("  1. Pencarian berdasarkan Nama (Sequential Search)")
	fmt.Println("  2. Pencarian berdasarkan Kategori (Binary Search)")
	fmt.Println("  0. Kembali")
	fmt.Println()

	pilihan := input("  Pilihan: ")
	switch pilihan {
	case "1":
		keyword := input("\n  Masukkan kata kunci nama: ")
		hasil := sequentialSearch(keyword)
		fmt.Printf("\n  %sHasil Sequential Search untuk \"%s\":%s\n", Bold, keyword, Reset)
		fmt.Printf("  Algoritma: memeriksa satu per satu dari awal hingga akhir data\n")
		fmt.Printf("  Ditemukan: %d hasil\n", len(hasil))
		printSubscriptionTable(hasil)

	case "2":
		fmt.Println("\n  Kategori tersedia:", strings.Join(kategoriList, ", "))
		keyword := input("  Masukkan kategori: ")
		hasil := binarySearchKategori(keyword)
		fmt.Printf("\n  %sHasil Binary Search untuk kategori \"%s\":%s\n", Bold, keyword, Reset)
		fmt.Printf("  Algoritma: data diurutkan, lalu dicari di tengah secara rekursif\n")
		fmt.Printf("  Ditemukan: %d hasil\n", len(hasil))
		printSubscriptionTable(hasil)
	}

	tekanEnterLanjut()
}

// ==================== PENGURUTAN ====================

// Selection Sort — urutkan berdasarkan biaya (descending)
func selectionSortBiaya(subs []Subscription) []Subscription {
	arr := make([]Subscription, len(subs))
	copy(arr, subs)
	n := len(arr)
	for i := 0; i < n-1; i++ {
		maxIdx := i
		for j := i + 1; j < n; j++ {
			if arr[j].BiayaBulanan > arr[maxIdx].BiayaBulanan {
				maxIdx = j
			}
		}
		arr[i], arr[maxIdx] = arr[maxIdx], arr[i]
	}
	return arr
}

// Insertion Sort — urutkan berdasarkan tanggal jatuh tempo (ascending)
func insertionSortTanggal(subs []Subscription) []Subscription {
	arr := make([]Subscription, len(subs))
	copy(arr, subs)
	n := len(arr)
	for i := 1; i < n; i++ {
		key := arr[i]
		j := i - 1
		for j >= 0 && arr[j].TanggalJatuhTempo > key.TanggalJatuhTempo {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
	return arr
}

func menuPengurutan() {
	printHeader("📊 PENGURUTAN LANGGANAN")
	fmt.Println()
	fmt.Println("  1. Urutkan berdasarkan Biaya Terbesar (Selection Sort)")
	fmt.Println("  2. Urutkan berdasarkan Tanggal Jatuh Tempo Terdekat (Insertion Sort)")
	fmt.Println("  0. Kembali")
	fmt.Println()

	pilihan := input("  Pilihan: ")
	switch pilihan {
	case "1":
		sorted := selectionSortBiaya(db.Subscriptions)
		fmt.Printf("\n  %sDiurutkan berdasarkan Biaya Terbesar (Selection Sort):%s\n", Bold, Reset)
		fmt.Println("  Algoritma: cari maksimum, tukar ke posisi depan, ulangi")
		printSubscriptionTable(sorted)

	case "2":
		sorted := insertionSortTanggal(db.Subscriptions)
		fmt.Printf("\n  %sDiurutkan berdasarkan Tanggal Jatuh Tempo (Insertion Sort):%s\n", Bold, Reset)
		fmt.Println("  Algoritma: ambil elemen, sisipkan ke posisi yang tepat secara bertahap")
		printSubscriptionTable(sorted)
	}

	tekanEnterLanjut()
}

// ==================== LAPORAN & REKOMENDASI ====================

func laporanBulanan() {
	printHeader("💰 LAPORAN PENGELUARAN BULANAN")
	fmt.Println()

	aktif := getActiveSubscriptions()
	if len(aktif) == 0 {
		fmt.Println(Yellow + "  Tidak ada langganan aktif." + Reset)
		tekanEnterLanjut()
		return
	}

	// Total per kategori
	perKategori := make(map[string]float64)
	totalBulanan := 0.0
	for _, s := range aktif {
		perKategori[s.Kategori] += s.BiayaBulanan
		totalBulanan += s.BiayaBulanan
	}

	// Urutkan kategori berdasarkan biaya
	type katBiaya struct {
		Kategori string
		Biaya    float64
	}
	var katList []katBiaya
	for k, v := range perKategori {
		katList = append(katList, katBiaya{k, v})
	}
	sort.Slice(katList, func(i, j int) bool {
		return katList[i].Biaya > katList[j].Biaya
	})

	fmt.Printf("  %s%-20s %-15s %s%s\n", Bold+Cyan, "Kategori", "Total/Bulan", "Proporsi", Reset)
	printLine("─", 50)
	for _, kb := range katList {
		persen := (kb.Biaya / totalBulanan) * 100
		bar := strings.Repeat("█", int(persen/5))
		fmt.Printf("  %-20s %-15s %.1f%% %s%s%s\n",
			kb.Kategori, formatRupiah(kb.Biaya), persen, Green, bar, Reset)
	}
	printLine("─", 50)
	fmt.Printf("  %s%-20s %s%s\n", Bold, "TOTAL BULANAN", formatRupiah(totalBulanan), Reset)
	fmt.Printf("  %-20s %s\n", "Total Tahunan (est.)", formatRupiah(totalBulanan*12))

	// Rekomendasi
	fmt.Printf("\n%s%s  💡 REKOMENDASI PENGHEMATAN  %s\n", Bold, Yellow, Reset)
	printLine("─", 50)

	// Urutkan aktif berdasarkan biaya (descending) - pakai selection sort
	sorted := selectionSortBiaya(aktif)

	threshold := totalBulanan * 0.20 // >20% dari total = perlu dipertimbangkan

	fmt.Println("\n  Langganan dengan biaya tertinggi:")
	for i, s := range sorted {
		if i >= 3 {
			break
		}
		persen := (s.BiayaBulanan / totalBulanan) * 100
		tag := ""
		if s.BiayaBulanan >= threshold {
			tag = Red + " ⚠️ Pertimbangkan untuk dievaluasi" + Reset
		}
		fmt.Printf("  %d. %-20s %s (%.1f%%)%s\n",
			i+1, s.Nama, formatRupiah(s.BiayaBulanan), persen, tag)
	}

	if totalBulanan > 500000 {
		fmt.Printf("\n  %s💡 Total pengeluaran Anda %s/bulan cukup besar.%s\n",
			Yellow, formatRupiah(totalBulanan), Reset)
		fmt.Println("  Pertimbangkan untuk:")
		fmt.Println("   • Berbagi akun (family plan) untuk layanan populer")
		fmt.Println("   • Berlangganan tahunan (biasanya lebih hemat ~20-30%)")
		fmt.Println("   • Nonaktifkan langganan yang jarang digunakan")

		// Simulasi hemat
		potensialHemat := sorted[0].BiayaBulanan
		fmt.Printf("\n  %sPotensial hemat jika hentikan \"%s\": %s/bulan (%s/tahun)%s\n",
			Green, sorted[0].Nama, formatRupiah(potensialHemat), formatRupiah(potensialHemat*12), Reset)
	} else {
		fmt.Println("\n  " + Green + "✅ Pengeluaran langganan Anda masih dalam batas wajar." + Reset)
	}

	tekanEnterLanjut()
}

// ==================== MENU UTAMA ====================

func tampilkanMenuUtama() {
	clearScreen()

	now := time.Now()

	// Hitung statistik cepat
	aktif := getActiveSubscriptions()
	total := 0.0
	for _, s := range aktif {
		total += s.BiayaBulanan
	}

	// Cek pengingat mendesak
	urgent := 0
	for _, s := range aktif {
		if daysUntilDue(s.TanggalJatuhTempo) <= 3 {
			urgent++
		}
	}

	printLine("═", 60)
	fmt.Printf("%s%s%s  %-54s%s\n", Bold, BgBlue, White, "  💳 SUBSCRIPTION TRACKER", Reset)
	fmt.Printf("  %-56s\n", "  Kelola Langganan Digital Anda")
	printLine("═", 60)
	fmt.Printf("  📅 %s\n", now.Format("Monday, 02 January 2006"))
	fmt.Printf("  📊 Langganan Aktif  : %s%d layanan%s\n", Bold, len(aktif), Reset)
	fmt.Printf("  💵 Total/Bulan      : %s%s%s\n", Bold+Green, formatRupiah(total), Reset)
	if urgent > 0 {
		fmt.Printf("  %s⚠️  %d langganan jatuh tempo dalam 3 hari!%s\n", Red+Bold, urgent, Reset)
	}
	printLine("─", 60)
	fmt.Println()
	fmt.Println("  " + Bold + "KELOLA LANGGANAN" + Reset)
	fmt.Println("  1. Lihat Semua Langganan")
	fmt.Println("  2. Tambah Langganan Baru")
	fmt.Println("  3. Ubah Langganan")
	fmt.Println("  4. Hapus Langganan")
	fmt.Println()
	fmt.Println("  " + Bold + "FITUR LANJUTAN" + Reset)
	fmt.Println("  5. 🔔 Pengingat Jatuh Tempo")
	fmt.Println("  6. 🔍 Pencarian (Sequential & Binary Search)")
	fmt.Println("  7. 📊 Pengurutan (Selection & Insertion Sort)")
	fmt.Println("  8. 💰 Laporan & Rekomendasi Penghematan")
	fmt.Println()
	fmt.Println("  0. Keluar")
	fmt.Println()
	printLine("─", 60)
}

func main() {
	loadDatabase()

	// Seed data default jika kosong
	if len(db.Subscriptions) == 0 {
		seedData()
	}

	for {
		tampilkanMenuUtama()
		pilihan := input("  Pilih menu: ")

		switch pilihan {
		case "1":
			lihatSemuaLangganan()
		case "2":
			tambahLangganan()
		case "3":
			ubahLangganan()
		case "4":
			hapusLangganan()
		case "5":
			tampilkanPengingat()
		case "6":
			menuPencarian()
		case "7":
			menuPengurutan()
		case "8":
			laporanBulanan()
		case "0":
			clearScreen()
			fmt.Println(Green + "\n  Terima kasih telah menggunakan Subscription Tracker!\n" + Reset)
			os.Exit(0)
		default:
			fmt.Println(Red + "  Pilihan tidak valid!" + Reset)
			time.Sleep(800 * time.Millisecond)
		}
	}
}

// ==================== SEED DATA ====================

func seedData() {
	contoh := []Subscription{
		{ID: 1, Nama: "Netflix", Kategori: "Film & TV", BiayaBulanan: 186000, MetodePembayaran: "Kartu Kredit", TanggalJatuhTempo: 5, Status: "aktif", TanggalMulai: "01-01-2024", Catatan: "Paket Standard"},
		{ID: 2, Nama: "Spotify Premium", Kategori: "Musik", BiayaBulanan: 54990, MetodePembayaran: "GoPay", TanggalJatuhTempo: 10, Status: "aktif", TanggalMulai: "01-03-2024", Catatan: "Individual"},
		{ID: 3, Nama: "Disney+ Hotstar", Kategori: "Film & TV", BiayaBulanan: 49000, MetodePembayaran: "Kartu Debit", TanggalJatuhTempo: 15, Status: "aktif", TanggalMulai: "01-06-2024", Catatan: ""},
		{ID: 4, Nama: "YouTube Premium", Kategori: "Hiburan", BiayaBulanan: 59000, MetodePembayaran: "OVO", TanggalJatuhTempo: 20, Status: "aktif", TanggalMulai: "01-09-2023", Catatan: ""},
		{ID: 5, Nama: "Google One", Kategori: "Cloud Storage", BiayaBulanan: 35000, MetodePembayaran: "Kartu Kredit", TanggalJatuhTempo: 25, Status: "aktif", TanggalMulai: "01-01-2023", Catatan: "100GB"},
		{ID: 6, Nama: "Canva Pro", Kategori: "Produktivitas", BiayaBulanan: 120000, MetodePembayaran: "Kartu Kredit", TanggalJatuhTempo: 8, Status: "aktif", TanggalMulai: "01-04-2024", Catatan: ""},
		{ID: 7, Nama: "Duolingo Plus", Kategori: "Pendidikan", BiayaBulanan: 89000, MetodePembayaran: "Dana", TanggalJatuhTempo: 12, Status: "nonaktif", TanggalMulai: "01-07-2023", Catatan: "Jarang dipakai"},
		{ID: 8, Nama: "Xbox Game Pass", Kategori: "Game", BiayaBulanan: 149000, MetodePembayaran: "Kartu Kredit", TanggalJatuhTempo: 3, Status: "aktif", TanggalMulai: "01-02-2024", Catatan: "Ultimate"},
	}
	db.Subscriptions = contoh
	db.NextID = 9
	saveDatabase()
}