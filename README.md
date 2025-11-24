## Loki Data Migration

Untuk dapat menggunakan, lakukan perubahan pada code. Commit changes. Kemudian execute `make build`.

Image baru akan dibuat under repository `prakasa1904/loki-scraper`. Ubah sesuai dengan scope registry kamu misal `aseppatahudin/loki-scraper`.

Selanjutnya lihat file `scripts/dev-docker.sh`. Jalankan `./scripts/dev-docker.sh` untuk menjalankan build terakhir.

Jika sudah berjalan sesuai keinginan, jalankan perintah `make publish` untuk mempublish base image sehingga dapat digunakan di manapun.