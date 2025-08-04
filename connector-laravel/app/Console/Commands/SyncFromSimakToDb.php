<?php

namespace App\Console\Commands;

use Carbon\Carbon;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class SyncFromSimakToDb extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:sync-from-simak-to-db';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Command description';

    /**
     * Execute the console command.
     */
    public function handle()
    {
        $mysql = DB::connection('mysql_old');
        $pgsql = DB::connection('pgsql');

        $this->info("🚀 Memulai sinkronisasi bipot → bill_templates...");

        $bipots = $mysql->table('bipot')->where('NA', 'N')->where('Tahun', '>=', '2018')->get();

        foreach ($bipots as $bipot) {
            // Gunakan 'code' sebagai kunci unik
            $code = $bipot->BIPOTID;

            // Upsert ke bill_templates
            $pgsql->table('bill_templates')->updateOrInsert(
                ['code' => $code],
                [
                    'name'          => $bipot->Nama,
                    'academic_year' => $bipot->Tahun,
                    'program_id'    => $bipot->ProgramID,
                    'prodi_id'      => $bipot->ProdiID,
                    'created_at'    => now(),
                    'updated_at'    => now(),
                ]
            );

            // Ambil ID dari bill_template yang baru dibuat/diperbarui
            $template = $pgsql->table('bill_templates')->where('code', $code)->first();
            $templateId = $template->id ?? null;

            if (!$templateId) {
                $this->error("❌ Gagal mengambil ID untuk template dengan kode: {$code}");
                continue;
            }

            $this->info("✅ [{$code}] {$bipot->Nama}");

            // Ambil semua bipot2 terkait
            $bipot2s = $mysql->table('bipot2')
                ->where('BIPOTID', $bipot->BIPOTID)
                ->where('NA', 'N')
                ->get();

            foreach ($bipot2s as $item) {
                $itemName = $item->BIPOT2ID;

                // Upsert ke bill_template_items
                $pgsql->table('bill_template_items')->updateOrInsert(
                    [
                        'bill_template_id' => $templateId,
                        'name'             => $itemName,
                    ],
                    [
                        'additional_name'  => $item->TambahanNama,
                        'amount'           => $item->Jumlah,
                        'ukt' => str_ireplace(".", "", $item->UKT),
                        'BIPOTNamaID' => $item->BIPOTNamaID,
                        'mulai_sesi' => $item->MulaiSesi,
                        'kali_sesi' => $item->KaliSesi,
                        'created_at'       => now(),
                        'updated_at'       => now(),
                    ]
                );
            }

            $this->info("  ↳ {$bipot2s->count()} item(s) disinkronkan.");
        }

        $this->info("🎉 Sinkronisasi selesai.");
    }

}
