package models

type Loan struct {
	LocationID                    string                    `json:"locationid"`
	RecDibuatLokasi               string                    `json:"rec_dibuat_lokasi"`
	RestrukturNoAkadAkhir         string                    `json:"restruktur_noakad_akhir"`
	RestrukturTanggalAkhirAkad    DateTZ                    `json:"restruktur_tanggalakhirakad"`
	RestrukturTanggalAwal         any                       `json:"restruktur_tanggalawal"`
	RestrukturTanggalAkhir        any                       `json:"restruktur_tanggalakhir"`
	RestrukturCara                any                       `json:"restruktur_cara"`
	RestrukturFrekuensi           string                    `json:"restruktur_frekuensi"`
	TanggalMulaiTunggakan         any                       `json:"tanggalmulaitunggakan"`
	Lokasi                        string                    `json:"lokasi"`
	ID                            string                    `json:"id"`
	NoPk                          string                    `json:"nopk"`
	NamaNasabah                   string                    `json:"namanasabah"`
	AliasNama                     string                    `json:"aliasnama"`
	StatusRekening                string                    `json:"statusrekening"`
	TglPencairan                  string                    `json:"tgl_pencairan"`
	RecDibuatOleh                 string                    `json:"rec_dibuat_oleh"`
	NoAlt                         string                    `json:"noalt"`
	ProdukJenisPinjaman           string                    `json:"produk_jenispinjaman"`
	ProdukID                      string                    `json:"produkid"`
	IDProduk                      string                    `json:"idproduk"`
	Currency                      string                    `json:"currency"`
	PlafondLimit                  string                    `json:"plafondlimit"`
	JmlPokokPinjaman              string                    `json:"jmlpokok_pinjaman"`
	JangkaWaktu                   string                    `json:"jangkawaktu"`
	ProdukJenisAngsuran           string                    `json:"produk_jenisangsuran"`
	TglAngsuran                   int64                     `json:"tgl_angsuran"`
	SidSifatKredit2               string                    `json:"sid_sifatkredit2"`
	SidJenisPenggunaan            string                    `json:"sid_jenispenggunaan"`
	SidSumberDanaPelunasan        string                    `json:"sid_sumberdanapelunasan"`
	SidGolonganKredit             string                    `json:"sid_golongankredit"`
	SidOrientasiPenggunaan        string                    `json:"sid_orientasipenggunaan"`
	SidSektorEkonomi              string                    `json:"sid_sektorekonomi"`
	PejabatKredit                 any                       `json:"pejabatkredit"`
	PejabatKreditDua              any                       `json:"pejabatkreditdua"`
	SidSektorEkonomi2             string                    `json:"sid_sektorekonomi2"`
	SidSifatKredit                string                    `json:"sid_sifatkredit"`
	DataPenjamin                  string                    `json:"datapenjamin"`
	MengetahuiSuamiIstri          string                    `json:"mengetahuisuamiistri"`
	DPNama                        any                       `json:"dp_nama"`
	DPNoKtp                       any                       `json:"dp_noktp"`
	DPAlamat                      any                       `json:"dp_alamat"`
	DPTempatLahir                 any                       `json:"dp_tempatlahir"`
	DPTglLahir                    any                       `json:"dp_tgllahir"`
	DPJenisKelamin                any                       `json:"dp_jeniskelamin"`
	DPGolDarah                    any                       `json:"dp_goldarah"`
	DPAgama                       any                       `json:"dp_agama"`
	DPStatusPerkawinan            any                       `json:"dp_statusperkawinan"`
	DPPekerjaan                   any                       `json:"dp_pekerjaan"`
	DPNoCif                       any                       `json:"dp_nocif"`
	DPHubungan                    any                       `json:"dp_hubungan"`
	DPKewarganegaraan             any                       `json:"dp_kewarganegaraan"`
	DPTempatTerbit                any                       `json:"dp_tempatterbit"`
	DPTglTerbit                   any                       `json:"dp_tglterbit"`
	DPBerlakuSeumurHidup          any                       `json:"dp_berlakuseumurhidup"`
	DPTglBerlakuSampai            any                       `json:"dp_tglberlakusampai"`
	SidJenisUsaha                 string                    `json:"sid_jenisusaha"`
	BungaFlat                     float64                   `json:"bungaflat"`
	NoPerjanjianKredit            string                    `json:"noperjanjiankredit"`
	JournalID                     string                    `json:"journalid"`
	PersenDendaTunggakan          int64                     `json:"persendendatunggakan"`
	BungaBerjenjang               string                    `json:"bungaberjenjang"`
	NoRekGabunganBnpl             string                    `json:"norekgabungan_bnpl"`
	TglTutup                      any                       `json:"tgl_tutup"`
	Titipan                       int64                     `json:"titipan"`
	Periode                       string                    `json:"periode"`
	ProdukSukuBunga               float64                   `json:"produk_sukubunga"`
	ProdukPerubahanSukuBunga      string                    `json:"produk_perubahansukubunga"`
	TujuanKredit                  any                       `json:"tujuankredit"`
	TglTerakhirBayarPokokDanBunga DateTZ                    `json:"tglterakhir_bayarpokokdanbunga"`
	TglBayarPokokBungaBerikutnya  DateTZ                    `json:"tglbayarpokokbunga_berikutnya"`
	TglJtTerakhir                 DateTZ                    `json:"tgljtterakhir"`
	TglJtBerikutnya               DateTZ                    `json:"tgljtberikutnya"`
	OutstandingPinjaman           string                    `json:"outstandingpinjaman"`
	TunggakanPokok                string                    `json:"tunggakanpokok"`
	Accrue                        string                    `json:"accrue"`
	DecimalPoint                  int64                     `json:"decimalpoint"`
	TunggakanBunga                string                    `json:"tunggakanbunga"`
	DendaTunggakan                string                    `json:"dendatunggakan"`
	Dpd                           int64                     `json:"dpd"`
	KolekBi                       int64                     `json:"kolekbi"`
	KolekBpr                      int64                     `json:"kolekbpr"`
	UpdateKolekBi                 string                    `json:"updatekolekbi"`
	TotalCollateralValue          int64                     `json:"totalcollateralvalue"`
	TotalAssetValue               int64                     `json:"totalassetvalue"`
	NoCif                         string                    `json:"nocif"`
	JenisNasabah                  string                    `json:"jenisnasabah"`
	TglBukaCif                    DateTZ                    `json:"tglbukacif"`
	StatusDokumen                 string                    `json:"status_dokumen"`
	DataAlamatKtpAlamat1          string                    `json:"dataalamat_ktp_alamat1"`
	DataAlamatKtpAlamat2          any                       `json:"dataalamat_ktp_alamat2"`
	DataAlamatKtpRt               string                    `json:"dataalamat_ktp_rt"`
	DataAlamatKtpRw               string                    `json:"dataalamat_ktp_rw"`
	DataAlamatKtpKelurahan        string                    `json:"dataalamat_ktp_kelurahan"`
	DataAlamatKtpKecamatan        string                    `json:"dataalamat_ktp_kecamatan"`
	DataAlamatKtpKota             string                    `json:"dataalamat_ktp_kota"`
	DataAlamatKtpPropinsi         string                    `json:"dataalamat_ktp_propinsi"`
	DataAlamatKtpKodePos          string                    `json:"dataalamat_ktp_kodepos"`
	DataAlamatRumahNoHp           string                    `json:"dataalamat_rumah_nohp"`
	NoRekTabPencairanPinjaman     string                    `json:"norektab_pencairanpinjaman"`
	NoRekTabBayarAngsuran         string                    `json:"norektab_bayarangsuran"`
	NoRekTabPencairanPinjaman2    string                    `json:"norektab_pencairanpinjaman2"`
	NoRekTabBayarAngsuran2        string                    `json:"norektab_bayarangsuran2"`
	JenisJaminan2                 string                    `json:"jenisjaminan2"`
	JenisJaminan                  string                    `json:"jenisjaminan"`
	TotalNilaiPasar               string                    `json:"totalnilaipasar"`
	Terpakai                      int64                     `json:"terpakai"`
	TotalNilaiJaminan             int64                     `json:"totalnilaijaminan"`
	Jaminan                       string                    `json:"jaminan"`
	JmlAgunan                     int64                     `json:"jmlagunan"`
	TglHapusBuku                  any                       `json:"tglhapusbuku"`
	TotalHapusBuku                string                    `json:"totalhapusbuku"`
	NilaiHapusBukuSaldoPinjaman   any                       `json:"nilaihapusbuku_saldopinjaman"`
	NilaiHapusBukuBungaBerjalan   any                       `json:"nilaihapusbuku_bungaberjalan"`
	NilaiHapusBukuTunggakanBunga  any                       `json:"nilaihapusbuku_tunggakanbunga"`
	NilaiHapusBukuTunggakanDenda  any                       `json:"nilaihapusbuku_tunggakandenda"`
	PpapBlnTerakhir               string                    `json:"ppapblnterakhir"`
	PpapTglTerakhir               string                    `json:"ppaptglterakhir"`
	Marketing                     string                    `json:"marketing"`
	DataCsNotes                   any                       `json:"datacs_notes"`
	AnalisKredit                  any                       `json:"analiskredit"`
	AnalisKreditNotes             any                       `json:"analiskredit_notes"`
	HTPokok                       string                    `json:"ht_pokok"`
	HTBunga                       string                    `json:"ht_bunga"`
	TglHapusTagih                 any                       `json:"tglhapustagih"`
	TotalHT                       any                       `json:"total_ht"`
	JmlPokokPinjaman2             int64                     `json:"jmlpokok_pinjaman2"`
	TotalNilaiPasar2              int64                     `json:"totalnilaipasar2"`
	ServiceID                     any                       `json:"serviceid"`
	GroupID                       any                       `json:"groupid"`
	DataTanahBangunan             []DatatanahBangunan       `json:"datatanah_bangunan"`
	JenisJaminanTanah             string                    `json:"jenisjaminantanah"`
	JenisJaminanLainnya           string                    `json:"jenisjaminanlainnya"`
	BiayaPencairan                []Biayapencairan          `json:"biayapencairan"`
	Tabungan                      []Tabungan                `json:"tabungan"`
	JadwalAngsuran                []Jadwalangsuran          `json:"jadwalangsuran"`
	HistoryBayar                  []Historybayar            `json:"historybayar"`
	TempatPenyimpanan             bool                      `json:"tempatpenyimpanan"`
	RekeningBungaBerjenjang       []RekeningBungaberjenjang `json:"rekening_bungaberjenjang"`
	Channeling                    bool                      `json:"channeling"`
	AsuransiData                  string                    `json:"asuransidata"`
}

type Biayapencairan struct {
	NamaBiaya   string `json:"namabiaya"`
	JumlahBiaya string `json:"jumlah_biaya"`
	HitungDwp   string `json:"hitungdwp"`
}

type DatatanahBangunan struct {
	TanahNoSertifikat          string `json:"tanah_nosertifikat"`
	TanahStatusSertifikat      string `json:"tanah_statussertifikat"`
	TanahImb                   any    `json:"tanah_imb"`
	TanahNoImb                 any    `json:"tanah_noimb"`
	TanahLokasi                any    `json:"tanah_lokasi"`
	TanahRt                    any    `json:"tanah_rt"`
	TanahRw                    any    `json:"tanah_rw"`
	TanahKelurahan             any    `json:"tanah_kelurahan"`
	TanahKecamatan             any    `json:"tanah_kecamatan"`
	TanahKota                  any    `json:"tanah_kota"`
	TanahProvinsi              any    `json:"tanah_provinsi"`
	TanahLuas                  any    `json:"tanah_luas"`
	TanahSuratTanahDikeluarkan string `json:"tanah_surattanahdikeluarkan"`
	TanahKodePos               any    `json:"tanah_kodepos"`
	TanahAtasNama              string `json:"tanah_atasnama"`
	TanahNilaiJaminan          int64  `json:"tanah_nilaijaminan"`
	TanahNilaiPasar            int64  `json:"tanah_nilaipasar"`
	SidJenisPengikatan         string `json:"sid_jenispengikatan"`
	SidDatiDua                 string `json:"sid_datidua"`
	Asuransi                   string `json:"ansuransi"`
	TglPenilaian               DateTZ `json:"tglpenilaian"`
	TanahNilaiJaminanReal      int64  `json:"tanah_nilaijaminanreal"`
	PersenPpap                 string `json:"persenppap"`
	TanahTglSertifikat         any    `json:"tanah_tglsertifikat"`
	TanahNoSuratUkur           any    `json:"tanah_nosuratukur"`
	TanahTglUkur               any    `json:"tanah_tglukur"`
}

type Historybayar struct {
	Tgl                 DateTZ `json:"tgl"`
	AngsuranKe          int64  `json:"angsuranke"`
	TglBayar            string `json:"tglbayar"`
	Currency            string `json:"currency"`
	Tgljt               string `json:"tgljt"`
	TotalBayar          string `json:"totalbayar"`
	BayarPokok          string `json:"bayar_pokok"`
	BayarBunga          string `json:"bayar_bunga"`
	BayarDenda          string `json:"bayar_denda"`
	BayarDendaPelunasan string `json:"bayar_dendapelunasan"`
	NominalDwp          string `json:"nominaldwp"`
	NoJurnal            string `json:"nojurnal"`
	Cabang              string `json:"cabang"`
	Keterangan          any    `json:"keterangan"`
	Officer             string `json:"officer"`
	Otor                string `json:"otor"`
}

type Jadwalangsuran struct {
	Tanggal      string `json:"tanggal"`
	Angsuran     string `json:"angsuran"`
	Bunga        string `json:"bunga"`
	Pokok        string `json:"pokok"`
	Denda        string `json:"denda"`
	BayarPokok   string `json:"bayar_pokok"`
	BayarDenda   string `json:"bayar_denda"`
	BayarBunga   string `json:"bayar_bunga"`
	SisaPinjaman string `json:"sisapinjaman"`
	StatusBayar  string `json:"statusbayar"`
	AngsuranKe   int64  `json:"angsuranke"`
}

type RekeningBungaberjenjang struct {
	ID           string       `json:"id"`
	NoUrut       string       `json:"nourut"`
	NoRekening   string       `json:"norekening"`
	Tenor        int64        `json:"tenor"`
	Bunga        float64      `json:"bunga"`
	RecTimestamp RecTimestamp `json:"rec_timestamp"`
}

type RecTimestamp struct {
	Date string `json:"date"`
	Type string `json:"type"`
}

type Tabungan struct {
	ID              string  `json:"id"`
	NamaNasabah     string  `json:"namanasabah"`
	ProdukID        string  `json:"produkid"`
	TglBukaRekening *DateTZ `json:"tglbukarekening"`
	Currency        string  `json:"currency"`
	StatusDokumen   string  `json:"status_dokumen"`
	Saldo           string  `json:"saldo"`
	SaldoDebit      string  `json:"saldodebit"`
}
