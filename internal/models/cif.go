package models

type LocationInfo struct {
	LocationName string `json:"locationname,omitempty"`
}

type CIF struct {
	// Common (biasanya wajib)
	ID             string `json:"id"`
	NamaNasabah    string `json:"namanasabah,omitempty"`
	JenisNasabah   string `json:"jenisnasabah,omitempty"`
	JenisIdentitas string `json:"jenisidentitas,omitempty"`
	NoAlt          string `json:"noalt,omitempty"`

	TglBukaCif *DateTZ `json:"tglbukacif,omitempty"`

	// Perorangan (optional)
	PeroranganNoKtp            string  `json:"perorangan_noktp,omitempty"`
	PeroranganTempatLahir      string  `json:"perorangan_tempatlahir,omitempty"`
	PeroranganTglLahir         *DateTZ `json:"perorangan_tgllahir,omitempty"`
	PeroranganJenisKelamin     string  `json:"perorangan_jeniskelamin,omitempty"`
	PeroranganAgama            string  `json:"perorangan_agama,omitempty"`
	PeroranganStatusPerkawinan string  `json:"perorangan_statusperkawinan,omitempty"`
	PeroranganPendidikanFormal string  `json:"perorangan_pendidikanformal,omitempty"`
	PeroranganNamaIbuKandung   string  `json:"perorangan_namaibukandung,omitempty"`
	PeroranganJenisAnggota     string  `json:"perorangan_jenisanggota,omitempty"`

	// Shared (exists in both)
	PeroranganLamaMenempatiTahun int64 `json:"perorangan_lamamenempatitahun,omitempty"`
	PeroranganLamaMenempatiBulan int64 `json:"perorangan_lamamenempatibulan,omitempty"`

	// Pekerjaan (perorangan punya jenispekerjaan, badan tidak)
	DataPekerjaanJenisPekerjaan string `json:"datapekerjaan_jenispekerjaan,omitempty"`

	// Shared lama bekerja
	DataPekerjaanLamaBekerjaTahun           int64 `json:"datapekerjaan_lamabekerjatahun,omitempty"`
	DataPekerjaanLamaBekerjaBulan           int64 `json:"datapekerjaan_lamabekerjabulan,omitempty"`
	DataPekerjaanLamaBekerjaSebelumnyaTahun int64 `json:"datapekerjaan_lamabekerjasebelumnyatahun,omitempty"`
	DataPekerjaanLamaBekerjaSebelumnyaBulan int64 `json:"datapekerjaan_lamabekerjasebelumnyabulan,omitempty"`

	// Alamat (perorangan)
	DataAlamatKtpAlamat1                 string `json:"dataalamat_ktp_alamat1,omitempty"`
	DataAlamatKtpRt                      string `json:"dataalamat_ktp_rt,omitempty"`
	DataAlamatKtpRw                      string `json:"dataalamat_ktp_rw,omitempty"`
	DataAlamatKtpKelurahan               string `json:"dataalamat_ktp_kelurahan,omitempty"`
	DataAlamatKtpKecamatan               string `json:"dataalamat_ktp_kecamatan,omitempty"`
	DataAlamatKtpKota                    string `json:"dataalamat_ktp_kota,omitempty"`
	DataAlamatKtpKodepos                 string `json:"dataalamat_ktp_kodepos,omitempty"`
	DataAlamatKtpPropinsi                string `json:"dataalamat_ktp_propinsi,omitempty"`
	DataAlamatAlamatRumahAdalahAlamatKtp string `json:"dataalamat_alamatrumahadalahalamatktp,omitempty"`

	DataAlamatRumahAlamat1   string `json:"dataalamat_rumah_alamat1,omitempty"`
	DataAlamatRumahRt        string `json:"dataalamat_rumah_rt,omitempty"`
	DataAlamatRumahRw        string `json:"dataalamat_rumah_rw,omitempty"`
	DataAlamatRumahKelurahan string `json:"dataalamat_rumah_kelurahan,omitempty"`
	DataAlamatRumahKecamatan string `json:"dataalamat_rumah_kecamatan,omitempty"`
	DataAlamatRumahKota      string `json:"dataalamat_rumah_kota,omitempty"`
	DataAlamatRumahKodepos   string `json:"dataalamat_rumah_kodepos,omitempty"`
	DataAlamatRumahPropinsi  string `json:"dataalamat_rumah_propinsi,omitempty"`
	DataAlamatRumahNoHp      string `json:"dataalamat_rumah_nohp,omitempty"`

	// Shared ints
	DataAlamatKantorKodepos      int64 `json:"dataalamat_kantor_kodepos,omitempty"`
	DataKontakLainnyaKodepos     int64 `json:"datakontaklainnya_kodepos,omitempty"`
	DataPenjaminKodepos          int64 `json:"datapenjamin_kodepos,omitempty"`
	DataPenjaminLamaBekerjaTahun int64 `json:"datapenjamin_lamabekerjatahun,omitempty"`
	DataPenjaminLamaBekerjaBulan int64 `json:"datapenjamin_lamabekerjabulan,omitempty"`

	// SID/Labul (perorangan)
	DataUntukSidNamaAlias       string `json:"datauntuksid_namaalias,omitempty"`
	DataUntukSidGolonganDebitur string `json:"datauntuksid_golongandebitur,omitempty"`
	DataUntukSidDati2Debitur    string `json:"datauntuksid_dati2debitur,omitempty"`
	DataUntukSidStatus          string `json:"datauntuksid_status,omitempty"`
	DataLabulGolonganDebitur    string `json:"datalabul_golongandebitur,omitempty"`

	// KYC (shared)
	DataKycLimitTransaksiSetoranTunai       int64 `json:"datakyc_limittransaksi_setorantunai,omitempty"`
	DataKycLimitTransaksiSetoranNontunai    int64 `json:"datakyc_limittransaksi_setorannontunai,omitempty"`
	DataKycLimitTransaksiPenarikanTunai     int64 `json:"datakyc_limittransaksi_penarikantunai,omitempty"`
	DataKycLimitTransaksiPenarikanNontunai  int64 `json:"datakyc_limittransaksi_penarikannontunai,omitempty"`
	DataKycLimitTransaksiFrekuensi          int64 `json:"datakyc_limittransaksi_frekuensi,omitempty"`
	DataKycDataNasabahPerusahaanPenghasilan int64 `json:"datakyc_datanasabahperusahaan_penghasilan,omitempty"`

	// Perusahaan (badan)
	PerusahaanNoNpwp             string  `json:"perusahaan_nonpwp,omitempty"`
	PerusahaanJenisBadanUsaha    string  `json:"perusahaan_jenisbadanusaha,omitempty"`
	PerusahaanNoAktaAwalBerdiri  string  `json:"perusahaan_noaktaawalberdiri,omitempty"`
	PerusahaanTempatAktaAwal     string  `json:"perusahaan_tempataktaawal,omitempty"`
	PerusahaanTglAktaAwal        *DateTZ `json:"perusahaan_tglaktaawal,omitempty"`
	PerusahaanNoAktaAkhirBerdiri string  `json:"perusahaan_noaktaakhirberdiri,omitempty"`
	PerusahaanTempatAktaAkhir    string  `json:"perusahaan_tempataktaakhir,omitempty"`
	PerusahaanTglAktaAkhir       *DateTZ `json:"perusahaan_tglaktaakhir,omitempty"`

	// Kolektibilitas (shared)
	KolekBiManual    int64 `json:"kolekbi_manual,omitempty"`
	KolekBprManual   int64 `json:"kolekbpr_manual,omitempty"`
	KolekBiPinjaman  int64 `json:"kolekbi_pinjaman,omitempty"`
	KolekBprPinjaman int64 `json:"kolekbpr_pinjaman,omitempty"`

	// Data KTP (perorangan)
	DataKtpNik              string  `json:"dataktp_nik,omitempty"`
	DataKtpNama             string  `json:"dataktp_nama,omitempty"`
	DataKtpTempatLahir      string  `json:"dataktp_tempatlahir,omitempty"`
	DataKtpJenisKelamin     string  `json:"dataktp_jeniskelamin,omitempty"`
	DataKtpAlamat           string  `json:"dataktp_alamat,omitempty"`
	DataKtpAgama            string  `json:"dataktp_agama,omitempty"`
	DataKtpStatusPerkawinan string  `json:"dataktp_statusperkawinan,omitempty"`
	DataKtpPekerjaan        string  `json:"dataktp_pekerjaan,omitempty"`
	DataKtpTglLahir         *DateTZ `json:"dataktp_tgllahir,omitempty"`

	// Profil risiko (shared)
	ProfilResikoIdentitasNasabah    string `json:"profilresiko_identitasnasabah,omitempty"`
	ProfilResikoLokasiUsaha         string `json:"profilresiko_lokasiusaha,omitempty"`
	ProfilResikoJumlahTransaksi     string `json:"profilresiko_jumlahtransaksi,omitempty"`
	ProfilResikoKegiatanUsaha       string `json:"profilresiko_kegiatanusaha,omitempty"`
	ProfilResikoStrukturKepemilikan string `json:"profilresiko_strukturkepemilikan,omitempty"`
	ProfilResikoProdukJasaJaringan  string `json:"profilresiko_produkjasajaringan,omitempty"`
	ProfilResikoInformasiLain       string `json:"profilresiko_informasilain,omitempty"`
	ProfilResikoResumeAkhir         string `json:"profilresiko_resumeakhir,omitempty"`
	ProfilResikoProfil              string `json:"profilresiko_profil,omitempty"`

	// Audit (pointer biar bisa omit)
	RecDibuatOleh     string `json:"rec_dibuat_oleh,omitempty"`
	RecDibuatTglJam   *Rec   `json:"rec_dibuat_tgljam,omitempty"`
	RecDibuatLokasi   string `json:"rec_dibuat_lokasi,omitempty"`
	RecDiupdateOleh   string `json:"rec_diupdate_oleh,omitempty"`
	RecDiupdateTglJam *Rec   `json:"rec_diupdate_tgljam,omitempty"`
	RecDiupdateLokasi string `json:"rec_diupdate_lokasi,omitempty"`
	RecTimestamp      *Rec   `json:"rec_timestamp,omitempty"`

	// Collections
	DataPengurusPerusahaan []DataPengurusPerusahaan `json:"datapengurusperusahaan,omitempty"`
	CustomField            []any                    `json:"customfield,omitempty"`
	DataDiklat             []any                    `json:"datadiklat,omitempty"`
	DataAhliWaris          []DataAhliWaris          `json:"dataahliwaris,omitempty"`

	StatusDokumen string `json:"status_dokumen,omitempty"`
	PlafondLimit  any    `json:"plafondlimit,omitempty"`

	Location *LocationInfo `json:"location,omitempty"`
}

type DataPengurusPerusahaan struct {
	ID                      string  `json:"id"`
	NoCif                   string  `json:"nocif"`
	NoUrut                  int64   `json:"nourut"`
	Nama                    string  `json:"nama"`
	NoNpwp                  string  `json:"nonpwp"`
	TempatLahir             string  `json:"tempatlahir"`
	TglLahir                DateTZ  `json:"tgllahir"`
	JenisKelamin            string  `json:"jeniskelamin"`
	NoKtp                   string  `json:"noktp"`
	TglDikeluarkan          any     `json:"tgldikeluarkan"`
	BerlakuSampai           any     `json:"berlakusampai"`
	BerlakuSeumurHidup      any     `json:"berlakuseumurhidup"`
	JabatanPengurus         string  `json:"jabatanpengurus"`
	KepemilikanSaham        int64   `json:"kepemilikansaham"`
	AlamatPengurusAlamat1   string  `json:"alamatpengurus_alamat1"`
	AlamatPengurusAlamat2   string  `json:"alamatpengurus_alamat2"`
	AlamatPengurusRt        string  `json:"alamatpengurus_rt"`
	AlamatPengurusRw        string  `json:"alamatpengurus_rw"`
	AlamatPengurusKelurahan string  `json:"alamatpengurus_kelurahan"`
	AlamatPengurusKecamatan string  `json:"alamatpengurus_kecamatan"`
	AlamatPengurusKota      string  `json:"alamatpengurus_kota"`
	AlamatPengurusPropinsi  string  `json:"alamatpengurus_propinsi"`
	AlamatPengurusKodePos   string  `json:"alamatpengurus_kodepos"`
	AlamatPengurusKodeArea  string  `json:"alamatpengurus_kodearea"`
	AlamatPengurusNoTelp    string  `json:"alamatpengurus_notelp"`
	AlamatPengurusNoHp      string  `json:"alamatpengurus_nohp"`
	AlamatPengurusNoFax     string  `json:"alamatpengurus_nofax"`
	TempatDikeluarkan       *string `json:"tempat_dikeluarkan"`
	RecTimestamp            Rec     `json:"rec_timestamp"`
}

type DataAhliWaris struct {
	ID                    string `json:"id"`
	NoUrut                string `json:"nourut"`
	NoCif                 string `json:"nocif"`
	AhliWarisNama         any    `json:"ahliwaris_nama"`
	AhliWarisHubDgnKontak any    `json:"ahliwaris_hubdgnkontak"`
	AhliWarisAlamat1      any    `json:"ahliwaris_alamat1"`
	AhliWarisAlamat2      any    `json:"ahliwaris_alamat2"`
	AhliWarisRt           any    `json:"ahliwaris_rt"`
	AhliWarisRw           any    `json:"ahliwaris_rw"`
	AhliWarisKelurahan    any    `json:"ahliwaris_kelurahan"`
	AhliWarisKecamatan    any    `json:"ahliwaris_kecamatan"`
	AhliWarisKota         any    `json:"ahliwaris_kota"`
	AhliWarisPropinsi     any    `json:"ahliwaris_propinsi"`
	AhliWarisNoTelp       any    `json:"ahliwaris_notelp"`
	AhliWarisNoHp         any    `json:"ahliwaris_nohp"`
	AhliWarisNoFax        any    `json:"ahliwaris_nofax"`
	AhliWarisKodePos      string `json:"ahliwaris_kodepos"`
	RecTimestamp          Rec    `json:"rec_timestamp"`
}
