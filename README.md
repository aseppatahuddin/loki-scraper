## Loki Data Migration

Untuk dapat menggunakan, lakukan perubahan pada code. Commit changes. Kemudian execute `make build`.

Image baru akan kebuat dan di push under user prakasa1904. Ubah sesuai dengan scope registry kamu.

Selanjutnya lihat file `scripts/dev-docker.sh`. Jalankan `./scripts/dev-docker.sh` untuk menjalankan build terakhir.

Jika sudah berjalan sesuai keinginan, jalankan perintah `make publish` untuk mempublish base image sehingga dapat digunakan di manapun.