package models

type Loan struct {
	Locationid                    string                    `json:"locationid"`
	RecDibuatLokasi               string                    `json:"rec_dibuat_lokasi"`
	RestrukturNoakadAkhir         string                    `json:"restruktur_noakad_akhir"`
	RestrukturTanggalakhirakad    DateTZ                    `json:"restruktur_tanggalakhirakad"`
	RestrukturTanggalawal         interface{}               `json:"restruktur_tanggalawal"`
	RestrukturTanggalakhir        interface{}               `json:"restruktur_tanggalakhir"`
	RestrukturCara                interface{}               `json:"restruktur_cara"`
	RestrukturFrekuensi           string                    `json:"restruktur_frekuensi"`
	Tanggalmulaitunggakan         interface{}               `json:"tanggalmulaitunggakan"`
	Lokasi                        string                    `json:"lokasi"`
	ID                            string                    `json:"id"`
	Nopk                          string                    `json:"nopk"`
	Namanasabah                   string                    `json:"namanasabah"`
	Aliasnama                     string                    `json:"aliasnama"`
	Statusrekening                string                    `json:"statusrekening"`
	TglPencairan                  string                    `json:"tgl_pencairan"`
	RecDibuatOleh                 string                    `json:"rec_dibuat_oleh"`
	Noalt                         string                    `json:"noalt"`
	ProdukJenispinjaman           string                    `json:"produk_jenispinjaman"`
	Produkid                      string                    `json:"produkid"`
	Idproduk                      string                    `json:"idproduk"`
	Currency                      string                    `json:"currency"`
	Plafondlimit                  string                    `json:"plafondlimit"`
	JmlpokokPinjaman              string                    `json:"jmlpokok_pinjaman"`
	Jangkawaktu                   string                    `json:"jangkawaktu"`
	ProdukJenisangsuran           string                    `json:"produk_jenisangsuran"`
	TglAngsuran                   int64                     `json:"tgl_angsuran"`
	SidSifatkredit2               string                    `json:"sid_sifatkredit2"`
	SidJenispenggunaan            string                    `json:"sid_jenispenggunaan"`
	SidSumberdanapelunasan        string                    `json:"sid_sumberdanapelunasan"`
	SidGolongankredit             string                    `json:"sid_golongankredit"`
	SidOrientasipenggunaan        string                    `json:"sid_orientasipenggunaan"`
	SidSektorekonomi              string                    `json:"sid_sektorekonomi"`
	Pejabatkredit                 interface{}               `json:"pejabatkredit"`
	Pejabatkreditdua              interface{}               `json:"pejabatkreditdua"`
	SidSektorekonomi2             string                    `json:"sid_sektorekonomi2"`
	SidSifatkredit                string                    `json:"sid_sifatkredit"`
	Datapenjamin                  string                    `json:"datapenjamin"`
	Mengetahuisuamiistri          string                    `json:"mengetahuisuamiistri"`
	DPNama                        interface{}               `json:"dp_nama"`
	DPNoktp                       interface{}               `json:"dp_noktp"`
	DPAlamat                      interface{}               `json:"dp_alamat"`
	DPTempatlahir                 interface{}               `json:"dp_tempatlahir"`
	DPTgllahir                    interface{}               `json:"dp_tgllahir"`
	DPJeniskelamin                interface{}               `json:"dp_jeniskelamin"`
	DPGoldarah                    interface{}               `json:"dp_goldarah"`
	DPAgama                       interface{}               `json:"dp_agama"`
	DPStatusperkawinan            interface{}               `json:"dp_statusperkawinan"`
	DPPekerjaan                   interface{}               `json:"dp_pekerjaan"`
	DPNocif                       interface{}               `json:"dp_nocif"`
	DPHubungan                    interface{}               `json:"dp_hubungan"`
	DPKewarganegaraan             interface{}               `json:"dp_kewarganegaraan"`
	DPTempatterbit                interface{}               `json:"dp_tempatterbit"`
	DPTglterbit                   interface{}               `json:"dp_tglterbit"`
	DPBerlakuseumurhidup          interface{}               `json:"dp_berlakuseumurhidup"`
	DPTglberlakusampai            interface{}               `json:"dp_tglberlakusampai"`
	SidJenisusaha                 string                    `json:"sid_jenisusaha"`
	Bungaflat                     float64                   `json:"bungaflat"`
	Noperjanjiankredit            string                    `json:"noperjanjiankredit"`
	Journalid                     string                    `json:"journalid"`
	Persendendatunggakan          int64                     `json:"persendendatunggakan"`
	Bungaberjenjang               string                    `json:"bungaberjenjang"`
	NorekgabunganBnpl             string                    `json:"norekgabungan_bnpl"`
	TglTutup                      interface{}               `json:"tgl_tutup"`
	Titipan                       int64                     `json:"titipan"`
	Periode                       string                    `json:"periode"`
	ProdukSukubunga               float64                   `json:"produk_sukubunga"`
	ProdukPerubahansukubunga      string                    `json:"produk_perubahansukubunga"`
	Tujuankredit                  interface{}               `json:"tujuankredit"`
	TglterakhirBayarpokokdanbunga DateTZ                    `json:"tglterakhir_bayarpokokdanbunga"`
	TglbayarpokokbungaBerikutnya  DateTZ                    `json:"tglbayarpokokbunga_berikutnya"`
	Tgljtterakhir                 DateTZ                    `json:"tgljtterakhir"`
	Tgljtberikutnya               DateTZ                    `json:"tgljtberikutnya"`
	Outstandingpinjaman           string                    `json:"outstandingpinjaman"`
	Tunggakanpokok                string                    `json:"tunggakanpokok"`
	Accrue                        string                    `json:"accrue"`
	Decimalpoint                  int64                     `json:"decimalpoint"`
	Tunggakanbunga                string                    `json:"tunggakanbunga"`
	Dendatunggakan                string                    `json:"dendatunggakan"`
	Dpd                           int64                     `json:"dpd"`
	Kolekbi                       int64                     `json:"kolekbi"`
	Kolekbpr                      int64                     `json:"kolekbpr"`
	Updatekolekbi                 string                    `json:"updatekolekbi"`
	Totalcollateralvalue          int64                     `json:"totalcollateralvalue"`
	Totalassetvalue               int64                     `json:"totalassetvalue"`
	Nocif                         string                    `json:"nocif"`
	Jenisnasabah                  string                    `json:"jenisnasabah"`
	Tglbukacif                    DateTZ                    `json:"tglbukacif"`
	StatusDokumen                 string                    `json:"status_dokumen"`
	DataalamatKtpAlamat1          string                    `json:"dataalamat_ktp_alamat1"`
	DataalamatKtpAlamat2          interface{}               `json:"dataalamat_ktp_alamat2"`
	DataalamatKtpRt               string                    `json:"dataalamat_ktp_rt"`
	DataalamatKtpRw               string                    `json:"dataalamat_ktp_rw"`
	DataalamatKtpKelurahan        string                    `json:"dataalamat_ktp_kelurahan"`
	DataalamatKtpKecamatan        string                    `json:"dataalamat_ktp_kecamatan"`
	DataalamatKtpKota             string                    `json:"dataalamat_ktp_kota"`
	DataalamatKtpPropinsi         string                    `json:"dataalamat_ktp_propinsi"`
	DataalamatKtpKodepos          string                    `json:"dataalamat_ktp_kodepos"`
	DataalamatRumahNohp           string                    `json:"dataalamat_rumah_nohp"`
	NorektabPencairanpinjaman     string                    `json:"norektab_pencairanpinjaman"`
	NorektabBayarangsuran         string                    `json:"norektab_bayarangsuran"`
	NorektabPencairanpinjaman2    string                    `json:"norektab_pencairanpinjaman2"`
	NorektabBayarangsuran2        string                    `json:"norektab_bayarangsuran2"`
	Jenisjaminan2                 string                    `json:"jenisjaminan2"`
	Jenisjaminan                  string                    `json:"jenisjaminan"`
	Totalnilaipasar               string                    `json:"totalnilaipasar"`
	Terpakai                      int64                     `json:"terpakai"`
	Totalnilaijaminan             int64                     `json:"totalnilaijaminan"`
	Jaminan                       string                    `json:"jaminan"`
	Jmlagunan                     int64                     `json:"jmlagunan"`
	Tglhapusbuku                  interface{}               `json:"tglhapusbuku"`
	Totalhapusbuku                string                    `json:"totalhapusbuku"`
	NilaihapusbukuSaldopinjaman   interface{}               `json:"nilaihapusbuku_saldopinjaman"`
	NilaihapusbukuBungaberjalan   interface{}               `json:"nilaihapusbuku_bungaberjalan"`
	NilaihapusbukuTunggakanbunga  interface{}               `json:"nilaihapusbuku_tunggakanbunga"`
	NilaihapusbukuTunggakandenda  interface{}               `json:"nilaihapusbuku_tunggakandenda"`
	Ppapblnterakhir               string                    `json:"ppapblnterakhir"`
	Ppaptglterakhir               string                    `json:"ppaptglterakhir"`
	Marketing                     string                    `json:"marketing"`
	DatacsNotes                   interface{}               `json:"datacs_notes"`
	Analiskredit                  interface{}               `json:"analiskredit"`
	AnaliskreditNotes             interface{}               `json:"analiskredit_notes"`
	HTPokok                       string                    `json:"ht_pokok"`
	HTBunga                       string                    `json:"ht_bunga"`
	Tglhapustagih                 interface{}               `json:"tglhapustagih"`
	TotalHT                       interface{}               `json:"total_ht"`
	JmlpokokPinjaman2             int64                     `json:"jmlpokok_pinjaman2"`
	Totalnilaipasar2              int64                     `json:"totalnilaipasar2"`
	Serviceid                     interface{}               `json:"serviceid"`
	Groupid                       interface{}               `json:"groupid"`
	DatatanahBangunan             []DatatanahBangunan       `json:"datatanah_bangunan"`
	Jenisjaminantanah             string                    `json:"jenisjaminantanah"`
	Jenisjaminanlainnya           string                    `json:"jenisjaminanlainnya"`
	Biayapencairan                []Biayapencairan          `json:"biayapencairan"`
	Tabungan                      []Tabungan                `json:"tabungan"`
	Jadwalangsuran                []Jadwalangsuran          `json:"jadwalangsuran"`
	Historybayar                  []Historybayar            `json:"historybayar"`
	Tempatpenyimpanan             bool                      `json:"tempatpenyimpanan"`
	RekeningBungaberjenjang       []RekeningBungaberjenjang `json:"rekening_bungaberjenjang"`
	Channeling                    bool                      `json:"channeling"`
	Asuransidata                  string                    `json:"asuransidata"`
}

type Biayapencairan struct {
	Namabiaya   string `json:"namabiaya"`
	JumlahBiaya string `json:"jumlah_biaya"`
	Hitungdwp   string `json:"hitungdwp"`
}

type DatatanahBangunan struct {
	TanahNosertifikat          string      `json:"tanah_nosertifikat"`
	TanahStatussertifikat      string      `json:"tanah_statussertifikat"`
	TanahImb                   interface{} `json:"tanah_imb"`
	TanahNoimb                 interface{} `json:"tanah_noimb"`
	TanahLokasi                interface{} `json:"tanah_lokasi"`
	TanahRt                    interface{} `json:"tanah_rt"`
	TanahRw                    interface{} `json:"tanah_rw"`
	TanahKelurahan             interface{} `json:"tanah_kelurahan"`
	TanahKecamatan             interface{} `json:"tanah_kecamatan"`
	TanahKota                  interface{} `json:"tanah_kota"`
	TanahProvinsi              interface{} `json:"tanah_provinsi"`
	TanahLuas                  interface{} `json:"tanah_luas"`
	TanahSurattanahdikeluarkan string      `json:"tanah_surattanahdikeluarkan"`
	TanahKodepos               interface{} `json:"tanah_kodepos"`
	TanahAtasnama              string      `json:"tanah_atasnama"`
	TanahNilaijaminan          int64       `json:"tanah_nilaijaminan"`
	TanahNilaipasar            int64       `json:"tanah_nilaipasar"`
	SidJenispengikatan         string      `json:"sid_jenispengikatan"`
	SidDatidua                 string      `json:"sid_datidua"`
	Ansuransi                  string      `json:"ansuransi"`
	Tglpenilaian               DateTZ      `json:"tglpenilaian"`
	TanahNilaijaminanreal      int64       `json:"tanah_nilaijaminanreal"`
	Persenppap                 string      `json:"persenppap"`
	TanahTglsertifikat         interface{} `json:"tanah_tglsertifikat"`
	TanahNosuratukur           interface{} `json:"tanah_nosuratukur"`
	TanahTglukur               interface{} `json:"tanah_tglukur"`
}

type Historybayar struct {
	Tgl                 DateTZ      `json:"tgl"`
	Angsuranke          int64       `json:"angsuranke"`
	Tglbayar            string      `json:"tglbayar"`
	Currency            string      `json:"currency"`
	Tgljt               string      `json:"tgljt"`
	Totalbayar          string      `json:"totalbayar"`
	BayarPokok          string      `json:"bayar_pokok"`
	BayarBunga          string      `json:"bayar_bunga"`
	BayarDenda          string      `json:"bayar_denda"`
	BayarDendapelunasan string      `json:"bayar_dendapelunasan"`
	Nominaldwp          string      `json:"nominaldwp"`
	Nojurnal            string      `json:"nojurnal"`
	Cabang              string      `json:"cabang"`
	Keterangan          interface{} `json:"keterangan"`
	Officer             string      `json:"officer"`
	Otor                string      `json:"otor"`
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
	Sisapinjaman string `json:"sisapinjaman"`
	Statusbayar  string `json:"statusbayar"`
	Angsuranke   int64  `json:"angsuranke"`
}

type RekeningBungaberjenjang struct {
	ID           string       `json:"id"`
	Nourut       string       `json:"nourut"`
	Norekening   string       `json:"norekening"`
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
	Namanasabah     string  `json:"namanasabah"`
	Produkid        string  `json:"produkid"`
	Tglbukarekening *DateTZ `json:"tglbukarekening"`
	Currency        string  `json:"currency"`
	StatusDokumen   string  `json:"status_dokumen"`
	Saldo           string  `json:"saldo"`
	Saldodebit      string  `json:"saldodebit"`
}
