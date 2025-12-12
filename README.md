> New version release from https://github.com/devetek/loki-scraper

## Loki Data Migration

Tools ini dapat digunakan dengan kubernetes pod ataupun kubernetes job. Berikut step-step cara menggunakan:

### Pod

Untuk dapat menggunakan image, lihat file `Makefile`. Ubah target registry image sesuai dengan yang kamu miliki. Disini menggunakan `prakasa1904/loki-scraper`. Kemudian jalankan perintah berikut:

1. Build image dengan perintah `make build-docker`
2. Publish image dengan perintah `make publish-docker`
3. Buat kubernetes pod, dengan image yang sudah dipublish

### Job

Buat kubernetes job, gunakan image yang kamu familiar. Download binary di `https://github.com/prakasa1904/loki-scraper/releases` sesuai dengan architecture yang kamu gunakan. Kemudian ikuti langkah-langkah berikut:

1. Ubah query sesuai dengan kebutuhan, set limit jika diperlukan, pilih range waktu Loki yang akan di extract menggunakan environment variable.
2. Jalankan perintah `./scripts/prod.sh`
