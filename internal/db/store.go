package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/ibldzn/dwh-v2/internal/models"
)

type Store struct {
	db *sql.DB
}

const (
	deadlockRetryMax   = 4
	deadlockBaseBackoff = 100 * time.Millisecond
)

func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	return &Store{db: db}, nil
}

func (s *Store) UpsertCIF(ctx context.Context, cif models.CIF) error {
	if cif.ID == "" {
		return fmt.Errorf("cif id is empty")
	}

	var lastErr error
	for attempt := 0; attempt <= deadlockRetryMax; attempt++ {
		if attempt > 0 {
			if err := sleepWithBackoff(ctx, attempt); err != nil {
				return err
			}
		}

		if err := s.upsertCIFOnce(ctx, cif); err != nil {
			lastErr = err
			if isRetryableMySQLError(err) {
				continue
			}
			return err
		}
		return nil
	}
	return lastErr
}

func (s *Store) upsertCIFOnce(ctx context.Context, cif models.CIF) error {
	ptaDate, ptaTZType, ptaTZ := datePartsPtr(cif.PerusahaanTglAktaAwal)
	ptkDate, ptkTZType, ptkTZ := datePartsPtr(cif.PerusahaanTglAktaAkhir)
	tglBukaDate, tglBukaTZType, tglBukaTZ := datePartsPtr(cif.TglBukaCif)
	peroranganTglLahirDate, peroranganTglLahirTZType, peroranganTglLahirTZ := datePartsPtr(cif.PeroranganTglLahir)
	dataKtpTglLahirDate, dataKtpTglLahirTZType, dataKtpTglLahirTZ := datePartsPtr(cif.DataKtpTglLahir)
	recDibuatDate, recDibuatType := recPartsPtr(cif.RecDibuatTglJam)
	recDiupdateDate, recDiupdateType := recPartsPtr(cif.RecDiupdateTglJam)
	recTimestampDate, recTimestampType := recPartsPtr(cif.RecTimestamp)

	plafondLimit := anyToString(cif.PlafondLimit)
	customField := anyToString(cif.CustomField)
	dataDiklat := anyToString(cif.DataDiklat)

	locationName := ""
	if cif.Location != nil {
		locationName = cif.Location.LocationName
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	args := []any{
		cif.ID,
		cif.NamaNasabah,
		cif.JenisNasabah,
		cif.JenisIdentitas,
		cif.NoAlt,
		tglBukaDate,
		tglBukaTZType,
		tglBukaTZ,
		cif.PeroranganNoKtp,
		cif.PeroranganTempatLahir,
		peroranganTglLahirDate,
		peroranganTglLahirTZType,
		peroranganTglLahirTZ,
		cif.PeroranganJenisKelamin,
		cif.PeroranganAgama,
		cif.PeroranganStatusPerkawinan,
		cif.PeroranganPendidikanFormal,
		cif.PeroranganNamaIbuKandung,
		cif.PeroranganJenisAnggota,
		cif.PeroranganLamaMenempatiTahun,
		cif.PeroranganLamaMenempatiBulan,
		cif.DataPekerjaanJenisPekerjaan,
		cif.DataPekerjaanLamaBekerjaTahun,
		cif.DataPekerjaanLamaBekerjaBulan,
		cif.DataPekerjaanLamaBekerjaSebelumnyaTahun,
		cif.DataPekerjaanLamaBekerjaSebelumnyaBulan,
		cif.PerusahaanNoNpwp,
		cif.PerusahaanJenisBadanUsaha,
		cif.PerusahaanNoAktaAwalBerdiri,
		cif.PerusahaanTempatAktaAwal,
		ptaDate,
		ptaTZType,
		ptaTZ,
		cif.PerusahaanNoAktaAkhirBerdiri,
		cif.PerusahaanTempatAktaAkhir,
		ptkDate,
		ptkTZType,
		ptkTZ,
		cif.DataAlamatKtpAlamat1,
		cif.DataAlamatKtpRt,
		cif.DataAlamatKtpRw,
		cif.DataAlamatKtpKelurahan,
		cif.DataAlamatKtpKecamatan,
		cif.DataAlamatKtpKota,
		cif.DataAlamatKtpKodepos,
		cif.DataAlamatKtpPropinsi,
		cif.DataAlamatAlamatRumahAdalahAlamatKtp,
		cif.DataAlamatRumahAlamat1,
		cif.DataAlamatRumahRt,
		cif.DataAlamatRumahRw,
		cif.DataAlamatRumahKelurahan,
		cif.DataAlamatRumahKecamatan,
		cif.DataAlamatRumahKota,
		cif.DataAlamatRumahKodepos,
		cif.DataAlamatRumahPropinsi,
		cif.DataAlamatRumahNoHp,
		cif.DataAlamatKantorKodepos,
		cif.DataKontakLainnyaKodepos,
		cif.DataPenjaminKodepos,
		cif.DataPenjaminLamaBekerjaTahun,
		cif.DataPenjaminLamaBekerjaBulan,
		cif.DataUntukSidNamaAlias,
		cif.DataUntukSidGolonganDebitur,
		cif.DataUntukSidDati2Debitur,
		cif.DataUntukSidStatus,
		cif.DataLabulGolonganDebitur,
		cif.DataKycLimitTransaksiSetoranTunai,
		cif.DataKycLimitTransaksiSetoranNontunai,
		cif.DataKycLimitTransaksiPenarikanTunai,
		cif.DataKycLimitTransaksiPenarikanNontunai,
		cif.DataKycLimitTransaksiFrekuensi,
		cif.DataKycDataNasabahPerusahaanPenghasilan,
		cif.KolekBiManual,
		cif.KolekBprManual,
		cif.KolekBiPinjaman,
		cif.KolekBprPinjaman,
		cif.DataKtpNik,
		cif.DataKtpNama,
		cif.DataKtpTempatLahir,
		cif.DataKtpJenisKelamin,
		cif.DataKtpAlamat,
		cif.DataKtpAgama,
		cif.DataKtpStatusPerkawinan,
		cif.DataKtpPekerjaan,
		dataKtpTglLahirDate,
		dataKtpTglLahirTZType,
		dataKtpTglLahirTZ,
		cif.ProfilResikoIdentitasNasabah,
		cif.ProfilResikoLokasiUsaha,
		cif.ProfilResikoJumlahTransaksi,
		cif.ProfilResikoKegiatanUsaha,
		cif.ProfilResikoStrukturKepemilikan,
		cif.ProfilResikoProdukJasaJaringan,
		cif.ProfilResikoInformasiLain,
		cif.ProfilResikoResumeAkhir,
		cif.ProfilResikoProfil,
		cif.RecDibuatOleh,
		recDibuatDate,
		recDibuatType,
		cif.RecDibuatLokasi,
		cif.RecDiupdateOleh,
		recDiupdateDate,
		recDiupdateType,
		cif.RecDiupdateLokasi,
		recTimestampDate,
		recTimestampType,
		cif.StatusDokumen,
		plafondLimit,
		customField,
		dataDiklat,
		locationName,
		time.Now().UTC(),
	}

	columns := []string{
		"id",
		"nama_nasabah",
		"jenis_nasabah",
		"jenisidentitas",
		"noalt",
		"tglbukacif_date",
		"tglbukacif_timezone_type",
		"tglbukacif_timezone",
		"perorangan_noktp",
		"perorangan_tempatlahir",
		"perorangan_tgllahir_date",
		"perorangan_tgllahir_timezone_type",
		"perorangan_tgllahir_timezone",
		"perorangan_jeniskelamin",
		"perorangan_agama",
		"perorangan_statusperkawinan",
		"perorangan_pendidikanformal",
		"perorangan_namaibukandung",
		"perorangan_jenisanggota",
		"perorangan_lamamenempatitahun",
		"perorangan_lamamenempatibulan",
		"datapekerjaan_jenispekerjaan",
		"datapekerjaan_lamabekerjatahun",
		"datapekerjaan_lamabekerjabulan",
		"datapekerjaan_lamabekerjasebelumnyatahun",
		"datapekerjaan_lamabekerjasebelumnyabulan",
		"perusahaan_nonpwp",
		"perusahaan_jenisbadanusaha",
		"perusahaan_noaktaawalberdiri",
		"perusahaan_tempataktaawal",
		"perusahaan_tglaktaawal_date",
		"perusahaan_tglaktaawal_timezone_type",
		"perusahaan_tglaktaawal_timezone",
		"perusahaan_noaktaakhirberdiri",
		"perusahaan_tempataktaakhir",
		"perusahaan_tglaktaakhir_date",
		"perusahaan_tglaktaakhir_timezone_type",
		"perusahaan_tglaktaakhir_timezone",
		"dataalamat_ktp_alamat1",
		"dataalamat_ktp_rt",
		"dataalamat_ktp_rw",
		"dataalamat_ktp_kelurahan",
		"dataalamat_ktp_kecamatan",
		"dataalamat_ktp_kota",
		"dataalamat_ktp_kodepos",
		"dataalamat_ktp_propinsi",
		"dataalamat_alamatrumahadalahalamatktp",
		"dataalamat_rumah_alamat1",
		"dataalamat_rumah_rt",
		"dataalamat_rumah_rw",
		"dataalamat_rumah_kelurahan",
		"dataalamat_rumah_kecamatan",
		"dataalamat_rumah_kota",
		"dataalamat_rumah_kodepos",
		"dataalamat_rumah_propinsi",
		"dataalamat_rumah_nohp",
		"dataalamat_kantor_kodepos",
		"datakontaklainnya_kodepos",
		"datapenjamin_kodepos",
		"datapenjamin_lamabekerjatahun",
		"datapenjamin_lamabekerjabulan",
		"datauntuksid_namaalias",
		"datauntuksid_golongandebitur",
		"datauntuksid_dati2debitur",
		"datauntuksid_status",
		"datalabul_golongandebitur",
		"datakyc_limittransaksi_setorantunai",
		"datakyc_limittransaksi_setorannontunai",
		"datakyc_limittransaksi_penarikantunai",
		"datakyc_limittransaksi_penarikannontunai",
		"datakyc_limittransaksi_frekuensi",
		"datakyc_datanasabahperusahaan_penghasilan",
		"kolekbi_manual",
		"kolekbpr_manual",
		"kolekbi_pinjaman",
		"kolekbpr_pinjaman",
		"dataktp_nik",
		"dataktp_nama",
		"dataktp_tempatlahir",
		"dataktp_jeniskelamin",
		"dataktp_alamat",
		"dataktp_agama",
		"dataktp_statusperkawinan",
		"dataktp_pekerjaan",
		"dataktp_tgllahir_date",
		"dataktp_tgllahir_timezone_type",
		"dataktp_tgllahir_timezone",
		"profilresiko_identitasnasabah",
		"profilresiko_lokasiusaha",
		"profilresiko_jumlahtransaksi",
		"profilresiko_kegiatanusaha",
		"profilresiko_strukturkepemilikan",
		"profilresiko_produkjasajaringan",
		"profilresiko_informasilain",
		"profilresiko_resumeakhir",
		"profilresiko_profil",
		"rec_dibuat_oleh",
		"rec_dibuat_tgljam_date",
		"rec_dibuat_tgljam_type",
		"rec_dibuat_lokasi",
		"rec_diupdate_oleh",
		"rec_diupdate_tgljam_date",
		"rec_diupdate_tgljam_type",
		"rec_diupdate_lokasi",
		"rec_timestamp_date",
		"rec_timestamp_type",
		"status_dokumen",
		"plafondlimit",
		"customfield",
		"datadiklat",
		"location_locationname",
		"fetched_at",
	}

	query := fmt.Sprintf(`
		INSERT INTO cifs (
			%s
		) VALUES (%s)
		ON DUPLICATE KEY UPDATE
			%s
	`, strings.Join(columns, ",\n\t\t\t"), placeholders(len(args)), buildUpdateList(columns, map[string]bool{"id": true}))

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM cif_data_pengurus_perusahaan WHERE cif_id = ?`, cif.ID); err != nil {
		return err
	}
	for _, pengurus := range cif.DataPengurusPerusahaan {
		tglLahirDate, tglLahirTZType, tglLahirTZ := dateParts(pengurus.TglLahir)
		recDate, recType := recParts(pengurus.RecTimestamp)

		args := []any{
			cif.ID,
			pengurus.ID,
			pengurus.NoCif,
			pengurus.NoUrut,
			pengurus.Nama,
			pengurus.NoNpwp,
			pengurus.TempatLahir,
			tglLahirDate,
			tglLahirTZType,
			tglLahirTZ,
			pengurus.JenisKelamin,
			pengurus.NoKtp,
			anyToString(pengurus.TglDikeluarkan),
			anyToString(pengurus.BerlakuSampai),
			anyToString(pengurus.BerlakuSeumurHidup),
			pengurus.JabatanPengurus,
			pengurus.KepemilikanSaham,
			pengurus.AlamatPengurusAlamat1,
			pengurus.AlamatPengurusAlamat2,
			pengurus.AlamatPengurusRt,
			pengurus.AlamatPengurusRw,
			pengurus.AlamatPengurusKelurahan,
			pengurus.AlamatPengurusKecamatan,
			pengurus.AlamatPengurusKota,
			pengurus.AlamatPengurusPropinsi,
			pengurus.AlamatPengurusKodePos,
			pengurus.AlamatPengurusKodeArea,
			pengurus.AlamatPengurusNoTelp,
			pengurus.AlamatPengurusNoHp,
			pengurus.AlamatPengurusNoFax,
			stringPtr(pengurus.TempatDikeluarkan),
			recDate,
			recType,
		}

		query := fmt.Sprintf(`
			INSERT INTO cif_data_pengurus_perusahaan (
				cif_id,
				id,
				nocif,
				nourut,
				nama,
				nonpwp,
				tempatlahir,
				tgllahir_date,
				tgllahir_timezone_type,
				tgllahir_timezone,
				jeniskelamin,
				noktp,
				tgldikeluarkan,
				berlakusampai,
				berlakuseumurhidup,
				jabatanpengurus,
				kepemilikansaham,
				alamatpengurus_alamat1,
				alamatpengurus_alamat2,
				alamatpengurus_rt,
				alamatpengurus_rw,
				alamatpengurus_kelurahan,
				alamatpengurus_kecamatan,
				alamatpengurus_kota,
				alamatpengurus_propinsi,
				alamatpengurus_kodepos,
				alamatpengurus_kodearea,
				alamatpengurus_notelp,
				alamatpengurus_nohp,
				alamatpengurus_nofax,
				tempat_dikeluarkan,
				rec_timestamp_date,
				rec_timestamp_type
			) VALUES (%s)
		`, placeholders(len(args)))

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM cif_data_ahli_waris WHERE cif_id = ?`, cif.ID); err != nil {
		return err
	}
	for _, ahli := range cif.DataAhliWaris {
		recDate, recType := recParts(ahli.RecTimestamp)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO cif_data_ahli_waris (
				cif_id,
				id,
				nourut,
				nocif,
				ahliwaris_nama,
				ahliwaris_hubdgnkontak,
				ahliwaris_alamat1,
				ahliwaris_alamat2,
				ahliwaris_rt,
				ahliwaris_rw,
				ahliwaris_kelurahan,
				ahliwaris_kecamatan,
				ahliwaris_kota,
				ahliwaris_propinsi,
				ahliwaris_notelp,
				ahliwaris_nohp,
				ahliwaris_nofax,
				ahliwaris_kodepos,
				rec_timestamp_date,
				rec_timestamp_type
			) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		`,
			cif.ID,
			ahli.ID,
			ahli.NoUrut,
			ahli.NoCif,
			anyToString(ahli.AhliWarisNama),
			anyToString(ahli.AhliWarisHubDgnKontak),
			anyToString(ahli.AhliWarisAlamat1),
			anyToString(ahli.AhliWarisAlamat2),
			anyToString(ahli.AhliWarisRt),
			anyToString(ahli.AhliWarisRw),
			anyToString(ahli.AhliWarisKelurahan),
			anyToString(ahli.AhliWarisKecamatan),
			anyToString(ahli.AhliWarisKota),
			anyToString(ahli.AhliWarisPropinsi),
			anyToString(ahli.AhliWarisNoTelp),
			anyToString(ahli.AhliWarisNoHp),
			anyToString(ahli.AhliWarisNoFax),
			ahli.AhliWarisKodePos,
			recDate,
			recType,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) UpsertSaving(ctx context.Context, saving models.Saving) error {
	if saving.NoRekening == "" {
		return fmt.Errorf("saving no_rekening is empty")
	}

	createDate, createTZType, createTZ := dateParts(saving.CreateDate)
	openDate, openTZType, openTZ := dateParts(saving.TglBukaRekening)
	cifDate, cifTZType, cifTZ := dateParts(saving.TglBukaCif)

	args := []any{
		saving.NoRekening,
		saving.LocationID,
		saving.RecDibuatLokasi,
		saving.NoAlt,
		anyToString(saving.NoAlt2),
		saving.NoCif,
		saving.ProdukID,
		saving.Nama,
		saving.StatusDokumen,
		saving.StatusCif,
		createDate,
		createTZType,
		createTZ,
		saving.CreatedBy,
		saving.ProdukJenisTabungan,
		saving.Currency,
		saving.SukuBungaKredit,
		saving.PembayaranBungaKreditTglTetap,
		anyToString(saving.SukuBungaDebit),
		saving.PembayaranBungaDebitTglTetap,
		saving.Overdraft,
		anyToString(saving.NoRekTujuanBunga),
		saving.NamaNasabah,
		saving.JenisTabungan,
		saving.CetakBilyet,
		saving.CetakBukuTabungan,
		saving.ProdukSukuBungaDebit,
		saving.ProdukSukuBungaKredit,
		anyToString(saving.Referral),
		saving.ProdukTransaksiAutoDebit,
		saving.TujuanPembukaanRekening,
		anyToString(saving.SumberDana),
		saving.JointAccount,
		anyToString(saving.AlasanBlokir),
		saving.NilaiBukaBlokir,
		anyToString(saving.TglBlokir),
		anyToString(saving.TglAkhirBlokir),
		anyToString(saving.KeteranganBlokir),
		anyToString(saving.TglBukaBlokir),
		anyToString(saving.AlasanBukaBlokir),
		saving.SaldoBlokir,
		openDate,
		openTZType,
		openTZ,
		anyToString(saving.TglClosed),
		anyToString(saving.BlokirDebet),
		anyToString(saving.BlokirKredit),
		saving.NamaRek,
		anyToString(saving.NamaLain),
		saving.JenisNasabah,
		saving.DataAlamatKtpAlamat1,
		cifDate,
		cifTZType,
		cifTZ,
		anyToString(saving.DataAlamatKtpAlamat2),
		saving.DataAlamatKtpRt,
		saving.DataAlamatKtpRw,
		saving.DataAlamatKtpPropinsi,
		saving.DataAlamatKtpKota,
		saving.DataAlamatKtpKecamatan,
		saving.DataAlamatKtpKelurahan,
		saving.DataAlamatKtpKodepos,
		saving.SaldoAwal,
		saving.SaldoAkhir,
		saving.MutasiDebit,
		saving.MutasiKredit,
		saving.SaldoAkru,
		saving.SaldoAkruDebit,
		saving.PlafondLimit,
		saving.Dpd,
		saving.KolekBi,
		anyToString(saving.Tunggakan),
		anyToString(saving.DendaTunggakan),
		anyToString(saving.TglMulaiTunggakan),
		anyToString(saving.TglBayarTerakhir),
		anyToString(saving.SidGolonganKredit),
		anyToString(saving.SidJenisPenggunaan),
		anyToString(saving.SidJenisUsaha),
		anyToString(saving.SidOrientasiPenggunaan),
		anyToString(saving.SidSumberDanaPelunasan),
		anyToString(saving.SidSifatKredit),
		anyToString(saving.SidSifatKredit2),
		anyToString(saving.SidSektorEkonomi),
		anyToString(saving.SidSektorEkonomi2),
		saving.Terpakai,
		saving.PlafondBlokir,
		saving.TotalSoAktif,
		saving.TotalUsedPlafondOd1,
		saving.TotalAvailablePlafondOd1,
		saving.TotalAvailablePlafond,
		saving.LimitFasilitas,
		anyToString(saving.PeriodeAwalFasilitas),
		anyToString(saving.PeriodeAkhirFasilitas),
		anyToString(saving.PeriodeBatasFasilitasPencairan),
		anyToString(saving.KeteranganFasilitas),
		saving.ProdukBnpl,
		saving.MidRate,
		anyToString(saving.ServiceID),
		anyToString(saving.GroupID),
		anyToString(saving.KodeMarketing),
		anyToString(saving.NotesMarketing),
		saving.TotalAvailablePlafondOd,
		saving.TotalUsedPlafondOd,
		saving.SaldoAwalKredit,
		saving.SaldoAwalDebit,
		saving.SaldoAkhirKredit,
		saving.SaldoAkhirKreditEquivalent,
		saving.SaldoAkhirDebit,
		saving.SaldoAkhirDebitEquivalent,
		saving.SaldoAkruBungaKredit,
		saving.SaldoAkruBungaDebit,
		anyToString(saving.DataStandingOrder),
		saving.AttKtp,
		saving.AttTtd,
		saving.StatusBlokir,
		time.Now().UTC(),
	}

	query := fmt.Sprintf(`
		INSERT INTO savings (
			norekening,
			locationid,
			rec_dibuat_lokasi,
			noalt,
			noalt2,
			nocif,
			produkid,
			nama,
			status_dokumen,
			status_cif,
			createdate_date,
			createdate_timezone_type,
			createdate_timezone,
			createdby,
			produk_jenistabungan,
			currency,
			sukubungakredit,
			pembayaranbunga_kredit_tgltetap,
			sukubungadebit,
			pembayaranbunga_debit_tgltetap,
			overdraft,
			norek_tujuan_bunga,
			namanasabah,
			jenistabungan,
			cetakbilyet,
			cetakbukutabungan,
			produk_sukubungadebit,
			produk_sukubungakredit,
			referral,
			produk_transaksiautodebit,
			tujuanpembukaanrekening,
			sumberdana,
			jointaccount,
			alasanblokir,
			nilaibukablokir,
			tglblokir,
			tglakhirblokir,
			keteranganblokir,
			tglbukablokir,
			alasanbukablokir,
			saldoblokir,
			tglbukarekening_date,
			tglbukarekening_timezone_type,
			tglbukarekening_timezone,
			tglclosed,
			blokirdebet,
			blokirkredit,
			namarek,
			namalain,
			jenisnasabah,
			dataalamat_ktp_alamat1,
			tglbukacif_date,
			tglbukacif_timezone_type,
			tglbukacif_timezone,
			dataalamat_ktp_alamat2,
			dataalamat_ktp_rt,
			dataalamat_ktp_rw,
			dataalamat_ktp_propinsi,
			dataalamat_ktp_kota,
			dataalamat_ktp_kecamatan,
			dataalamat_ktp_kelurahan,
			dataalamat_ktp_kodepos,
			saldoawal,
			saldoakhir,
			mutasidebit,
			mutasikredit,
			saldoakru,
			saldoakrudebit,
			plafondlimit,
			dpd,
			kolekbi,
			tunggakan,
			denda_tunggakan,
			tgl_mulai_tunggakan,
			tgl_bayar_terakhir,
			sid_golongankredit,
			sid_jenispenggunaan,
			sid_jenisusaha,
			sid_orientasipenggunaan,
			sid_sumberdanapelunasan,
			sid_sifatkredit,
			sid_sifatkredit2,
			sid_sektorekonomi,
			sid_sektorekonomi2,
			terpakai,
			plafond_blokir,
			total_so_aktif,
			totalusedplafondod1,
			totalavailableplafondod1,
			totalavailableplafond,
			limitfasilitas,
			periodeawal_fasilitas,
			periodeakhir_fasilitas,
			periode_batas_fasilitas_pencairan,
			keterangan_fasilitas,
			produk_bnpl,
			mid_rate,
			serviceid,
			groupid,
			kode_marketing,
			notes_marketing,
			totalavailableplafondod,
			totalusedplafondod,
			saldoawalkredit,
			saldoawaldebit,
			saldoakhirkredit,
			saldoakhirkredit_equivalent,
			saldoakhirdebit,
			saldoakhirdebit_equivalent,
			saldoakrubungakredit,
			saldoakrubungadebit,
			datastandingorder,
			attKtp,
			attTtd,
			statusblokir,
			fetched_at
		) VALUES (%s)
		ON DUPLICATE KEY UPDATE
			locationid = VALUES(locationid),
			rec_dibuat_lokasi = VALUES(rec_dibuat_lokasi),
			noalt = VALUES(noalt),
			noalt2 = VALUES(noalt2),
			nocif = VALUES(nocif),
			produkid = VALUES(produkid),
			nama = VALUES(nama),
			status_dokumen = VALUES(status_dokumen),
			status_cif = VALUES(status_cif),
			createdate_date = VALUES(createdate_date),
			createdate_timezone_type = VALUES(createdate_timezone_type),
			createdate_timezone = VALUES(createdate_timezone),
			createdby = VALUES(createdby),
			produk_jenistabungan = VALUES(produk_jenistabungan),
			currency = VALUES(currency),
			sukubungakredit = VALUES(sukubungakredit),
			pembayaranbunga_kredit_tgltetap = VALUES(pembayaranbunga_kredit_tgltetap),
			sukubungadebit = VALUES(sukubungadebit),
			pembayaranbunga_debit_tgltetap = VALUES(pembayaranbunga_debit_tgltetap),
			overdraft = VALUES(overdraft),
			norek_tujuan_bunga = VALUES(norek_tujuan_bunga),
			namanasabah = VALUES(namanasabah),
			jenistabungan = VALUES(jenistabungan),
			cetakbilyet = VALUES(cetakbilyet),
			cetakbukutabungan = VALUES(cetakbukutabungan),
			produk_sukubungadebit = VALUES(produk_sukubungadebit),
			produk_sukubungakredit = VALUES(produk_sukubungakredit),
			referral = VALUES(referral),
			produk_transaksiautodebit = VALUES(produk_transaksiautodebit),
			tujuanpembukaanrekening = VALUES(tujuanpembukaanrekening),
			sumberdana = VALUES(sumberdana),
			jointaccount = VALUES(jointaccount),
			alasanblokir = VALUES(alasanblokir),
			nilaibukablokir = VALUES(nilaibukablokir),
			tglblokir = VALUES(tglblokir),
			tglakhirblokir = VALUES(tglakhirblokir),
			keteranganblokir = VALUES(keteranganblokir),
			tglbukablokir = VALUES(tglbukablokir),
			alasanbukablokir = VALUES(alasanbukablokir),
			saldoblokir = VALUES(saldoblokir),
			tglbukarekening_date = VALUES(tglbukarekening_date),
			tglbukarekening_timezone_type = VALUES(tglbukarekening_timezone_type),
			tglbukarekening_timezone = VALUES(tglbukarekening_timezone),
			tglclosed = VALUES(tglclosed),
			blokirdebet = VALUES(blokirdebet),
			blokirkredit = VALUES(blokirkredit),
			namarek = VALUES(namarek),
			namalain = VALUES(namalain),
			jenisnasabah = VALUES(jenisnasabah),
			dataalamat_ktp_alamat1 = VALUES(dataalamat_ktp_alamat1),
			tglbukacif_date = VALUES(tglbukacif_date),
			tglbukacif_timezone_type = VALUES(tglbukacif_timezone_type),
			tglbukacif_timezone = VALUES(tglbukacif_timezone),
			dataalamat_ktp_alamat2 = VALUES(dataalamat_ktp_alamat2),
			dataalamat_ktp_rt = VALUES(dataalamat_ktp_rt),
			dataalamat_ktp_rw = VALUES(dataalamat_ktp_rw),
			dataalamat_ktp_propinsi = VALUES(dataalamat_ktp_propinsi),
			dataalamat_ktp_kota = VALUES(dataalamat_ktp_kota),
			dataalamat_ktp_kecamatan = VALUES(dataalamat_ktp_kecamatan),
			dataalamat_ktp_kelurahan = VALUES(dataalamat_ktp_kelurahan),
			dataalamat_ktp_kodepos = VALUES(dataalamat_ktp_kodepos),
			saldoawal = VALUES(saldoawal),
			saldoakhir = VALUES(saldoakhir),
			mutasidebit = VALUES(mutasidebit),
			mutasikredit = VALUES(mutasikredit),
			saldoakru = VALUES(saldoakru),
			saldoakrudebit = VALUES(saldoakrudebit),
			plafondlimit = VALUES(plafondlimit),
			dpd = VALUES(dpd),
			kolekbi = VALUES(kolekbi),
			tunggakan = VALUES(tunggakan),
			denda_tunggakan = VALUES(denda_tunggakan),
			tgl_mulai_tunggakan = VALUES(tgl_mulai_tunggakan),
			tgl_bayar_terakhir = VALUES(tgl_bayar_terakhir),
			sid_golongankredit = VALUES(sid_golongankredit),
			sid_jenispenggunaan = VALUES(sid_jenispenggunaan),
			sid_jenisusaha = VALUES(sid_jenisusaha),
			sid_orientasipenggunaan = VALUES(sid_orientasipenggunaan),
			sid_sumberdanapelunasan = VALUES(sid_sumberdanapelunasan),
			sid_sifatkredit = VALUES(sid_sifatkredit),
			sid_sifatkredit2 = VALUES(sid_sifatkredit2),
			sid_sektorekonomi = VALUES(sid_sektorekonomi),
			sid_sektorekonomi2 = VALUES(sid_sektorekonomi2),
			terpakai = VALUES(terpakai),
			plafond_blokir = VALUES(plafond_blokir),
			total_so_aktif = VALUES(total_so_aktif),
			totalusedplafondod1 = VALUES(totalusedplafondod1),
			totalavailableplafondod1 = VALUES(totalavailableplafondod1),
			totalavailableplafond = VALUES(totalavailableplafond),
			limitfasilitas = VALUES(limitfasilitas),
			periodeawal_fasilitas = VALUES(periodeawal_fasilitas),
			periodeakhir_fasilitas = VALUES(periodeakhir_fasilitas),
			periode_batas_fasilitas_pencairan = VALUES(periode_batas_fasilitas_pencairan),
			keterangan_fasilitas = VALUES(keterangan_fasilitas),
			produk_bnpl = VALUES(produk_bnpl),
			mid_rate = VALUES(mid_rate),
			serviceid = VALUES(serviceid),
			groupid = VALUES(groupid),
			kode_marketing = VALUES(kode_marketing),
			notes_marketing = VALUES(notes_marketing),
			totalavailableplafondod = VALUES(totalavailableplafondod),
			totalusedplafondod = VALUES(totalusedplafondod),
			saldoawalkredit = VALUES(saldoawalkredit),
			saldoawaldebit = VALUES(saldoawaldebit),
			saldoakhirkredit = VALUES(saldoakhirkredit),
			saldoakhirkredit_equivalent = VALUES(saldoakhirkredit_equivalent),
			saldoakhirdebit = VALUES(saldoakhirdebit),
			saldoakhirdebit_equivalent = VALUES(saldoakhirdebit_equivalent),
			saldoakrubungakredit = VALUES(saldoakrubungakredit),
			saldoakrubungadebit = VALUES(saldoakrubungadebit),
			datastandingorder = VALUES(datastandingorder),
			attKtp = VALUES(attKtp),
			attTtd = VALUES(attTtd),
			statusblokir = VALUES(statusblokir),
			fetched_at = VALUES(fetched_at)
	`, placeholders(len(args)))

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) UpsertLoan(ctx context.Context, loan models.Loan) error {
	if loan.ID == "" {
		return fmt.Errorf("loan id is empty")
	}

	var lastErr error
	for attempt := 0; attempt <= deadlockRetryMax; attempt++ {
		if attempt > 0 {
			if err := sleepWithBackoff(ctx, attempt); err != nil {
				return err
			}
		}

		if err := s.upsertLoanOnce(ctx, loan); err != nil {
			lastErr = err
			if isRetryableMySQLError(err) {
				continue
			}
			return err
		}
		return nil
	}
	return lastErr
}

func (s *Store) upsertLoanOnce(ctx context.Context, loan models.Loan) error {
	restrukturDate, restrukturTZType, restrukturTZ := dateParts(loan.RestrukturTanggalAkhirAkad)
	bayarTerakhirDate, bayarTerakhirTZType, bayarTerakhirTZ := dateParts(loan.TglTerakhirBayarPokokDanBunga)
	bayarBerikutDate, bayarBerikutTZType, bayarBerikutTZ := dateParts(loan.TglBayarPokokBungaBerikutnya)
	jtTerakhirDate, jtTerakhirTZType, jtTerakhirTZ := dateParts(loan.TglJtTerakhir)
	jtBerikutDate, jtBerikutTZType, jtBerikutTZ := dateParts(loan.TglJtBerikutnya)
	tglBukaDate, tglBukaTZType, tglBukaTZ := dateParts(loan.TglBukaCif)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	args := []any{
		loan.ID,
		loan.LocationID,
		loan.RecDibuatLokasi,
		loan.RestrukturNoAkadAkhir,
		restrukturDate,
		restrukturTZType,
		restrukturTZ,
		anyToString(loan.RestrukturTanggalAwal),
		anyToString(loan.RestrukturTanggalAkhir),
		anyToString(loan.RestrukturCara),
		loan.RestrukturFrekuensi,
		anyToString(loan.TanggalMulaiTunggakan),
		loan.Lokasi,
		loan.NoPk,
		loan.NamaNasabah,
		loan.AliasNama,
		loan.StatusRekening,
		loan.TglPencairan,
		loan.RecDibuatOleh,
		loan.NoAlt,
		loan.ProdukJenisPinjaman,
		loan.ProdukID,
		loan.IDProduk,
		loan.Currency,
		loan.PlafondLimit,
		loan.JmlPokokPinjaman,
		loan.JangkaWaktu,
		loan.ProdukJenisAngsuran,
		loan.TglAngsuran,
		loan.SidSifatKredit2,
		loan.SidJenisPenggunaan,
		loan.SidSumberDanaPelunasan,
		loan.SidGolonganKredit,
		loan.SidOrientasiPenggunaan,
		loan.SidSektorEkonomi,
		anyToString(loan.PejabatKredit),
		anyToString(loan.PejabatKreditDua),
		loan.SidSektorEkonomi2,
		loan.SidSifatKredit,
		loan.DataPenjamin,
		loan.MengetahuiSuamiIstri,
		anyToString(loan.DPNama),
		anyToString(loan.DPNoKtp),
		anyToString(loan.DPAlamat),
		anyToString(loan.DPTempatLahir),
		anyToString(loan.DPTglLahir),
		anyToString(loan.DPJenisKelamin),
		anyToString(loan.DPGolDarah),
		anyToString(loan.DPAgama),
		anyToString(loan.DPStatusPerkawinan),
		anyToString(loan.DPPekerjaan),
		anyToString(loan.DPNoCif),
		anyToString(loan.DPHubungan),
		anyToString(loan.DPKewarganegaraan),
		anyToString(loan.DPTempatTerbit),
		anyToString(loan.DPTglTerbit),
		anyToString(loan.DPBerlakuSeumurHidup),
		anyToString(loan.DPTglBerlakuSampai),
		loan.SidJenisUsaha,
		loan.BungaFlat,
		loan.NoPerjanjianKredit,
		loan.JournalID,
		loan.PersenDendaTunggakan,
		loan.BungaBerjenjang,
		loan.NoRekGabunganBnpl,
		anyToString(loan.TglTutup),
		loan.Titipan,
		loan.Periode,
		loan.ProdukSukuBunga,
		loan.ProdukPerubahanSukuBunga,
		anyToString(loan.TujuanKredit),
		bayarTerakhirDate,
		bayarTerakhirTZType,
		bayarTerakhirTZ,
		bayarBerikutDate,
		bayarBerikutTZType,
		bayarBerikutTZ,
		jtTerakhirDate,
		jtTerakhirTZType,
		jtTerakhirTZ,
		jtBerikutDate,
		jtBerikutTZType,
		jtBerikutTZ,
		loan.OutstandingPinjaman,
		loan.TunggakanPokok,
		loan.Accrue,
		loan.DecimalPoint,
		loan.TunggakanBunga,
		loan.DendaTunggakan,
		loan.Dpd,
		loan.KolekBi,
		loan.KolekBpr,
		loan.UpdateKolekBi,
		loan.TotalCollateralValue,
		loan.TotalAssetValue,
		loan.NoCif,
		loan.JenisNasabah,
		tglBukaDate,
		tglBukaTZType,
		tglBukaTZ,
		loan.StatusDokumen,
		loan.DataAlamatKtpAlamat1,
		anyToString(loan.DataAlamatKtpAlamat2),
		loan.DataAlamatKtpRt,
		loan.DataAlamatKtpRw,
		loan.DataAlamatKtpKelurahan,
		loan.DataAlamatKtpKecamatan,
		loan.DataAlamatKtpKota,
		loan.DataAlamatKtpPropinsi,
		loan.DataAlamatKtpKodePos,
		loan.DataAlamatRumahNoHp,
		loan.NoRekTabPencairanPinjaman,
		loan.NoRekTabBayarAngsuran,
		loan.NoRekTabPencairanPinjaman2,
		loan.NoRekTabBayarAngsuran2,
		loan.JenisJaminan2,
		loan.JenisJaminan,
		loan.TotalNilaiPasar,
		loan.Terpakai,
		loan.TotalNilaiJaminan,
		loan.Jaminan,
		loan.JmlAgunan,
		anyToString(loan.TglHapusBuku),
		loan.TotalHapusBuku,
		anyToString(loan.NilaiHapusBukuSaldoPinjaman),
		anyToString(loan.NilaiHapusBukuBungaBerjalan),
		anyToString(loan.NilaiHapusBukuTunggakanBunga),
		anyToString(loan.NilaiHapusBukuTunggakanDenda),
		loan.PpapBlnTerakhir,
		loan.PpapTglTerakhir,
		loan.Marketing,
		anyToString(loan.DataCsNotes),
		anyToString(loan.AnalisKredit),
		anyToString(loan.AnalisKreditNotes),
		loan.HTPokok,
		loan.HTBunga,
		anyToString(loan.TglHapusTagih),
		anyToString(loan.TotalHT),
		loan.JmlPokokPinjaman2,
		loan.TotalNilaiPasar2,
		anyToString(loan.ServiceID),
		anyToString(loan.GroupID),
		loan.JenisJaminanTanah,
		loan.JenisJaminanLainnya,
		loan.TempatPenyimpanan,
		loan.Channeling,
		loan.AsuransiData,
		time.Now().UTC(),
	}

	columns := []string{
		"id",
		"locationid",
		"rec_dibuat_lokasi",
		"restruktur_noakad_akhir",
		"restruktur_tanggalakhirakad_date",
		"restruktur_tanggalakhirakad_timezone_type",
		"restruktur_tanggalakhirakad_timezone",
		"restruktur_tanggalawal",
		"restruktur_tanggalakhir",
		"restruktur_cara",
		"restruktur_frekuensi",
		"tanggalmulaitunggakan",
		"lokasi",
		"nopk",
		"namanasabah",
		"aliasnama",
		"statusrekening",
		"tgl_pencairan",
		"rec_dibuat_oleh",
		"noalt",
		"produk_jenispinjaman",
		"produkid",
		"idproduk",
		"currency",
		"plafondlimit",
		"jmlpokok_pinjaman",
		"jangkawaktu",
		"produk_jenisangsuran",
		"tgl_angsuran",
		"sid_sifatkredit2",
		"sid_jenispenggunaan",
		"sid_sumberdanapelunasan",
		"sid_golongankredit",
		"sid_orientasipenggunaan",
		"sid_sektorekonomi",
		"pejabatkredit",
		"pejabatkreditdua",
		"sid_sektorekonomi2",
		"sid_sifatkredit",
		"datapenjamin",
		"mengetahuisuamiistri",
		"dp_nama",
		"dp_noktp",
		"dp_alamat",
		"dp_tempatlahir",
		"dp_tgllahir",
		"dp_jeniskelamin",
		"dp_goldarah",
		"dp_agama",
		"dp_statusperkawinan",
		"dp_pekerjaan",
		"dp_nocif",
		"dp_hubungan",
		"dp_kewarganegaraan",
		"dp_tempatterbit",
		"dp_tglterbit",
		"dp_berlakuseumurhidup",
		"dp_tglberlakusampai",
		"sid_jenisusaha",
		"bungaflat",
		"noperjanjiankredit",
		"journalid",
		"persendendatunggakan",
		"bungaberjenjang",
		"norekgabungan_bnpl",
		"tgl_tutup",
		"titipan",
		"periode",
		"produk_sukubunga",
		"produk_perubahansukubunga",
		"tujuankredit",
		"tglterakhir_bayarpokokdanbunga_date",
		"tglterakhir_bayarpokokdanbunga_timezone_type",
		"tglterakhir_bayarpokokdanbunga_timezone",
		"tglbayarpokokbunga_berikutnya_date",
		"tglbayarpokokbunga_berikutnya_timezone_type",
		"tglbayarpokokbunga_berikutnya_timezone",
		"tgljtterakhir_date",
		"tgljtterakhir_timezone_type",
		"tgljtterakhir_timezone",
		"tgljtberikutnya_date",
		"tgljtberikutnya_timezone_type",
		"tgljtberikutnya_timezone",
		"outstandingpinjaman",
		"tunggakanpokok",
		"accrue",
		"decimalpoint",
		"tunggakanbunga",
		"denda_tunggakan",
		"dpd",
		"kolekbi",
		"kolekbpr",
		"updatekolekbi",
		"totalcollateralvalue",
		"totalassetvalue",
		"nocif",
		"jenisnasabah",
		"tglbukacif_date",
		"tglbukacif_timezone_type",
		"tglbukacif_timezone",
		"status_dokumen",
		"dataalamat_ktp_alamat1",
		"dataalamat_ktp_alamat2",
		"dataalamat_ktp_rt",
		"dataalamat_ktp_rw",
		"dataalamat_ktp_kelurahan",
		"dataalamat_ktp_kecamatan",
		"dataalamat_ktp_kota",
		"dataalamat_ktp_propinsi",
		"dataalamat_ktp_kodepos",
		"dataalamat_rumah_nohp",
		"norektab_pencairanpinjaman",
		"norektab_bayarangsuran",
		"norektab_pencairanpinjaman2",
		"norektab_bayarangsuran2",
		"jenisjaminan2",
		"jenisjaminan",
		"totalnilaipasar",
		"terpakai",
		"totalnilaijaminan",
		"jaminan",
		"jmlagunan",
		"tglhapusbuku",
		"totalhapusbuku",
		"nilaihapusbuku_saldopinjaman",
		"nilaihapusbuku_bungaberjalan",
		"nilaihapusbuku_tunggakanbunga",
		"nilaihapusbuku_tunggakandenda",
		"ppapblnterakhir",
		"ppaptglterakhir",
		"marketing",
		"datacs_notes",
		"analiskredit",
		"analiskredit_notes",
		"ht_pokok",
		"ht_bunga",
		"tglhapustagih",
		"total_ht",
		"jmlpokok_pinjaman2",
		"totalnilaipasar2",
		"serviceid",
		"groupid",
		"jenisjaminantanah",
		"jenisjaminanlainnya",
		"tempatpenyimpanan",
		"channeling",
		"asuransidata",
		"fetched_at",
	}

	query := fmt.Sprintf(`
		INSERT INTO loans (
			%s
		) VALUES (%s)
		ON DUPLICATE KEY UPDATE
			%s
	`, strings.Join(columns, ",\n\t\t\t"), placeholders(len(args)), buildUpdateList(columns, map[string]bool{"id": true}))

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM loan_datatanah_bangunan WHERE loan_id = ?`, loan.ID); err != nil {
		return err
	}
	for _, tanah := range loan.DataTanahBangunan {
		penilaianDate, penilaianTZType, penilaianTZ := dateParts(tanah.TglPenilaian)
		args := []any{
			loan.ID,
			tanah.TanahNoSertifikat,
			tanah.TanahStatusSertifikat,
			anyToString(tanah.TanahImb),
			anyToString(tanah.TanahNoImb),
			anyToString(tanah.TanahLokasi),
			anyToString(tanah.TanahRt),
			anyToString(tanah.TanahRw),
			anyToString(tanah.TanahKelurahan),
			anyToString(tanah.TanahKecamatan),
			anyToString(tanah.TanahKota),
			anyToString(tanah.TanahProvinsi),
			anyToString(tanah.TanahLuas),
			tanah.TanahSuratTanahDikeluarkan,
			anyToString(tanah.TanahKodePos),
			tanah.TanahAtasNama,
			tanah.TanahNilaiJaminan,
			tanah.TanahNilaiPasar,
			tanah.SidJenisPengikatan,
			tanah.SidDatiDua,
			tanah.Asuransi,
			penilaianDate,
			penilaianTZType,
			penilaianTZ,
			tanah.TanahNilaiJaminanReal,
			tanah.PersenPpap,
			anyToString(tanah.TanahTglSertifikat),
			anyToString(tanah.TanahNoSuratUkur),
			anyToString(tanah.TanahTglUkur),
		}

		query := fmt.Sprintf(`
			INSERT INTO loan_datatanah_bangunan (
				loan_id,
				tanah_nosertifikat,
				tanah_statussertifikat,
				tanah_imb,
				tanah_noimb,
				tanah_lokasi,
				tanah_rt,
				tanah_rw,
				tanah_kelurahan,
				tanah_kecamatan,
				tanah_kota,
				tanah_provinsi,
				tanah_luas,
				tanah_surattanahdikeluarkan,
				tanah_kodepos,
				tanah_atasnama,
				tanah_nilaijaminan,
				tanah_nilaipasar,
				sid_jenispengikatan,
				sid_datidua,
				ansuransi,
				tglpenilaian_date,
				tglpenilaian_timezone_type,
				tglpenilaian_timezone,
				tanah_nilaijaminanreal,
				persenppap,
				tanah_tglsertifikat,
				tanah_nosuratukur,
				tanah_tglukur
			) VALUES (%s)
		`, placeholders(len(args)))

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM loan_biaya_pencairan WHERE loan_id = ?`, loan.ID); err != nil {
		return err
	}
	for _, biaya := range loan.BiayaPencairan {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO loan_biaya_pencairan (
				loan_id,
				namabiaya,
				jumlah_biaya,
				hitungdwp
			) VALUES (?,?,?,?)
		`,
			loan.ID,
			biaya.NamaBiaya,
			biaya.JumlahBiaya,
			biaya.HitungDwp,
		)
		if err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM loan_tabungan WHERE loan_id = ?`, loan.ID); err != nil {
		return err
	}
	for _, tab := range loan.Tabungan {
		openDate, openTZType, openTZ := datePartsPtr(tab.TglBukaRekening)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO loan_tabungan (
				loan_id,
				id,
				namanasabah,
				produkid,
				tglbukarekening_date,
				tglbukarekening_timezone_type,
				tglbukarekening_timezone,
				currency,
				status_dokumen,
				saldo,
				saldodebit
			) VALUES (?,?,?,?,?,?,?,?,?,?,?)
		`,
			loan.ID,
			tab.ID,
			tab.NamaNasabah,
			tab.ProdukID,
			openDate,
			openTZType,
			openTZ,
			tab.Currency,
			tab.StatusDokumen,
			tab.Saldo,
			tab.SaldoDebit,
		)
		if err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM loan_jadwal_angsuran WHERE loan_id = ?`, loan.ID); err != nil {
		return err
	}
	for _, jadwal := range loan.JadwalAngsuran {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO loan_jadwal_angsuran (
				loan_id,
				tanggal,
				angsuran,
				bunga,
				pokok,
				denda,
				bayar_pokok,
				bayar_denda,
				bayar_bunga,
				sisapinjaman,
				statusbayar,
				angsuranke
			) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		`,
			loan.ID,
			jadwal.Tanggal,
			jadwal.Angsuran,
			jadwal.Bunga,
			jadwal.Pokok,
			jadwal.Denda,
			jadwal.BayarPokok,
			jadwal.BayarDenda,
			jadwal.BayarBunga,
			jadwal.SisaPinjaman,
			jadwal.StatusBayar,
			jadwal.AngsuranKe,
		)
		if err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM loan_history_bayar WHERE loan_id = ?`, loan.ID); err != nil {
		return err
	}
	for _, hist := range loan.HistoryBayar {
		tglDate, tglTZType, tglTZ := dateParts(hist.Tgl)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO loan_history_bayar (
				loan_id,
				tgl_date,
				tgl_timezone_type,
				tgl_timezone,
				angsuranke,
				tglbayar,
				currency,
				tgljt,
				totalbayar,
				bayar_pokok,
				bayar_bunga,
				bayar_denda,
				bayar_dendapelunasan,
				nominaldwp,
				nojurnal,
				cabang,
				keterangan,
				officer,
				otor
			) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		`,
			loan.ID,
			tglDate,
			tglTZType,
			tglTZ,
			hist.AngsuranKe,
			hist.TglBayar,
			hist.Currency,
			hist.Tgljt,
			hist.TotalBayar,
			hist.BayarPokok,
			hist.BayarBunga,
			hist.BayarDenda,
			hist.BayarDendaPelunasan,
			hist.NominalDwp,
			hist.NoJurnal,
			hist.Cabang,
			anyToString(hist.Keterangan),
			hist.Officer,
			hist.Otor,
		)
		if err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM loan_rekening_bungaberjenjang WHERE loan_id = ?`, loan.ID); err != nil {
		return err
	}
	for _, rek := range loan.RekeningBungaBerjenjang {
		recDate, recType := recTimestampParts(rek.RecTimestamp)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO loan_rekening_bungaberjenjang (
				loan_id,
				id,
				nourut,
				norekening,
				tenor,
				bunga,
				rec_timestamp_date,
				rec_timestamp_type
			) VALUES (?,?,?,?,?,?,?,?)
		`,
			loan.ID,
			rek.ID,
			rek.NoUrut,
			rek.NoRekening,
			rek.Tenor,
			rek.Bunga,
			recDate,
			recType,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func anyToString(value any) any {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		return v
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(data)
	}
}

func stringPtr(value *string) any {
	if value == nil {
		return nil
	}
	if *value == "" {
		return nil
	}
	return *value
}

func dateParts(value models.DateTZ) (any, any, any) {
	if value.Date == "" && value.Timezone == "" && value.TimezoneType == 0 {
		return nil, nil, nil
	}
	return value.Date, value.TimezoneType, value.Timezone
}

func datePartsPtr(value *models.DateTZ) (any, any, any) {
	if value == nil {
		return nil, nil, nil
	}
	return dateParts(*value)
}

func recParts(value models.Rec) (any, any) {
	if value.Date == "" && value.Type == "" {
		return nil, nil
	}
	return value.Date, value.Type
}

func recPartsPtr(value *models.Rec) (any, any) {
	if value == nil {
		return nil, nil
	}
	return recParts(*value)
}

func recTimestampParts(value models.RecTimestamp) (any, any) {
	if value.Date == "" && value.Type == "" {
		return nil, nil
	}
	return value.Date, value.Type
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func buildUpdateList(columns []string, skip map[string]bool) string {
	parts := make([]string, 0, len(columns))
	for _, column := range columns {
		if skip[column] {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s = VALUES(%s)", column, column))
	}
	return strings.Join(parts, ",\n\t\t\t")
}

func isRetryableMySQLError(err error) bool {
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) {
		switch myErr.Number {
		case 1213, 1205:
			return true
		}
	}
	return false
}

func sleepWithBackoff(ctx context.Context, attempt int) error {
	backoff := deadlockBaseBackoff * time.Duration(1<<attempt)
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
