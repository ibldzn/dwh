package models

type TimeDeposit struct {
	NamaNasabahDepo           string           `json:"namanasabahdepo"`
	NamaNasabah               string           `json:"namanasabah"`
	JointAccountType          any              `json:"jointaccounttype"`
	JointAccount              string           `json:"jointaccount"`
	NamaLain                  any              `json:"namalain"`
	JenisNasabah              string           `json:"jenisnasabah"`
	DataAlamatKtpAlamat1      string           `json:"dataalamat_ktp_alamat1"`
	TglBukaCif                DateTZ           `json:"tglbukacif"`
	DataAlamatKtpAlamat2      any              `json:"dataalamat_ktp_alamat2"`
	DataAlamatKtpRt           string           `json:"dataalamat_ktp_rt"`
	DataAlamatKtpRw           string           `json:"dataalamat_ktp_rw"`
	DataAlamatKtpPropinsi     string           `json:"dataalamat_ktp_propinsi"`
	StatusCif                 string           `json:"status_cif"`
	DataAlamatKtpKota         string           `json:"dataalamat_ktp_kota"`
	DataAlamatKtpKecamatan    string           `json:"dataalamat_ktp_kecamatan"`
	DataAlamatKtpKelurahan    string           `json:"dataalamat_ktp_kelurahan"`
	DataAlamatKtpKodepos      string           `json:"dataalamat_ktp_kodepos"`
	ID                        string           `json:"id"`
	JangkaWaktu               string           `json:"jangkawaktu"`
	TglBukaRekening           DateTZ           `json:"tglbukarekening"`
	NoBilyet                  string           `json:"nobilyet"`
	NoCif                     string           `json:"nocif"`
	StatusDokumen             string           `json:"status_dokumen"`
	ProdukID                  string           `json:"produkid"`
	Currency                  string           `json:"currency"`
	ProdukPerubahanSukuBunga  string           `json:"produk_perubahansukubunga"`
	ProdukCetakBilyet         string           `json:"produk_cetakbilyet"`
	ProdukAutomaticRollover   string           `json:"produk_automaticrollover"`
	ProdukJenisDeposito       string           `json:"produk_jenisdeposito"`
	ProdukBungaBerbunga       string           `json:"produk_bungaberbunga"`
	SumberDana                any              `json:"sumberdana"`
	TujuanPembukaanRekening   string           `json:"tujuanpembukaanrekening"`
	Keterangan                any              `json:"keterangan"`
	Nominal                   string           `json:"nominal"`
	AccrueInterest            string           `json:"accrueinterest"`
	Periode                   string           `json:"periode"`
	PembayaranBungaTerakhir   DateTZ           `json:"pembayaran_bungaterakhir"`
	PembayaranBungaBerikutnya DateTZ           `json:"pembayaran_bungaberikutnya"`
	ProdukSukuBunga           float64          `json:"produk_sukubunga"`
	CreatedBy                 string           `json:"createdby"`
	CreatedDate               DateTZ           `json:"createddate"`
	NoRekeningSumberDana      string           `json:"norekeningsumberdana"`
	NoRekTujuanBunga          string           `json:"norektujuanbunga"`
	NoRekeningPencairan       string           `json:"norekeningpencairan"`
	ProdukPembayaranBunga     string           `json:"produk_pembayaran_bunga"`
	TglPencairan              DateTZ           `json:"tglpencairan"`
	LocationID                string           `json:"locationid"`
	RecDibuatLokasi           string           `json:"rec_dibuat_lokasi"`
	DataCsID                  any              `json:"datacs_id"`
	DataCsNotes               any              `json:"datacs_notes"`
	GroupID                   any              `json:"groupid"`
	ServiceID                 any              `json:"serviceid"`
	LinkTabungan              bool             `json:"linktabungan"`
	LinkPinjaman              bool             `json:"linkpinjaman"`
	MutasiDeposito            []Mutasideposito `json:"mutasideposito"`
	RekeningDataJointAccount  any              `json:"rekening_datajointaccount"`
}

type Mutasideposito struct {
	TglTransaksi   string `json:"tgltransaksi"`
	JenisTransaksi string `json:"jenistransaksi"`
	Currency       string `json:"currency"`
	Nominal        string `json:"nominal"`
	Periode        string `json:"periode"`
	JangkaWaktu    string `json:"jangkawaktu"`
	SukuBunga      string `json:"sukubunga"`
	Referensi      string `json:"referensi"`
	Cabang         string `json:"cabang"`
	Officer        string `json:"officer"`
	NoJurnal       string `json:"nojurnal"`
	Keterangan     string `json:"keterangan"`
}
